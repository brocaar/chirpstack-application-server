package storage

import (
	"database/sql"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

var organizationNameRegexp = regexp.MustCompile(`^[\w-]+$`)

// Organization represents an organization.
type Organization struct {
	ID              int64     `db:"id"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
	Name            string    `db:"name"`
	DisplayName     string    `db:"display_name"`
	CanHaveGateways bool      `db:"can_have_gateways"`
}

// Validate validates the data of the Organization.
func (o Organization) Validate() error {
	if !organizationNameRegexp.MatchString(o.Name) {
		return ErrOrganizationInvalidName
	}
	return nil
}

// CreateOrganization creates the given Organization.
func CreateOrganization(db *sqlx.DB, org *Organization) error {
	if err := org.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	now := time.Now()

	err := db.Get(&org.ID, `
		insert into organization (
			created_at,
			updated_at,
			name,
			display_name,
			can_have_gateways
		) values ($1, $2, $3, $4, $5) returning id`,
		now,
		now,
		org.Name,
		org.DisplayName,
		org.CanHaveGateways,
	)
	if err != nil {
		switch err := err.(type) {
		case *pq.Error:
			switch err.Code.Name() {
			case "unique_violation":
				return ErrAlreadyExists
			default:
				return errors.Wrap(err, "insert error")
			}
		default:
			return errors.Wrap(err, "insert error")
		}
	}
	org.CreatedAt = now
	org.UpdatedAt = now
	log.WithFields(log.Fields{
		"id":   org.ID,
		"name": org.Name,
	}).Info("organization created")
	return nil
}

// GetOrganization returns the Organization for the given id.
func GetOrganization(db *sqlx.DB, id int64) (Organization, error) {
	var org Organization
	err := db.Get(&org, "select * from organization where id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return org, ErrDoesNotExist
		}
		return org, errors.Wrap(err, "select error")
	}
	return org, nil
}

// GetOrganizationCount returns the total number of organizations.
func GetOrganizationCount(db *sqlx.DB) (int, error) {
	var count int
	err := db.Get(&count, "select count(*) from organization")
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetOrganizations returns a slice of organizations, sorted by name and
// respecting the given limit and offset.
func GetOrganizations(db *sqlx.DB, limit, offset int) ([]Organization, error) {
	var orgs []Organization
	err := db.Select(&orgs, "select * from organization order by name limit $1 offset $2", limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return orgs, nil
}

// UpdateOrganization updates the given organization.
func UpdateOrganization(db *sqlx.DB, org *Organization) error {
	if err := org.Validate(); err != nil {
		return errors.Wrap(err, "validation error")
	}

	now := time.Now()
	res, err := db.Exec(`
		update organization
		set
			name = $2,
			display_name = $3,
			can_have_gateways = $4,
			updated_at = $5
		where id = $1`,
		org.ID,
		org.Name,
		org.DisplayName,
		org.CanHaveGateways,
		now,
	)

	if err != nil {
		switch err := err.(type) {
		case *pq.Error:
			switch err.Code.Name() {
			case "unique_violation":
				return ErrAlreadyExists
			default:
				return errors.Wrap(err, "update error")
			}
		default:
			return errors.Wrap(err, "update error")
		}
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	org.UpdatedAt = now
	log.WithFields(log.Fields{
		"name": org.Name,
		"id":   org.ID,
	}).Info("organization updated")
	return nil
}

// DeleteOrganization deletes the organization matching the given id.
func DeleteOrganization(db *sqlx.DB, id int64) error {
	res, err := db.Exec("delete from organization where id = $1", id)
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

	log.WithField("id", id).Info("organization deleted")
	return nil
}
