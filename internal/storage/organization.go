package storage

import (
	"context"
	"regexp"
	"time"

	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
	UserID         int64     `db:"user_id"`
	Username       string    `db:"username"`
	IsAdmin        bool      `db:"is_admin"`
	IsDeviceAdmin  bool      `db:"is_device_admin"`
	IsGatewayAdmin bool      `db:"is_gateway_admin"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// CreateOrganization creates the given Organization.
func CreateOrganization(ctx context.Context, db sqlx.Queryer, org *Organization) error {
	if err := org.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	now := time.Now()

	err := sqlx.Get(db, &org.ID, `
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
		return handlePSQLError(Insert, err, "insert error")
	}
	org.CreatedAt = now
	org.UpdatedAt = now
	log.WithFields(log.Fields{
		"id":     org.ID,
		"name":   org.Name,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("organization created")
	return nil
}

// GetOrganization returns the Organization for the given id.
func GetOrganization(ctx context.Context, db sqlx.Queryer, id int64) (Organization, error) {
	var org Organization
	err := sqlx.Get(db, &org, "select * from organization where id = $1", id)
	if err != nil {
		return org, handlePSQLError(Select, err, "select error")
	}
	return org, nil
}

// GetOrganizationCount returns the total number of organizations.
func GetOrganizationCount(ctx context.Context, db sqlx.Queryer, search string) (int, error) {
	var count int

	if search != "" {
		search = "%" + search + "%"
	}

	err := sqlx.Get(db, &count, `
		select count(*)
		from organization
		where
			($1 != '' and display_name ilike $1)
			or ($1 = '')`,
		search,
	)
	if err != nil {
		return count, handlePSQLError(Select, err, "select error")
	}
	return count, nil
}

// GetOrganizationCountForUser returns the number of organizations to which
// the given user is member of.
func GetOrganizationCountForUser(ctx context.Context, db sqlx.Queryer, username string, search string) (int, error) {
	var count int

	if search != "" {
		search = "%" + search + "%"
	}

	err := sqlx.Get(db, &count, `
		select
			count(o.*)
		from organization o
		inner join organization_user ou
			on ou.organization_id = o.id
		inner join "user" u
			on u.id = ou.user_id
		where
			u.username = $1
			and (
				($2 != '' and o.display_name ilike $2)
				or ($2 = '')
			)`,
		username,
		search,
	)
	if err != nil {
		return count, handlePSQLError(Select, err, "select error")
	}
	return count, nil
}

// GetOrganizations returns a slice of organizations, sorted by name and
// respecting the given limit and offset.
func GetOrganizations(ctx context.Context, db sqlx.Queryer, limit, offset int, search string) ([]Organization, error) {
	var orgs []Organization

	if search != "" {
		search = "%" + search + "%"
	}

	err := sqlx.Select(db, &orgs, `
		select *
		from organization
		where
			($3 != '' and display_name ilike $3)
			or ($3 = '')
		order by display_name
		limit $1 offset $2`, limit, offset, search)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}
	return orgs, nil
}

// GetOrganizationsForUser returns a slice of organizations to which the given
// user is member of.
func GetOrganizationsForUser(ctx context.Context, db sqlx.Queryer, username string, limit, offset int, search string) ([]Organization, error) {
	var orgs []Organization

	if search != "" {
		search = "%" + search + "%"
	}

	err := sqlx.Select(db, &orgs, `
		select
			o.*
		from organization o
		inner join organization_user ou
			on ou.organization_id = o.id
		inner join "user" u
			on u.id = ou.user_id
		where
			u.username = $1
			and (
				($4 != '' and o.display_name ilike $4)
				or ($4 = '')
			)
		order by o.display_name
		limit $2 offset $3`,
		username,
		limit,
		offset,
		search,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}
	return orgs, nil
}

// UpdateOrganization updates the given organization.
func UpdateOrganization(ctx context.Context, db sqlx.Execer, org *Organization) error {
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
		return handlePSQLError(Update, err, "update error")
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
		"name":   org.Name,
		"id":     org.ID,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("organization updated")
	return nil
}

// DeleteOrganization deletes the organization matching the given id.
func DeleteOrganization(ctx context.Context, db sqlx.Ext, id int64) error {
	err := DeleteAllApplicationsForOrganizationID(ctx, db, id)
	if err != nil {
		return errors.Wrap(err, "delete all applications error")
	}

	err = DeleteAllServiceProfilesForOrganizationID(ctx, db, id)
	if err != nil {
		return errors.Wrap(err, "delete all service-profiles error")
	}

	err = DeleteAllDeviceProfilesForOrganizationID(ctx, db, id)
	if err != nil {
		return errors.Wrap(err, "delete all device-profiles error")
	}

	res, err := db.Exec("delete from organization where id = $1", id)
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
		"id":     id,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("organization deleted")
	return nil
}

// CreateOrganizationUser adds the given user to the organization.
func CreateOrganizationUser(ctx context.Context, db sqlx.Execer, organizationID, userID int64, isAdmin, isDeviceAdmin, isGatewayAdmin bool) error {
	_, err := db.Exec(`
		insert into organization_user (
			organization_id,
			user_id,
			is_admin,
			is_device_admin,
			is_gateway_admin,
			created_at,
			updated_at
		) values ($1, $2, $3, $4, $5, now(), now())`,
		organizationID,
		userID,
		isAdmin,
		isDeviceAdmin,
		isGatewayAdmin,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	log.WithFields(log.Fields{
		"user_id":          userID,
		"organization_id":  organizationID,
		"is_admin":         isAdmin,
		"is_device_admin":  isDeviceAdmin,
		"is_gateway_admin": isGatewayAdmin,
		"ctx_id":           ctx.Value(logging.ContextIDKey),
	}).Info("user added to organization")
	return nil
}

// UpdateOrganizationUser updates the given user of the organization.
func UpdateOrganizationUser(ctx context.Context, db sqlx.Execer, organizationID, userID int64, isAdmin, isDeviceAdmin, isGatewayAdmin bool) error {
	res, err := db.Exec(`
		update organization_user
		set
			is_admin = $3,
			is_device_admin = $4,
			is_gateway_admin = $5
		where
			organization_id = $1
			and user_id = $2
	`, organizationID, userID, isAdmin, isDeviceAdmin, isGatewayAdmin)
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
		"user_id":          userID,
		"organization_id":  organizationID,
		"is_admin":         isAdmin,
		"is_device_admin":  isDeviceAdmin,
		"is_gateway_admin": isGatewayAdmin,
		"ctx_id":           ctx.Value(logging.ContextIDKey),
	}).Info("organization user updated")
	return nil
}

// DeleteOrganizationUser deletes the given organization user.
func DeleteOrganizationUser(ctx context.Context, db sqlx.Execer, organizationID, userID int64) error {
	res, err := db.Exec(`delete from organization_user where organization_id = $1 and user_id = $2`, organizationID, userID)
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
		"user_id":         userID,
		"organization_id": organizationID,
		"ctx_id":          ctx.Value(logging.ContextIDKey),
	}).Info("organization user deleted")
	return nil
}

// GetOrganizationUser gets the information of the given organization user.
func GetOrganizationUser(ctx context.Context, db sqlx.Queryer, organizationID, userID int64) (OrganizationUser, error) {
	var u OrganizationUser
	err := sqlx.Get(db, &u, `
		select
			u.id as user_id,
			u.username as username,
			ou.created_at as created_at,
			ou.updated_at as updated_at,
			ou.is_admin as is_admin,
			ou.is_device_admin as is_device_admin,
			ou.is_gateway_admin as is_gateway_admin
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
		return u, handlePSQLError(Select, err, "select error")
	}
	return u, nil
}

// GetOrganizationUserCount returns the number of users for the given organization.
func GetOrganizationUserCount(ctx context.Context, db sqlx.Queryer, organizationID int64) (int, error) {
	var count int
	err := sqlx.Get(db, &count, `
		select count(*)
		from organization_user
		where
			organization_id = $1`,
		organizationID,
	)
	if err != nil {
		return count, handlePSQLError(Select, err, "select error")
	}
	return count, nil
}

// GetOrganizationUsers returns the users for the given organization.
func GetOrganizationUsers(ctx context.Context, db sqlx.Queryer, organizationID int64, limit, offset int) ([]OrganizationUser, error) {
	var users []OrganizationUser
	err := sqlx.Select(db, &users, `
		select
			u.id as user_id,
			u.username as username,
			ou.created_at as created_at,
			ou.updated_at as updated_at,
			ou.is_admin as is_admin,
			ou.is_device_admin as is_device_admin,
			ou.is_gateway_admin as is_gateway_admin
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
		return nil, handlePSQLError(Select, err, "select error")
	}
	return users, nil
}
