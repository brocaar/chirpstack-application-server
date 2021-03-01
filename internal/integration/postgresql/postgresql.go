package postgresql

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/golang/protobuf/ptypes"
	"github.com/jmoiron/sqlx"

	"github.com/lib/pq/hstore"
	"github.com/mmcloughlin/geohash"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

// Migrations
//go:embed migrations/*
var migrations embed.FS

// Integration implements a PostgreSQL integration.
type Integration struct {
	db *sqlx.DB
}

// New creates a new PostgreSQL integration.
func New(conf config.IntegrationPostgreSQLConfig) (*Integration, error) {
	log.Info("integration/postgresql: connecting to PostgreSQL database")
	d, err := sqlx.Open("postgres", conf.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "integration/postgresql: PostgreSQL connection error")
	}
	for {
		if err := d.Ping(); err != nil {
			log.WithError(err).Warning("integration/postgresql: ping PostgreSQL database error, will retry in 2s")
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}

	d.SetMaxOpenConns(conf.MaxOpenConnections)
	d.SetMaxIdleConns(conf.MaxIdleConnections)

	if err := MigrateUp(d); err != nil {
		return nil, err
	}
	return &Integration{
		db: d,
	}, nil
}

// Close closes the integration.
func (i *Integration) Close() error {
	if err := i.db.Close(); err != nil {
		return errors.Wrap(err, "close database error")
	}
	return nil
}

// MigrateUp configure postgres-integration migration down
func MigrateUp(db *sqlx.DB) error {
	log.Info("integration/postgresql: applying PostgreSQL schema migrations")

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("integration/postgresql: migrate postgres driver error: %w", err)
	}

	src, err := httpfs.New(http.FS(migrations), "/migrations")
	if err != nil {
		return fmt.Errorf("new httpfs error: %w", err)
	}

	m, err := migrate.NewWithInstance("httpfs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("integration/postgresql: new migrate instance error: %w", err)
	}

	oldVersion, _, _ := m.Version()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("integration/postgresql: migrate up error: %w", err)
	}

	newVersion, _, _ := m.Version()

	if oldVersion != newVersion {
		log.WithFields(log.Fields{
			"from_version": oldVersion,
			"to_version":   newVersion,
		}).Info("integration/postgresql: applied database migrations")
	}

	return nil
}

// MigrateDown configure postgres-integration migration down
func MigrateDown(db *sqlx.DB) error {
	log.Info("integration/postgresql: reverting PostgreSQL schema migrations")

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("integration/postgresql: migrate postgres driver error: %w", err)
	}

	src, err := httpfs.New(http.FS(migrations), "/migrations")
	if err != nil {
		return fmt.Errorf("new httpfs error: %w", err)
	}

	m, err := migrate.NewWithInstance("httpfs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("integration/postgresql: new migrate instance error: %w", err)
	}

	oldVersion, _, _ := m.Version()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("integration/postgresql: migrate down error: %w", err)
	}

	newVersion, _, _ := m.Version()

	if oldVersion != newVersion {
		log.WithFields(log.Fields{
			"from_version": oldVersion,
			"to_version":   newVersion,
		}).Info("integration/postgresql: applied database migrations")
	}

	return nil
}

// HandleUplinkEvent writes an UplinkEvent into the database.
func (i *Integration) HandleUplinkEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.UplinkEvent) error {
	// use an UUID here so that we can later refactor this for correlation
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

	// get the rxTime either using the system-time or using one of the
	// gateway provided timestamps.
	rxTime := time.Now()
	for _, rxInfo := range pl.RxInfo {
		if rxInfo.Time != nil {
			ts, err := ptypes.Timestamp(rxInfo.Time)
			if err != nil {
				return errors.Wrap(err, "protobuf timestamp error")
			}
			rxTime = ts
		}
	}

	rxInfoJSON, err := getRXInfoJSON(pl.RxInfo)
	if err != nil {
		return errors.Wrap(err, "get rxInfo json error")
	}

	objectJSON := json.RawMessage("null")
	if pl.ObjectJson != "" {
		objectJSON = json.RawMessage(pl.ObjectJson)
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	_, err = i.db.Exec(`
		insert into device_up (
			id,
			received_at,
			dev_eui,
			device_name,
			application_id,
			application_name,
			frequency,
			dr,
			adr,
			f_cnt,
			f_port,
			data,
			rx_info,
			object,
			tags
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		id,
		rxTime,
		devEUI,
		pl.DeviceName,
		pl.ApplicationId,
		pl.ApplicationName,
		pl.GetTxInfo().GetFrequency(),
		pl.Dr,
		pl.Adr,
		pl.FCnt,
		pl.FPort,
		pl.Data,
		rxInfoJSON,
		objectJSON,
		tagsToHstore(pl.Tags),
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}

	log.WithFields(log.Fields{
		"event":   "up",
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/postgresql: event stored")

	return nil
}

// HandleStatusEvent writes a StatusEvent into the database.
func (i *Integration) HandleStatusEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.StatusEvent) error {
	// use an UUID here so that we can later refactor this for correlation
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

	// TODO: refactor to use gateway provided time once RXInfo is available
	// and timestamp is provided by gateway.
	rxTime := time.Now()

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	_, err = i.db.Exec(`
		insert into device_status (
			id,
			received_at,
			dev_eui,
			device_name,
			application_id,
			application_name,
			margin,
			external_power_source,
			battery_level_unavailable,
			battery_level,
			tags
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		id,
		rxTime,
		devEUI,
		pl.DeviceName,
		pl.ApplicationId,
		pl.ApplicationName,
		pl.Margin,
		pl.ExternalPowerSource,
		pl.BatteryLevelUnavailable,
		pl.BatteryLevel,
		tagsToHstore(pl.Tags),
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}

	log.WithFields(log.Fields{
		"event":   "status",
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/postgresql: event stored")

	return nil
}

// HandleJoinEvent writes a JoinEvent into the database.
func (i *Integration) HandleJoinEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.JoinEvent) error {
	// use an UUID here so that we can later refactor this for correlation
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

	// TODO: refactor to use gateway provided time once RXInfo is available
	// and timestamp is provided by gateway.
	rxTime := time.Now()

	var devEUI lorawan.EUI64
	var devAddr lorawan.DevAddr
	copy(devEUI[:], pl.DevEui)
	copy(devAddr[:], pl.DevAddr)

	_, err = i.db.Exec(`
		insert into device_join (
			id,
			received_at,
			dev_eui,
			device_name,
			application_id,
			application_name,
			dev_addr,
			tags
		) values ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id,
		rxTime,
		devEUI,
		pl.DeviceName,
		pl.ApplicationId,
		pl.ApplicationName,
		devAddr[:],
		tagsToHstore(pl.Tags),
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}

	log.WithFields(log.Fields{
		"event":   "join",
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/postgresql: event stored")

	return nil
}

// HandleAckEvent writes an AckEvent into the database.
func (i *Integration) HandleAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.AckEvent) error {
	// use an UUID here so that we can later refactor this for correlation
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	// TODO: refactor to use gateway provided time once RXInfo is available
	// and timestamp is provided by gateway.
	rxTime := time.Now()

	_, err = i.db.Exec(`
		insert into device_ack (
			id,
			received_at,
			dev_eui,
			device_name,
			application_id,
			application_name,
			acknowledged,
			f_cnt,
			tags
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id,
		rxTime,
		devEUI,
		pl.DeviceName,
		pl.ApplicationId,
		pl.ApplicationName,
		pl.Acknowledged,
		pl.FCnt,
		tagsToHstore(pl.Tags),
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}

	log.WithFields(log.Fields{
		"event":   "ack",
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/postgresql: event stored")

	return nil
}

// HandleErrorEvent writes an ErrorEvent into the database.
func (i *Integration) HandleErrorEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.ErrorEvent) error {
	// use an UUID here so that we can later refactor this for correlation
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	// TODO: refactor to use gateway provided time once RXInfo is available
	// and timestamp is provided by gateway.
	rxTime := time.Now()

	_, err = i.db.Exec(`
		insert into device_error (
			id,
			received_at,
			dev_eui,
			device_name,
			application_id,
			application_name,
			type,
			error,
			f_cnt,
			tags
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		id,
		rxTime,
		devEUI,
		pl.DeviceName,
		pl.ApplicationId,
		pl.ApplicationName,
		pl.Type.String(),
		pl.Error,
		pl.FCnt,
		tagsToHstore(pl.Tags),
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}

	log.WithFields(log.Fields{
		"event":   "error",
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/postgresql: event stored")

	return nil
}

// HandleLocationEvent writes a LocationEvent into the database.
func (i *Integration) HandleLocationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.LocationEvent) error {
	// use an UUID here so that we can later refactor this for correlation
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	// TODO: refactor to use gateway provided time once RXInfo is available
	// and timestamp is provided by gateway.
	rxTime := time.Now()

	_, err = i.db.Exec(`
		insert into device_location (
			id,
			received_at,
			dev_eui,
			device_name,
			application_id,
			application_name,
			altitude,
			latitude,
			longitude,
			geohash,
			accuracy,
			tags
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		id,
		rxTime,
		devEUI,
		pl.DeviceName,
		pl.ApplicationId,
		pl.ApplicationName,
		pl.GetLocation().GetAltitude(),
		pl.GetLocation().GetLatitude(),
		pl.GetLocation().GetLongitude(),
		geohash.Encode(pl.GetLocation().GetLatitude(), pl.GetLocation().GetLongitude()),
		0,
		tagsToHstore(pl.Tags),
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}

	log.WithFields(log.Fields{
		"event":   "location",
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/postgresql: event stored")

	return nil
}

// HandleTxAckEvent is not implemented.
// TODO: implement this + schema migrations for the PostgreSQL database!
func (i *Integration) HandleTxAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.TxAckEvent) error {
	return nil
}

// HandleIntegrationEvent is not implemented.
// TODO: implement this + schema migrations for the PostgreSQL database!
func (i *Integration) HandleIntegrationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.IntegrationEvent) error {
	return nil
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return nil
}

func getRXInfoJSON(rxInfo []*gw.UplinkRXInfo) (json.RawMessage, error) {
	var out []models.RXInfo
	var gatewayIDs []lorawan.EUI64

	for i := range rxInfo {
		rx := models.RXInfo{
			RSSI:    int(rxInfo[i].Rssi),
			LoRaSNR: rxInfo[i].LoraSnr,
		}

		copy(rx.GatewayID[:], rxInfo[i].GatewayId)
		copy(rx.UplinkID[:], rxInfo[i].UplinkId)

		if rxInfo[i].Time != nil {
			ts, err := ptypes.Timestamp(rxInfo[i].Time)
			if err != nil {
				return nil, errors.Wrap(err, "proto timestamp error")
			}

			rx.Time = &ts
		}

		if rxInfo[i].Location != nil {
			rx.Location = &models.Location{
				Latitude:  rxInfo[i].Location.Latitude,
				Longitude: rxInfo[i].Location.Longitude,
				Altitude:  rxInfo[i].Location.Altitude,
			}
		}

		gatewayIDs = append(gatewayIDs, rx.GatewayID)
		out = append(out, rx)
	}

	gws, err := storage.GetGatewaysForMACs(context.Background(), storage.DB(), gatewayIDs)
	if err != nil {
		return nil, errors.Wrap(err, "get gateways for ids error")
	}
	for i := range out {
		if gw, ok := gws[out[i].GatewayID]; ok {
			out[i].Name = gw.Name
		}
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, errors.Wrap(err, "marshal json error")
	}

	return json.RawMessage(b), nil
}

func tagsToHstore(tags map[string]string) hstore.Hstore {
	out := hstore.Hstore{
		Map: make(map[string]sql.NullString),
	}

	for k, v := range tags {
		out.Map[k] = sql.NullString{String: v, Valid: true}
	}

	return out
}
