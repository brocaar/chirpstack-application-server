package storage

import (
	"database/sql"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

var gatewayNameRegexp = regexp.MustCompile(`^[\w-]+$`)

// Gateway represents a gateway.
type Gateway struct {
	MAC            lorawan.EUI64 `db:"mac"`
	CreatedAt      time.Time     `db:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at"`
	Name           string        `db:"name"`
	Description    string        `db:"description"`
	OrganizationID int64         `db:"organization_id"`
}

// Validate validates the gateway data.
func (g Gateway) Validate() error {
	if !gatewayNameRegexp.MatchString(g.Name) {
		return ErrGatewayInvalidName
	}
	return nil
}

// CreateGateway creates the given Gateway.
func CreateGateway(db sqlx.Execer, gw *Gateway) error {
	if err := gw.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	now := time.Now()

	_, err := db.Exec(`
		insert into gateway (
			mac,
			created_at,
			updated_at,
			name,
			description,
			organization_id
		) values ($1, $2, $3, $4, $5, $6)`,
		gw.MAC[:],
		now,
		now,
		gw.Name,
		gw.Description,
		gw.OrganizationID,
	)
	if err != nil {
		switch err := err.(type) {
		case *pq.Error:
			switch err.Code.Name() {
			case "unique_violation":
				return ErrAlreadyExists
			case "foreign_key_violation":
				return ErrDoesNotExist
			default:
				return errors.Wrap(err, "insert error")
			}
		default:
			return errors.Wrap(err, "insert error")
		}
	}

	gw.CreatedAt = now
	gw.UpdatedAt = now

	log.WithFields(log.Fields{
		"mac":  gw.MAC,
		"name": gw.Name,
	}).Info("gateway created")
	return nil
}

// UpdateGateway updates the given Gateway.
func UpdateGateway(db sqlx.Execer, gw *Gateway) error {
	if err := gw.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	now := time.Now()

	res, err := db.Exec(`
		update gateway
			set updated_at = $2,
			name = $3,
			description = $4,
			organization_id = $5
		where
			mac = $1`,
		gw.MAC[:],
		now,
		gw.Name,
		gw.Description,
		gw.OrganizationID,
	)
	if err != nil {
		switch err := err.(type) {
		case *pq.Error:
			switch err.Code.Name() {
			case "unique_violation":
				return ErrAlreadyExists
			case "foreign_key_violation":
				return ErrDoesNotExist
			default:
				return errors.Wrap(err, "insert error")
			}
		default:
			return errors.Wrap(err, "insert error")
		}
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	gw.UpdatedAt = now
	log.WithFields(log.Fields{
		"mac":  gw.MAC,
		"name": gw.Name,
	}).Info("gateway updated")
	return nil
}

// DeleteGateway deletes the gateway matching the given MAC.
func DeleteGateway(db sqlx.Execer, mac lorawan.EUI64) error {
	res, err := db.Exec("delete from gateway where mac = $1", mac[:])
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
	log.WithField("mac", mac).Info("gateway deleted")
	return nil
}

// GetGateway returns the gateway for the given mac.
func GetGateway(db *sqlx.DB, mac lorawan.EUI64) (Gateway, error) {
	var gw Gateway
	err := db.Get(&gw, "select * from gateway where mac = $1", mac[:])
	if err != nil {
		if err == sql.ErrNoRows {
			return gw, ErrDoesNotExist
		}
	}
	return gw, nil
}

// GetGatewayCount returns the total number of gateways.
func GetGatewayCount(db *sqlx.DB) (int, error) {
	var count int
	err := db.Get(&count, "select count(*) from gateway")
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetGateways returns a slice of gateways sorted by name.
func GetGateways(db *sqlx.DB, limit, offset int) ([]Gateway, error) {
	var gws []Gateway
	err := db.Select(&gws, `
		select *
		from gateway
		order by name
		limit $1 offset $2`,
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return gws, nil
}

// GetGatewayCountForOrganizationID returns the total number of gateways
// given an organization ID.
func GetGatewayCountForOrganizationID(db *sqlx.DB, organizationID int64) (int, error) {
	var count int
	err := db.Get(&count, `
		select count(*)
		from gateway
		where
			organization_id = $1`,
		organizationID,
	)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetGatewaysForOrganizationID returns a slice of gateways sorted by name
// for the given organization ID.
func GetGatewaysForOrganizationID(db *sqlx.DB, organizationID int64, limit, offset int) ([]Gateway, error) {
	var gws []Gateway
	err := db.Select(&gws, `
		select *
		from gateway
		where
			organization_id = $1
		order by name
		limit $2 offset $3`,
		organizationID,
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return gws, nil
}

// GetGatewayCountForUser returns the total number of gateways to which the
// given user has access.
func GetGatewayCountForUser(db *sqlx.DB, username string) (int, error) {
	var count int
	err := db.Get(&count, `
		select count(g.*)
		from gateway g
		inner join organization o
			on o.id = g.organization_id
		inner join organization_user ou
			on ou.organization_id = o.id
		inner join "user" u
			on u.id = ou.user_id
		where
			u.username = $1`,
		username,
	)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetGatewaysForUser returns a slice of gateways sorted by name to which the
// given user has access.
func GetGatewaysForUser(db *sqlx.DB, username string, limit, offset int) ([]Gateway, error) {
	var gws []Gateway
	err := db.Select(&gws, `
		select g.*
		from gateway g
		inner join organization o
			on o.id = g.organization_id
		inner join organization_user ou
			on ou.organization_id = o.id
		inner join "user" u
			on u.id = ou.user_id
		where
			u.username = $1
		order by g.name
		limit $2 offset $3`,
		username,
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return gws, nil
}
