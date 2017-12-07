package storage

import (
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
)

// DeviceQueueMapping holds the data mapping a device-queue item to a
// reference.
type DeviceQueueMapping struct {
	ID        int64         `db:"id"`
	CreatedAt time.Time     `db:"created_at"`
	Reference string        `db:"reference"`
	DevEUI    lorawan.EUI64 `db:"dev_eui"`
	FCnt      uint32        `db:"f_cnt"`
}

// CreateDeviceQueueMapping creates the given device-queue mapping.
func CreateDeviceQueueMapping(db sqlx.Queryer, dqm *DeviceQueueMapping) error {
	dqm.CreatedAt = time.Now()

	err := sqlx.Get(db, &dqm.ID, `
		insert into device_queue_mapping (
			created_at,
			reference,
			dev_eui,
			f_cnt
		) values ($1, $2, $3, $4)
		returning id`,
		dqm.CreatedAt,
		dqm.Reference,
		dqm.DevEUI[:],
		dqm.FCnt,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	log.WithFields(log.Fields{
		"dev_eui":   dqm.DevEUI,
		"f_cnt":     dqm.FCnt,
		"reference": dqm.Reference,
		"id":        dqm.ID,
	}).Info("device-queue mapping created")

	return nil
}

// DeleteDeviceQueueMapping deletes the device-queue mapping matching the
// given ID.
func DeleteDeviceQueueMapping(db sqlx.Execer, id int64) error {
	res, err := db.Exec("delete from device_queue_mapping where id = $1", id)
	if err != nil {
		return handlePSQLError(Delete, err, "delete error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}
	log.WithField("id", id).Info("device-queue mapping deleted")
	return nil
}

// GetDeviceQueueMappingForDevEUIAndFCnt returns the device-queue mapping for
// the given DevEUI and FCnt.
func GetDeviceQueueMappingForDevEUIAndFCnt(db sqlx.Ext, devEUI lorawan.EUI64, fCnt uint32) (DeviceQueueMapping, error) {
	for {
		var dqm DeviceQueueMapping
		err := sqlx.Get(db, &dqm, `
			select
				*
			from
				device_queue_mapping
			where
				dev_eui = $1
			order by id`,
			devEUI[:],
		)
		if err != nil {
			return dqm, handlePSQLError(Select, err, "select error")
		}

		// frame-counters match
		if dqm.FCnt == fCnt {
			return dqm, nil
		}

		// Avoid that we are discarding FCnt+1 mappings.
		if fCnt < dqm.FCnt {
			return DeviceQueueMapping{}, ErrDoesNotExist
		}

		// Clean up old mappings.
		if err := DeleteDeviceQueueMapping(db, dqm.ID); err != nil {
			return DeviceQueueMapping{}, errors.Wrap(err, "delete device-queue mapping error")
		}
	}
}

// FlushDeviceQueueMappingForDevEUI flushes the device-queue mapping for the
// given DevEUI.
func FlushDeviceQueueMappingForDevEUI(db sqlx.Execer, devEUI lorawan.EUI64) error {
	_, err := db.Exec("delete from device_queue_mapping where dev_eui = $1", devEUI[:])
	if err != nil {
		return handlePSQLError(Delete, err, "delete error")
	}
	return nil
}
