package storage

import (
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lorawan"
)

// RemoteMulticastSetupState defines the state type.
type RemoteMulticastSetupState string

// Possible states
const (
	RemoteMulticastSetupSetup  RemoteMulticastSetupState = "SETUP"
	RemoteMulticastSetupDelete RemoteMulticastSetupState = "DELETE"
)

// RemoteMulticastSetup defines a remote multicast-setup record.
type RemoteMulticastSetup struct {
	DevEUI           lorawan.EUI64             `db:"dev_eui"`
	MulticastGroupID uuid.UUID                 `db:"multicast_group_id"`
	CreatedAt        time.Time                 `db:"created_at"`
	UpdatedAt        time.Time                 `db:"updated_at"`
	McGroupID        int                       `db:"mc_group_id"`
	McAddr           lorawan.DevAddr           `db:"mc_addr"`
	McKeyEncrypted   lorawan.AES128Key         `db:"mc_key_encrypted"`
	MinMcFCnt        uint32                    `db:"min_mc_f_cnt"`
	MaxMcFCnt        uint32                    `db:"max_mc_f_cnt"`
	State            RemoteMulticastSetupState `db:"state"`
	StateProvisioned bool                      `db:"state_provisioned"`
	RetryInterval    time.Duration             `db:"retry_interval"`
	RetryAfter       time.Time                 `db:"retry_after"`
	RetryCount       int                       `db:"retry_count"`
}

// CreateRemoteMulticastSetup creates the given multicast-setup.
func CreateRemoteMulticastSetup(db sqlx.Ext, dms *RemoteMulticastSetup) error {
	now := time.Now()
	dms.CreatedAt = now
	dms.UpdatedAt = now

	_, err := db.Exec(`
		insert into remote_multicast_setup (
			dev_eui,
			multicast_group_id,
			created_at,
			updated_at,
			mc_group_id,
			mc_addr,
			mc_key_encrypted,
			min_mc_f_cnt,
			max_mc_f_cnt,
			state,
			state_provisioned,
			retry_after,
			retry_count,
			retry_interval
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		dms.DevEUI[:],
		dms.MulticastGroupID,
		dms.CreatedAt,
		dms.UpdatedAt,
		dms.McGroupID,
		dms.McAddr[:],
		dms.McKeyEncrypted[:],
		dms.MinMcFCnt,
		dms.MaxMcFCnt,
		dms.State,
		dms.StateProvisioned,
		dms.RetryAfter,
		dms.RetryCount,
		dms.RetryInterval,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	log.WithFields(log.Fields{
		"dev_eui":            dms.DevEUI,
		"multicast_group_id": dms.MulticastGroupID,
	}).Info("remote multicast-setup created")
	return nil
}

// GetRemoteMulticastSetup returns the multicast-setup given a multicast-group ID and DevEUI.
func GetRemoteMulticastSetup(db sqlx.Queryer, devEUI lorawan.EUI64, multicastGroupID uuid.UUID, forUpdate bool) (RemoteMulticastSetup, error) {
	var fu string
	if forUpdate {
		fu = " for update"
	}

	var dmg RemoteMulticastSetup
	if err := sqlx.Get(db, &dmg, `
		select
			*
		from
			remote_multicast_setup
		where
			dev_eui = $1
			and multicast_group_id = $2`+fu,
		devEUI,
		multicastGroupID,
	); err != nil {
		return dmg, handlePSQLError(Select, err, "select error")
	}

	return dmg, nil
}

// GetRemoteMulticastSetupByGroupID returns the multicast-setup given a DevEUI and McGroupID.
func GetRemoteMulticastSetupByGroupID(db sqlx.Queryer, devEUI lorawan.EUI64, mcGroupID int, forUpdate bool) (RemoteMulticastSetup, error) {
	var fu string
	if forUpdate {
		fu = " for update"
	}

	var dmg RemoteMulticastSetup
	if err := sqlx.Get(db, &dmg, `
		select
			*
		from
			remote_multicast_setup
		where
			dev_eui = $1
			and mc_group_id = $2`+fu,
		devEUI,
		mcGroupID,
	); err != nil {
		return dmg, handlePSQLError(Select, err, "select error")
	}

	return dmg, nil
}

// GetPendingRemoteMulticastSetupItems returns a slice of pending remote multicast-setup items.
// The selected items will be locked.
func GetPendingRemoteMulticastSetupItems(db sqlx.Queryer, limit, maxRetryCount int) ([]RemoteMulticastSetup, error) {
	var items []RemoteMulticastSetup

	if err := sqlx.Select(db, &items, `
		select
			*
		from
			remote_multicast_setup
		where
			state_provisioned = false
			and retry_count < $1
			and retry_after < $2
		limit $3
		for update
		skip locked`,
		maxRetryCount,
		time.Now(),
		limit,
	); err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return items, nil
}

// UpdateRemoteMulticastSetup updates the given update multicast-group setup.
func UpdateRemoteMulticastSetup(db sqlx.Ext, dmg *RemoteMulticastSetup) error {
	dmg.UpdatedAt = time.Now()

	res, err := db.Exec(`
		update
			remote_multicast_setup
		set
			updated_at = $3,
			mc_group_id = $4,
			mc_addr = $5,
			mc_key_encrypted = $6,
			min_mc_f_cnt = $7,
			max_mc_f_cnt = $8,
			state = $9,
			state_provisioned = $10,
			retry_after = $11,
			retry_count = $12,
			retry_interval = $13
		where
			dev_eui = $1
			and multicast_group_id = $2`,
		dmg.DevEUI,
		dmg.MulticastGroupID,
		dmg.UpdatedAt,
		dmg.McGroupID,
		dmg.McAddr[:],
		dmg.McKeyEncrypted[:],
		dmg.MinMcFCnt,
		dmg.MaxMcFCnt,
		dmg.State,
		dmg.StateProvisioned,
		dmg.RetryAfter,
		dmg.RetryCount,
		dmg.RetryInterval,
	)
	if err != nil {
		return handlePSQLError(Update, err, "update error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	log.WithFields(log.Fields{
		"dev_eui":            dmg.DevEUI,
		"multicast_group_id": dmg.MulticastGroupID,
	}).Info("remote multicast-setup updated")
	return nil
}

// DeleteRemoteMulticastSetup deletes the multicast-setup given a multicast-group ID and DevEUI.
func DeleteRemoteMulticastSetup(db sqlx.Ext, devEUI lorawan.EUI64, multicastGroupID uuid.UUID) error {
	res, err := db.Exec(`
		delete from remote_multicast_setup
		where
			dev_eui = $1
			and multicast_group_id = $2`,
		devEUI,
		multicastGroupID,
	)
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

	log.WithFields(log.Fields{
		"dev_eui":            devEUI,
		"multicast_group_id": multicastGroupID,
	}).Info("remote multicast-setup deleted")
	return nil
}
