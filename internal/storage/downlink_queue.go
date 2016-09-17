package storage

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

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
		return fmt.Errorf("enqueue downlink queue item error: %s", err)
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
		return qi, fmt.Errorf("get downlink queue item error: %s", err)
	}
	return qi, nil
}

// GetPendingDownlinkQueueItem returns an item from the downlink queue that
// is pending.
func GetPendingDownlinkQueueItem(db *sqlx.DB, devEUI lorawan.EUI64) (DownlinkQueueItem, error) {
	var qi DownlinkQueueItem
	err := db.Get(&qi, "select * from downlink_queue where dev_eui = $1 and pending = $2", devEUI[:], true)
	if err != nil {
		return qi, fmt.Errorf("get pending downlink queue item error: %s", err)
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
		return fmt.Errorf("update downlink queue item error: %s", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return fmt.Errorf("downlink queue item id %d does not exist", item.ID)
	}
	log.WithField("id", item.ID).Info("downlink queue item updated")
	return nil
}

// DeleteDownlinkQueueItem deletes an item from the downlink queue.
func DeleteDownlinkQueueItem(db *sqlx.DB, id int64) error {
	res, err := db.Exec("delete from downlink_queue where id = $1", id)
	if err != nil {
		return fmt.Errorf("delete downlink queue item error: %s", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return fmt.Errorf("downlink queue item id %d does not exist", id)
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
		return nil, fmt.Errorf("get downlink queue items error: %s", err)
	}
	return items, nil
}
