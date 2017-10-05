package storage

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
)

// DeviceQueueItem represents an item in the device queue (downlink).
type DeviceQueueItem struct {
	ID        int64         `db:"id"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
	Reference string        `db:"reference"`
	DevEUI    lorawan.EUI64 `db:"dev_eui"`
	Confirmed bool          `db:"confirmed"`
	Pending   bool          `db:"pending"`
	FPort     uint8         `db:"fport"`
	Data      []byte        `db:"data"`
}

// CreateDeviceQueueItem adds the given item to the device queue.
func CreateDeviceQueueItem(db *sqlx.DB, item *DeviceQueueItem) error {
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	err := db.Get(&item.ID, `
		insert into device_queue (
			created_at,
			updated_at,
			dev_eui,
			reference,
			confirmed,
			pending,
			fport,
			data
		) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id`,
		item.CreatedAt,
		item.UpdatedAt,
		item.DevEUI[:],
		item.Reference,
		item.Confirmed,
		item.Pending,
		item.FPort,
		item.Data,
	)
	if err != nil {
		return handlePSQLError(err, "insert error")
	}

	log.WithFields(log.Fields{
		"dev_eui": item.DevEUI,
		"id":      item.ID,
	}).Info("device-queue item created")

	return nil
}

// GetDeviceQueueItem returns the device-queue item matching the given id.
func GetDeviceQueueItem(db *sqlx.DB, id int64) (DeviceQueueItem, error) {
	var qi DeviceQueueItem
	err := db.Get(&qi, "select * from device_queue where id = $1", id)
	if err != nil {
		return qi, handlePSQLError(err, "select error")
	}
	return qi, nil
}

// GetDeviceQueueItemCount returns the number of items in the device-queue.
func GetDeviceQueueItemCount(db *sqlx.DB, devEUI lorawan.EUI64) (int, error) {
	var count int
	err := db.Get(&count, "select count(*) from device_queue where dev_eui = $1", devEUI[:])
	if err != nil {
		return count, handlePSQLError(err, "select error")
	}
	return count, nil
}

// GetPendingDeviceQueueItem returns an item from the device-queue that
// is pending.
func GetPendingDeviceQueueItem(db *sqlx.DB, devEUI lorawan.EUI64) (DeviceQueueItem, error) {
	var qi DeviceQueueItem
	err := db.Get(&qi, "select * from device_queue where dev_eui = $1 and pending = $2", devEUI[:], true)
	if err != nil {
		return qi, handlePSQLError(err, "select error")
	}
	return qi, nil
}

// UpdateDeviceQueueItem updates the given device-queue item.
func UpdateDeviceQueueItem(db *sqlx.DB, item *DeviceQueueItem) error {
	item.UpdatedAt = time.Now()

	res, err := db.Exec(`
		update device_queue
		set
			updated_at = $2,
			dev_eui = $3,
			reference = $4,
			confirmed = $5,
			pending = $6,
			fport = $7,
			data = $8
		where id = $1`,
		item.ID,
		item.UpdatedAt,
		item.DevEUI[:],
		item.Reference,
		item.Confirmed,
		item.Pending,
		item.FPort,
		item.Data,
	)
	if err != nil {
		return handlePSQLError(err, "update error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	log.WithField("id", item.ID).Info("device-queue item updated")

	return nil
}

// DeleteDeviceQueueItem deletes the device-queue item matching the given id.
func DeleteDeviceQueueItem(db *sqlx.DB, id int64) error {
	res, err := db.Exec("delete from device_queue where id = $1", id)
	if err != nil {
		return handlePSQLError(err, "delete error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}
	log.WithField("id", id).Info("device-queue item deleted")

	return nil
}

// GetDeviceQueueItems returns a list of device-queue items for the
// given DevEUI.
func GetDeviceQueueItems(db *sqlx.DB, devEUI lorawan.EUI64) ([]DeviceQueueItem, error) {
	var items []DeviceQueueItem
	err := db.Select(&items, "select * from device_queue where dev_eui = $1 order by id", devEUI[:])
	if err != nil {
		return nil, handlePSQLError(err, "select error")
	}
	return items, nil
}

// DeleteDeviceQueueItemsForDevEUI deletes all device-queue items for the given
// DevEUI.
func DeleteDeviceQueueItemsForDevEUI(db *sqlx.DB, devEUI lorawan.EUI64) error {
	_, err := db.Exec("delete from device_queue where dev_eui = $1", devEUI[:])
	if err != nil {
		return handlePSQLError(err, "delete error")
	}
	return nil
}

// GetNextDeviceQueueItem returns the next device-queue item from the queue, respecting
// the given maxPayloadSize. If an item exceeds this size, it is discarded and
// the next item is retrieved from the queue.
// When the queue is empty, nil is returned.
func GetNextDeviceQueueItem(db *sqlx.DB, devEUI lorawan.EUI64, maxPayloadSize int) (*DeviceQueueItem, error) {
	for {
		var qi DeviceQueueItem
		err := db.Get(&qi, "select * from device_queue where dev_eui = $1 order by id limit 1", devEUI[:])
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, handlePSQLError(err, "select error")
		}

		if len(qi.Data) > maxPayloadSize {
			log.WithFields(log.Fields{
				"reference":        qi.Reference,
				"dev_eui":          qi.DevEUI,
				"max_payload_size": maxPayloadSize,
				"payload_size":     len(qi.Data),
			}).Warning("queue item discarded as it exceeds max payload size")

			if err := DeleteDeviceQueueItem(db, qi.ID); err != nil {
				return nil, errors.Wrap(err, "delete device-queue item error")
			}
			continue
		}

		return &qi, nil
	}
}
