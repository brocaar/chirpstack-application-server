package code

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

// MigrateGatewayStats imports the gateway stats from the network-server.
func MigrateGatewayStats(db sqlx.Ext) error {
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
		if err := migrateGatewayStatsForGatewayID(db, id); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"gateway_id": id,
			}).Error("migrate gateway stats error")
		}
	}

	return nil
}

func migrateGatewayStatsForGatewayID(db sqlx.Ext, gatewayID lorawan.EUI64) error {
	gw, err := storage.GetGateway(context.Background(), db, gatewayID, true)
	if err != nil {
		return errors.Wrap(err, "get gateway error")
	}

	n, err := storage.GetNetworkServer(context.Background(), db, gw.NetworkServerID)
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

	if err := storage.UpdateGateway(context.Background(), db, &gw); err != nil {
		return errors.Wrap(err, "update gateway error")
	}

	if err := migrateGatewayStatsForGatewayIDInterval(nsClient, gatewayID, ns.AggregationInterval_MINUTE, time.Now().Add(-config.C.Metrics.Redis.MinuteAggregationTTL), time.Now()); err != nil {
		return err
	}

	if err := migrateGatewayStatsForGatewayIDInterval(nsClient, gatewayID, ns.AggregationInterval_HOUR, time.Now().Add(-config.C.Metrics.Redis.HourAggregationTTL), time.Now()); err != nil {
		return err
	}

	if err := migrateGatewayStatsForGatewayIDInterval(nsClient, gatewayID, ns.AggregationInterval_DAY, time.Now().Add(-config.C.Metrics.Redis.DayAggregationTTL), time.Now()); err != nil {
		return err
	}

	if err := migrateGatewayStatsForGatewayIDInterval(nsClient, gatewayID, ns.AggregationInterval_MONTH, time.Now().Add(-config.C.Metrics.Redis.MonthAggregationTTL), time.Now()); err != nil {
		return err
	}

	return nil
}

func migrateGatewayStatsForGatewayIDInterval(nsClient ns.NetworkServerServiceClient, gatewayID lorawan.EUI64, interval ns.AggregationInterval, start, end time.Time) error {
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

		err = storage.SaveMetricsForInterval(context.Background(), storage.AggregationInterval(interval.String()), storage.GetRedisKey("gw:%s", gatewayID), storage.MetricsRecord{
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
