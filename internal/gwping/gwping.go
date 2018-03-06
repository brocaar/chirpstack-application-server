package gwping

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/as"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

const (
	micLookupExpire = time.Second * 10
	micLookupTempl  = "lora:as:gwping:%s"
)

// SendPingLoop is a never returning function sending the gateway pings.
func SendPingLoop() {
	for {
		if err := sendGatewayPing(); err != nil {
			log.Errorf("send gateway ping error: %s", err)
		}
		time.Sleep(time.Second)
	}
}

// HandleReceivedPing handles a ping received by one or multiple gateways.
func HandleReceivedPing(req *as.HandleProprietaryUplinkRequest) error {
	var mic lorawan.MIC
	copy(mic[:], req.Mic)

	id, err := getPingLookup(mic)
	if err != nil {
		return errors.Wrap(err, "get ping lookup error")
	}

	if err = deletePingLookup(mic); err != nil {
		log.Errorf("delete ping lookup error: %s", err)
	}

	ping, err := storage.GetGatewayPing(config.C.PostgreSQL.DB, id)
	if err != nil {
		return errors.Wrap(err, "get gateway ping error")
	}

	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		for _, rx := range req.RxInfo {
			var receivedAt *time.Time
			var mac lorawan.EUI64
			copy(mac[:], rx.Mac)

			// ignore pings received by the sending gateway
			if ping.GatewayMAC == mac {
				continue
			}

			if rx.Time != "" {
				t, err := time.Parse(time.RFC3339Nano, rx.Time)
				if err != nil {
					return errors.Wrap(err, "parse time error")
				}
				receivedAt = &t
			}

			err := storage.CreateGatewayPingRX(tx, &storage.GatewayPingRX{
				PingID:     id,
				GatewayMAC: mac,
				ReceivedAt: receivedAt,
				RSSI:       int(rx.Rssi),
				LoRaSNR:    rx.LoRaSNR,
				Location: storage.GPSPoint{
					Latitude:  rx.Latitude,
					Longitude: rx.Longitude,
				},
				Altitude: rx.Altitude,
			})
			if err != nil {
				return errors.Wrap(err, "create gateway ping rx error")
			}
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "transaction error")
	}

	return nil
}

// sendGatewayPing selects the next gateway to ping, creates the "ping"
// frame and sends this frame to the network-server for transmission.
func sendGatewayPing() error {
	return storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		gw, err := getGatewayForPing(tx)
		if err != nil {
			return errors.Wrap(err, "get gateway for ping error")
		}
		if gw == nil {
			return nil
		}

		ping := storage.GatewayPing{
			GatewayMAC: gw.MAC,
			Frequency:  config.C.ApplicationServer.GatewayDiscovery.Frequency,
			DR:         config.C.ApplicationServer.GatewayDiscovery.DR,
		}
		err = storage.CreateGatewayPing(tx, &ping)
		if err != nil {
			return errors.Wrap(err, "create gateway ping error")
		}

		var mic lorawan.MIC
		if _, err = rand.Read(mic[:]); err != nil {
			return errors.Wrap(err, "read random bytes error")
		}

		err = CreatePingLookup(mic, ping.ID)
		if err != nil {
			return errors.Wrap(err, "store mic lookup error")
		}

		err = sendPing(mic, ping)
		if err != nil {
			return errors.Wrap(err, "send ping error")
		}

		gw.LastPingID = &ping.ID
		gw.LastPingSentAt = &ping.CreatedAt

		err = storage.UpdateGateway(tx, gw)
		if err != nil {
			return errors.Wrap(err, "update gateway error")
		}

		return nil
	})
}

// getGatewayForPing returns the next gateway for sending a ping. If no gateway
// matches the filter criteria, nil is returned.
func getGatewayForPing(tx sqlx.Ext) (*storage.Gateway, error) {
	var gw storage.Gateway

	err := sqlx.Get(tx, &gw, `
		select *
		from gateway
		where
			ping = true
			and (last_ping_sent_at is null or last_ping_sent_at <= $1)
		order by last_ping_sent_at
		limit 1
		for update`,
		time.Now().Add(-config.C.ApplicationServer.GatewayDiscovery.Interval),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "select error")
	}

	return &gw, nil
}

func sendPing(mic lorawan.MIC, ping storage.GatewayPing) error {
	gw, err := storage.GetGateway(config.C.PostgreSQL.DB, ping.GatewayMAC, false)
	if err != nil {
		return errors.Wrap(err, "get gateway error")
	}

	n, err := storage.GetNetworkServer(config.C.PostgreSQL.DB, gw.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.SendProprietaryPayload(context.Background(), &ns.SendProprietaryPayloadRequest{
		Mic:         mic[:],
		GatewayMACs: [][]byte{ping.GatewayMAC[:]},
		IPol:        false,
		Frequency:   uint32(ping.Frequency),
		Dr:          uint32(ping.DR),
	})
	if err != nil {
		return errors.Wrap(err, "send proprietary payload error")
	}

	log.WithFields(log.Fields{
		"gateway_mac": ping.GatewayMAC,
		"id":          ping.ID,
	}).Info("gateway ping sent to network-server")

	return nil
}

// CreatePingLookup creates an automatically expiring MIC to ping id lookup.
func CreatePingLookup(mic lorawan.MIC, id int64) error {
	c := config.C.Redis.Pool.Get()
	defer c.Close()

	_, err := redis.String(c.Do("PSETEX", fmt.Sprintf(micLookupTempl, mic), int64(micLookupExpire)/int64(time.Millisecond), id))
	if err != nil {
		return errors.Wrap(err, "set mic lookup error")
	}
	return nil
}

func getPingLookup(mic lorawan.MIC) (int64, error) {
	c := config.C.Redis.Pool.Get()
	defer c.Close()

	id, err := redis.Int64(c.Do("GET", fmt.Sprintf(micLookupTempl, mic)))
	if err != nil {
		return 0, errors.Wrap(err, "get ping lookup error")
	}

	return id, nil
}

func deletePingLookup(mic lorawan.MIC) error {
	c := config.C.Redis.Pool.Get()
	defer c.Close()

	_, err := redis.Int(c.Do("DEL", fmt.Sprintf(micLookupTempl, mic)))
	if err != nil {
		return errors.Wrap(err, "delete ping lookup error")
	}

	return nil
}
