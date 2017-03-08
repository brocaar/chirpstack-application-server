package storage

import (
	"database/sql"

	log "github.com/Sirupsen/logrus"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
)

// DownlinkQueueItem represents an item in the downlink queue.
type DownlinkQueueItem struct {
	ID        int64         `db:"id"`
	Reference string        `db:"reference"`
	DevEUI    lorawan.EUI64 `db:"dev_eui"`
	Confirmed bool          `db:"confirmed"`
	Pending   bool          `db:"pending"`
	FPort     uint8         `db:"fport"`
	Data      []byte        `db:"data"`
}

// CreateDownlinkQueueItem adds an item to the downlink queue.
func CreateDownlinkQueueItem(db *sqlx.DB, item *DownlinkQueueItem) error {
	err := db.Get(&item.ID, `
		insert into downlink_queue (
			dev_eui,
			reference,
			confirmed,
			pending,
			fport,
			data
		) values ($1, $2, $3, $4, $5, $6) returning id`,
		item.DevEUI[:],
		item.Reference,
		item.Confirmed,
		item.Pending,
		item.FPort,
		item.Data,
	)
	if err != nil {
		switch err := err.(type) {
		case *pq.Error:
			switch err.Code.Name() {
			case "foreign_key_violation":
				return ErrDoesNotExist
			default:
				return errors.Wrap(err, "insert error")
			}
		default:
			return errors.Wrap(err, "insert error")
		}
	}
	log.WithFields(log.Fields{
		"dev_eui": item.DevEUI,
		"id":      item.ID,
	}).Info("downlink queue item enqueued")
	return nil
}

// GetDownlinkQueueItem gets an item from the downlink queue.
func GetDownlinkQueueItem(db *sqlx.DB, id int64) (DownlinkQueueItem, error) {
	var qi DownlinkQueueItem
	err := db.Get(&qi, "select * from downlink_queue where id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return qi, ErrDoesNotExist
		}
		return qi, errors.Wrap(err, "select error")
	}
	return qi, nil
}

// GetDownlinkQueueSize returns the size of the downlink queue.
func GetDownlinkQueueSize(db *sqlx.DB, devEUI lorawan.EUI64) (int, error) {
	var count int
	err := db.Get(&count, "select count(*) from downlink_queue where dev_eui = $1", devEUI[:])
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetPendingDownlinkQueueItem returns an item from the downlink queue that
// is pending.
func GetPendingDownlinkQueueItem(db *sqlx.DB, devEUI lorawan.EUI64) (DownlinkQueueItem, error) {
	var qi DownlinkQueueItem
	err := db.Get(&qi, "select * from downlink_queue where dev_eui = $1 and pending = $2", devEUI[:], true)
	if err != nil {
		return qi, errors.Wrap(err, "select error")
	}
	return qi, nil
}

// UpdateDownlinkQueueItem updates and item in the downlink queue.
func UpdateDownlinkQueueItem(db *sqlx.DB, item DownlinkQueueItem) error {
	res, err := db.Exec(`
		update downlink_queue
		set
			dev_eui = $1,
			reference = $2,
			confirmed = $3,
			pending = $4,
			fport = $5,
			data = $6
		where id = $7`,
		item.DevEUI[:],
		item.Reference,
		item.Confirmed,
		item.Pending,
		item.FPort,
		item.Data,
		item.ID,
	)
	if err != nil {
		return errors.Wrap(err, "update error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}
	log.WithField("id", item.ID).Info("downlink queue item updated")
	return nil
}

// DeleteDownlinkQueueItem deletes an item from the downlink queue.
func DeleteDownlinkQueueItem(db *sqlx.DB, id int64) error {
	res, err := db.Exec("delete from downlink_queue where id = $1", id)
	if err != nil {
		return errors.Wrap(err, "delete error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}
	log.WithField("id", id).Info("downlink queue item deleted")
	return nil
}

// GetDownlinkQueueItems returns a list of downlink queue items for the
// given DevEUI.
func GetDownlinkQueueItems(db *sqlx.DB, devEUI lorawan.EUI64) ([]DownlinkQueueItem, error) {
	var items []DownlinkQueueItem
	err := db.Select(&items, "select * from downlink_queue where dev_eui = $1 order by id", devEUI[:])
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return items, nil
}

// DeleteDownlinkQueueItemsForDevEUI deletes all queue items for the given
// DevEUI.
func DeleteDownlinkQueueItemsForDevEUI(db *sqlx.DB, devEUI lorawan.EUI64) error {
	_, err := db.Exec("delete from downlink_queue where dev_eui = $1", devEUI[:])
	if err != nil {
		return errors.Wrap(err, "delete error")
	}
	return nil
}

// GetNextDownlinkQueueItem returns the next item from the queue, respecting
// the given maxPayloadSize. If an item exceeds this size, it is discarded and
// the next item is retrieved from the queue.
// When the queue is empty, nil is returned.
func GetNextDownlinkQueueItem(db *sqlx.DB, devEUI lorawan.EUI64, maxPayloadSize int) (*DownlinkQueueItem, error) {
	for {
		var qi DownlinkQueueItem
		err := db.Get(&qi, "select * from downlink_queue where dev_eui = $1 order by id limit 1", devEUI[:])
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, errors.Wrap(err, "select error")
		}

		if len(qi.Data) > maxPayloadSize {
			log.WithFields(log.Fields{
				"reference":        qi.Reference,
				"dev_eui":          qi.DevEUI,
				"max_payload_size": maxPayloadSize,
				"payload_size":     len(qi.Data),
			}).Warning("queue item discarded as it exceeds max payload size")

			if err := DeleteDownlinkQueueItem(db, qi.ID); err != nil {
				return nil, errors.Wrap(err, "delete downlink queue item error")
			}
			continue
		}

		return &qi, nil
	}
}
