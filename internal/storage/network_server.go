package storage

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// NetworkServer defines the information to connect to a network-server.
type NetworkServer struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Name      string    `db:"name"`
	Server    string    `db:"server"`
}

// Validate validates the network-server data.
func (ns NetworkServer) Validate() error {
	return nil
}

// CreateNetworkServer creates the given network-server.
func CreateNetworkServer(db sqlx.Queryer, ns *NetworkServer) error {
	if err := ns.Validate(); err != nil {
		return errors.Wrap(err, "validation error")
	}

	now := time.Now()
	ns.CreatedAt = now
	ns.UpdatedAt = now

	err := sqlx.Get(db, &ns.ID, `
		insert into network_server (
			created_at,
			updated_at,
			name,
			server
		) values ($1, $2, $3, $4)
		returning id`,
		ns.CreatedAt,
		ns.UpdatedAt,
		ns.Name,
		ns.Server,
	)
	if err != nil {
		return handlePSQLError(err, "insert error")
	}

	log.WithFields(log.Fields{
		"id":     ns.ID,
		"name":   ns.Name,
		"server": ns.Server,
	}).Info("network-server created")
	return nil
}

// GetNetworkServer returns the network-server matching the given id.
func GetNetworkServer(db sqlx.Queryer, id int64) (NetworkServer, error) {
	var ns NetworkServer
	err := sqlx.Get(db, &ns, "select * from network_server where id = $1", id)
	if err != nil {
		return ns, handlePSQLError(err, "select error")
	}

	return ns, nil
}

// UpdateNetworkServer updates the given network-server.
func UpdateNetworkServer(db sqlx.Execer, ns *NetworkServer) error {
	if err := ns.Validate(); err != nil {
		return errors.Wrap(err, "validation error")
	}

	ns.UpdatedAt = time.Now()

	res, err := db.Exec(`
		update network_server
		set
			updated_at = $2,
			name = $3,
			server = $4
		where id = $1`,
		ns.ID,
		ns.UpdatedAt,
		ns.Name,
		ns.Server,
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

	log.WithFields(log.Fields{
		"id":     ns.ID,
		"name":   ns.Name,
		"server": ns.Server,
	}).Info("network-server updated")
	return nil
}

// DeleteNetworkServer deletes the network-server matching the given id.
func DeleteNetworkServer(db sqlx.Execer, id int64) error {
	res, err := db.Exec("delete from network_server where id = $1", id)
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

	log.WithField("id", id).Info("network-server deleted")
	return nil
}
