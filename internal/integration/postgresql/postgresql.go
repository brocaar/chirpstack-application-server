package postgresql

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mmcloughlin/geohash"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/integration"
)

type Config struct {
	DSN string `json:"dsn"`
}

type Integration struct {
	db *sqlx.DB
}

// New creates a new PostgreSQL integration.
func New(conf Config) (*Integration, error) {
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

// SendDataUp stores the uplink data into the device_up table.
func (i *Integration) SendDataUp(pl integration.DataUpPayload) error {
	// use an UUID here so that we can later refactor this for correlation
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

	// get the rxTime either using the system-time or using one of the
	// gateway provided timestamps.
	rxTime := time.Now()
	for i := range pl.RXInfo {
		if pl.RXInfo[i].Time != nil {
			rxTime = *pl.RXInfo[i].Time
			break
		}
	}

	objectB, err := json.Marshal(pl.Object)
	if err != nil {
		return errors.Wrap(err, "marshal data error")
	}

	rxInfoB, err := json.Marshal(pl.RXInfo)
	if err != nil {
		return errors.Wrap(err, "marshal rx_info error")
	}

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
			object
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		id,
		rxTime,
		pl.DevEUI,
		pl.DeviceName,
		pl.ApplicationID,
		pl.ApplicationName,
		pl.TXInfo.Frequency,
		pl.TXInfo.DR,
		pl.ADR,
		pl.FCnt,
		pl.FPort,
		pl.Data,
		json.RawMessage(rxInfoB),
		json.RawMessage(objectB),
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}

	log.WithFields(log.Fields{
		"event":   "up",
		"dev_eui": pl.DevEUI,
	}).Info("integration/postgresql: event stored")

	return nil
}

// SendStatusNotification stores the device-status in the device_status table.
func (i *Integration) SendStatusNotification(pl integration.StatusNotification) error {
	// use an UUID here so that we can later refactor this for correlation
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

	// TODO: refactor to use gateway provided time once RXInfo is available
	// and timestamp is provided by gateway.
	rxTime := time.Now()

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
			battery_level
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		id,
		rxTime,
		pl.DevEUI,
		pl.DeviceName,
		pl.ApplicationID,
		pl.ApplicationName,
		pl.Margin,
		pl.ExternalPowerSource,
		pl.BatteryLevelUnavailable,
		pl.BatteryLevel,
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}

	log.WithFields(log.Fields{
		"event":   "status",
		"dev_eui": pl.DevEUI,
	}).Info("integration/postgresql: event stored")

	return nil
}

// SendJoinNotification stores the join in the device_join table.
func (i *Integration) SendJoinNotification(pl integration.JoinNotification) error {
	// use an UUID here so that we can later refactor this for correlation
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

	// TODO: refactor to use gateway provided time once RXInfo is available
	// and timestamp is provided by gateway.
	rxTime := time.Now()

	_, err = i.db.Exec(`
		insert into device_join (
			id,
			received_at,
			dev_eui,
			device_name,
			application_id,
			application_name,
			dev_addr
		) values ($1, $2, $3, $4, $5, $6, $7)`,
		id,
		rxTime,
		pl.DevEUI,
		pl.DeviceName,
		pl.ApplicationID,
		pl.ApplicationName,
		pl.DevAddr[:],
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}

	log.WithFields(log.Fields{
		"event":   "join",
		"dev_eui": pl.DevEUI,
	}).Info("integration/postgresql: event stored")

	return nil
}

// SendACKNotification stores the ACK in the device_ack table.
func (i *Integration) SendACKNotification(pl integration.ACKNotification) error {
	// use an UUID here so that we can later refactor this for correlation
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

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
			f_cnt
		) values ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id,
		rxTime,
		pl.DevEUI,
		pl.DeviceName,
		pl.ApplicationID,
		pl.ApplicationName,
		pl.Acknowledged,
		pl.FCnt,
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}

	log.WithFields(log.Fields{
		"event":   "ack",
		"dev_eui": pl.DevEUI,
	}).Info("integration/postgresql: event stored")

	return nil
}

// SendErrorNotification stores the error in the device_error table.
func (i *Integration) SendErrorNotification(pl integration.ErrorNotification) error {
	// use an UUID here so that we can later refactor this for correlation
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

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
			f_cnt
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id,
		rxTime,
		pl.DevEUI,
		pl.DeviceName,
		pl.ApplicationID,
		pl.ApplicationName,
		pl.Type,
		pl.Error,
		pl.FCnt,
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}

	log.WithFields(log.Fields{
		"event":   "error",
		"dev_eui": pl.DevEUI,
	}).Info("integration/postgresql: event stored")

	return nil
}

// SendLocationNotification stores the location in the device_location table.
func (i *Integration) SendLocationNotification(pl integration.LocationNotification) error {
	// use an UUID here so that we can later refactor this for correlation
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

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
			accuracy
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		id,
		rxTime,
		pl.DevEUI,
		pl.DeviceName,
		pl.ApplicationID,
		pl.ApplicationName,
		pl.Location.Altitude,
		pl.Location.Latitude,
		pl.Location.Longitude,
		geohash.Encode(pl.Location.Latitude, pl.Location.Longitude),
		0,
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}

	log.WithFields(log.Fields{
		"event":   "location",
		"dev_eui": pl.DevEUI,
	}).Info("integration/postgresql: event stored")

	return nil
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}
