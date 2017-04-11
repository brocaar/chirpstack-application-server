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

// OrganizationUser represents an organization user.
type OrganizationUser struct {
	UserID    int64     `db:"user_id"`
	Username  string    `db:"username"`
	IsAdmin   bool      `db:"is_admin"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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

// GetOrganizationCountForUser returns the number of organizations to which
// the given user is related to.
// The user has a relation to an organization:
// - when it has a reference to a specific application within the organization
// - when it has a reference to the organization itself
func GetOrganizationCountForUser(db *sqlx.DB, username string) (int, error) {
	var count int
	err := db.Get(&count, `
		select count(o.*)
		from organization o
		inner join "user" u
			on u.username = $1
		left join organization_user ou
			on o.id = ou.organization_id and u.id = ou.user_id
		left join application a
			on o.id = a.organization_id
		left join application_user au
			on a.id = au.application_id and u.id = au.user_id
		where
			au.user_id is not null or ou.user_id is not null`,
		username,
	)
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

// GetOrganizationsForUser returns a slice of organizations to which the given
// user is related to.
// The user has a relation to an organization:
// - when it has a reference to a specific application within the organization
// - when it has a reference to the organization itself
func GetOrganizationsForUser(db *sqlx.DB, username string, limit, offset int) ([]Organization, error) {
	var orgs []Organization
	err := db.Select(&orgs, `
		select o.*
		from organization o
		inner join "user" u
			on u.username = $1
		left join organization_user ou
			on o.id = ou.organization_id and u.id = ou.user_id
		left join application a
			on o.id = a.organization_id
		left join application_user au
			on a.id = au.application_id and u.id = au.user_id
		where
			au.user_id is not null or ou.user_id is not null
		order by name
		limit $2 offset $3`,
		username,
		limit,
		offset,
	)
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

// CreateOrganizationUser adds the given user to the organization.
func CreateOrganizationUser(db *sqlx.DB, organizationID, userID int64, isAdmin bool) error {
	_, err := db.Exec(`
		insert into organization_user (
			organization_id,
			user_id,
			is_admin,
			created_at,
			updated_at
		) values ($1, $2, $3, now(), now())`,
		organizationID,
		userID,
		isAdmin,
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

	log.WithFields(log.Fields{
		"user_id":         userID,
		"organization_id": organizationID,
		"is_admin":        isAdmin,
	}).Info("user added to organization")
	return nil
}

// UpdateOrganizationUser updates the given user of the organization.
func UpdateOrganizationUser(db *sqlx.DB, organizationID, userID int64, isAdmin bool) error {
	res, err := db.Exec(`
		update organization_user
		set
			is_admin = $3
		where
			organization_id = $1
			and user_id = $2
	`, organizationID, userID, isAdmin)
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

	log.WithFields(log.Fields{
		"user_id":         userID,
		"organization_id": organizationID,
		"is_admin":        isAdmin,
	}).Info("organization user updated")
	return nil
}

// DeleteOrganizationUser deletes the given organization user.
func DeleteOrganizationUser(db *sqlx.DB, organizationID, userID int64) error {
	res, err := db.Exec(`delete from organization_user where organization_id = $1 and user_id = $2`, organizationID, userID)
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	log.WithFields(log.Fields{
		"user_id":         userID,
		"organization_id": organizationID,
	}).Info("organization user deleted")
	return nil
}

// GetOrganizationUser gets the information of the given organization user.
func GetOrganizationUser(db *sqlx.DB, organizationID, userID int64) (OrganizationUser, error) {
	var u OrganizationUser
	err := db.Get(&u, `
		select
			u.id as user_id,
			u.username as username,
			ou.created_at as created_at,
			ou.updated_at as updated_at,
			ou.is_admin as is_admin
		from organization_user ou
		inner join "user" u
			on u.id = ou.user_id
		where
			ou.organization_id = $1
			and ou.user_id = $2`,
		organizationID,
		userID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return u, ErrDoesNotExist
		}
		return u, errors.Wrap(err, "select error")
	}
	return u, nil
}

// GetOrganizationUserCount returns the number of users for the given organization.
func GetOrganizationUserCount(db *sqlx.DB, organizationID int64) (int, error) {
	var count int
	err := db.Get(&count, `
		select count(*)
		from organization_user
		where
			organization_id = $1`,
		organizationID,
	)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetOrganizationUsers returns the users for the given organization.
func GetOrganizationUsers(db *sqlx.DB, organizationID int64, limit, offset int) ([]OrganizationUser, error) {
	var users []OrganizationUser
	err := db.Select(&users, `
		select
			u.id as user_id,
			u.username as username,
			ou.created_at as created_at,
			ou.updated_at as updated_at,
			ou.is_admin as is_admin
		from organization_user ou
		inner join "user" u
			on u.id = ou.user_id
		where
			ou.organization_id = $1
		order by u.username
		limit $2 offset $3`,
		organizationID,
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return users, nil
}
