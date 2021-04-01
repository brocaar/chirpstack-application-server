package code

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
	uuid "github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/lib/pq/hstore"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/lorawan"
)

type Gateway struct {
	MAC              lorawan.EUI64 `db:"mac"`
	CreatedAt        time.Time     `db:"created_at"`
	UpdatedAt        time.Time     `db:"updated_at"`
	FirstSeenAt      *time.Time    `db:"first_seen_at"`
	LastSeenAt       *time.Time    `db:"last_seen_at"`
	Name             string        `db:"name"`
	Description      string        `db:"description"`
	OrganizationID   int64         `db:"organization_id"`
	Ping             bool          `db:"ping"`
	LastPingID       *int64        `db:"last_ping_id"`
	LastPingSentAt   *time.Time    `db:"last_ping_sent_at"`
	NetworkServerID  int64         `db:"network_server_id"`
	GatewayProfileID *uuid.UUID    `db:"gateway_profile_id"`
	ServiceProfileID *uuid.UUID    `db:"service_profile_id"`
	Latitude         float64       `db:"latitude"`
	Longitude        float64       `db:"longitude"`
	Altitude         float64       `db:"altitude"`
	Tags             hstore.Hstore `db:"tags"`
	Metadata         hstore.Hstore `db:"metadata"`
}

type NetworkServer struct {
	ID                          int64     `db:"id"`
	CreatedAt                   time.Time `db:"created_at"`
	UpdatedAt                   time.Time `db:"updated_at"`
	Name                        string    `db:"name"`
	Server                      string    `db:"server"`
	CACert                      string    `db:"ca_cert"`
	TLSCert                     string    `db:"tls_cert"`
	TLSKey                      string    `db:"tls_key"`
	RoutingProfileCACert        string    `db:"routing_profile_ca_cert"`
	RoutingProfileTLSCert       string    `db:"routing_profile_tls_cert"`
	RoutingProfileTLSKey        string    `db:"routing_profile_tls_key"`
	GatewayDiscoveryEnabled     bool      `db:"gateway_discovery_enabled"`
	GatewayDiscoveryInterval    int       `db:"gateway_discovery_interval"`
	GatewayDiscoveryTXFrequency int       `db:"gateway_discovery_tx_frequency"`
	GatewayDiscoveryDR          int       `db:"gateway_discovery_dr"`
}

// MigrateGatewayStats imports the gateway stats from the network-server.
func MigrateGatewayStats(redisClient redis.UniversalClient, db sqlx.Ext, conf config.Config) error {
	var ids []lorawan.EUI64
	err := sqlx.Select(db, &ids, `
		select
			mac
		from
			gateway
	`)
	if err != nil {
		return errors.Wrap(err, "select gateway ids error")
	}

	for _, id := range ids {
		if err := migrateGatewayStatsForGatewayID(redisClient, conf, db, id); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"gateway_id": id,
			}).Error("migrate gateway stats error")
		}
	}

	return nil
}

func migrateGatewayStatsForGatewayID(redisClient redis.UniversalClient, conf config.Config, db sqlx.Ext, gatewayID lorawan.EUI64) error {
	var gw Gateway
	err := sqlx.Get(db, &gw, "select * from gateway where mac = $1"+" for update", gatewayID)
	if err != nil {
		return errors.Wrap(err, "get gateway error")
	}

	var n NetworkServer
	err = sqlx.Get(db, &n, "select * from network_server where id = $1", gw.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	nsGw, err := nsClient.GetGateway(context.Background(), &ns.GetGatewayRequest{
		Id: gatewayID[:],
	})
	if err != nil {
		return errors.Wrap(err, "get gateway from network-server error")
	}

	if nsGw.Gateway != nil && nsGw.Gateway.Location != nil {
		gw.Latitude = nsGw.Gateway.Location.Latitude
		gw.Longitude = nsGw.Gateway.Location.Longitude
		gw.Altitude = nsGw.Gateway.Location.Altitude
	}

	if err := UpdateGateway(context.Background(), db, gw); err != nil {
		return errors.Wrap(err, "update gateway error")
	}

	if err := migrateGatewayStatsForGatewayIDInterval(conf, redisClient, nsClient, gatewayID, ns.AggregationInterval_MINUTE, time.Now().Add(-config.C.Metrics.Redis.MinuteAggregationTTL), time.Now()); err != nil {
		return err
	}

	if err := migrateGatewayStatsForGatewayIDInterval(conf, redisClient, nsClient, gatewayID, ns.AggregationInterval_HOUR, time.Now().Add(-config.C.Metrics.Redis.HourAggregationTTL), time.Now()); err != nil {
		return err
	}

	if err := migrateGatewayStatsForGatewayIDInterval(conf, redisClient, nsClient, gatewayID, ns.AggregationInterval_DAY, time.Now().Add(-config.C.Metrics.Redis.DayAggregationTTL), time.Now()); err != nil {
		return err
	}

	if err := migrateGatewayStatsForGatewayIDInterval(conf, redisClient, nsClient, gatewayID, ns.AggregationInterval_MONTH, time.Now().Add(-config.C.Metrics.Redis.MonthAggregationTTL), time.Now()); err != nil {
		return err
	}

	return nil
}

func migrateGatewayStatsForGatewayIDInterval(conf config.Config, redisClient redis.UniversalClient, nsClient ns.NetworkServerServiceClient, gatewayID lorawan.EUI64, interval ns.AggregationInterval, start, end time.Time) error {
	startPB, err := ptypes.TimestampProto(start)
	if err != nil {
		return err
	}

	endPB, err := ptypes.TimestampProto(end)
	if err != nil {
		return err
	}

	metrics, err := nsClient.GetGatewayStats(context.Background(), &ns.GetGatewayStatsRequest{
		GatewayId:      gatewayID[:],
		Interval:       interval,
		StartTimestamp: startPB,
		EndTimestamp:   endPB,
	})
	if err != nil {
		return errors.Wrap(err, "get gateway stats from network-server error")
	}

	for _, m := range metrics.Result {
		ts, err := ptypes.Timestamp(m.Timestamp)
		if err != nil {
			return err
		}

		gwRedisKey := fmt.Sprintf("gw:%s", gatewayID)
		err = SaveMetricsForInterval(conf, redisClient, AggregationInterval(interval.String()), gwRedisKey, MetricsRecord{
			Time: ts,
			Metrics: map[string]float64{
				"rx_count":    float64(m.RxPacketsReceived),
				"rx_ok_count": float64(m.RxPacketsReceivedOk),
				"tx_count":    float64(m.TxPacketsReceived),
				"tx_ok_count": float64(m.TxPacketsEmitted),
			},
		})
		if err != nil {
			return errors.Wrap(err, "save metrics for interval error")
		}
	}

	return nil
}

// UpdateGateway updates the given Gateway.
func UpdateGateway(ctx context.Context, db sqlx.Execer, gw Gateway) error {
	gw.UpdatedAt = time.Now()

	res, err := db.Exec(`
		update gateway
			set updated_at = $2,
			name = $3,
			description = $4,
			organization_id = $5,
			ping = $6,
			last_ping_id = $7,
			last_ping_sent_at = $8,
			network_server_id = $9,
			gateway_profile_id = $10,
			first_seen_at = $11,
			last_seen_at = $12,
			latitude = $13,
			longitude = $14,
			altitude = $15,
			tags = $16,
			metadata = $17,
			service_profile_id = $18
		where
			mac = $1`,
		gw.MAC[:],
		gw.UpdatedAt,
		gw.Name,
		gw.Description,
		gw.OrganizationID,
		gw.Ping,
		gw.LastPingID,
		gw.LastPingSentAt,
		gw.NetworkServerID,
		gw.GatewayProfileID,
		gw.FirstSeenAt,
		gw.LastSeenAt,
		gw.Latitude,
		gw.Longitude,
		gw.Altitude,
		gw.Tags,
		gw.Metadata,
		gw.ServiceProfileID,
	)
	if err != nil {
		switch err := err.(type) {
		case *pq.Error:
			switch err.Code.Name() {
			case "unique_violation":
				return nil
			}
		}

		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return nil
	}
	return nil
}
