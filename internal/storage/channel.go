package storage

import (
	"database/sql"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// ChannelList represents a list of channels.
// This list will be used for the FCList field (if allowed for the used band).
type ChannelList struct {
	ID       int64   `db:"id"`
	Name     string  `db:"name"`
	Channels []int64 `db:"channels"`
}

// CreateChannelList creates the given ChannelList.
func CreateChannelList(db *sqlx.DB, cl *ChannelList) error {
	err := db.Get(&cl.ID, "insert into channel_list (name, channels) values ($1, $2) returning id",
		cl.Name,
		pq.Int64Array(cl.Channels),
	)
	if err != nil {
		return errors.Wrap(err, "insert error")
	}
	log.WithFields(log.Fields{
		"id":   cl.ID,
		"name": cl.Name,
	}).Info("channel-list created")
	return nil
}

// UpdateChannelList updates the given ChannelList.
func UpdateChannelList(db *sqlx.DB, cl ChannelList) error {
	res, err := db.Exec("update channel_list set name = $1, channels = $2 where id = $3",
		cl.Name,
		pq.Array(cl.Channels),
		cl.ID,
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
	log.WithField("id", cl.ID).Info("channel-list updated")
	return nil
}

// GetChannelList returns the ChannelList for the given id.
func GetChannelList(db *sqlx.DB, id int64) (ChannelList, error) {
	var cl ChannelList
	err := db.QueryRow("select id, name, channels from channel_list where id = $1", id).Scan(&cl.ID, &cl.Name, pq.Array(&cl.Channels))
	if err != nil {
		if err == sql.ErrNoRows {
			return cl, ErrDoesNotExist
		}
		return cl, errors.Wrap(err, "select error")
	}
	return cl, nil
}

// GetChannelLists returns a list of ChannelList items.
func GetChannelLists(db *sqlx.DB, limit, offset int) ([]ChannelList, error) {
	var channelLists []ChannelList
	rows, err := db.Query("select id, name, channels from channel_list order by name limit $1 offset $2", limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	defer rows.Close()
	for rows.Next() {
		var cl ChannelList
		if err := rows.Scan(&cl.ID, &cl.Name, pq.Array(&cl.Channels)); err != nil {
			return nil, errors.Wrap(err, "scan row error")
		}
		channelLists = append(channelLists, cl)
	}
	return channelLists, nil
}

// GetChannelListsCount returns the total number of channel-lists.
func GetChannelListsCount(db *sqlx.DB) (int, error) {
	var count struct {
		Count int
	}
	err := db.Get(&count, "select count(*) as count from channel_list")
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count.Count, nil
}

// DeleteChannelList deletes the ChannelList matching the given id.
func DeleteChannelList(db *sqlx.DB, id int64) error {
	res, err := db.Exec("delete from channel_list where id = $1",
		id,
	)
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
	log.WithField("id", id).Info("channel-list deleted")
	return nil
}
