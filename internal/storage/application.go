package storage

import (
	"context"
	"fmt"
	"regexp"

	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	uuid "github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var applicationNameRegexp = regexp.MustCompile(`^[\w-]+$`)

// Application represents an application.
type Application struct {
	ID                   int64      `db:"id"`
	Name                 string     `db:"name"`
	Description          string     `db:"description"`
	OrganizationID       int64      `db:"organization_id"`
	ServiceProfileID     uuid.UUID  `db:"service_profile_id"`
	PayloadCodec         codec.Type `db:"payload_codec"`
	PayloadEncoderScript string     `db:"payload_encoder_script"`
	PayloadDecoderScript string     `db:"payload_decoder_script"`
}

// ApplicationListItem devices the application as a list item.
type ApplicationListItem struct {
	Application
	ServiceProfileName string `db:"service_profile_name"`
}

// Validate validates the data of the Application.
func (a Application) Validate() error {
	if !applicationNameRegexp.MatchString(a.Name) {
		return ErrApplicationInvalidName
	}

	return nil
}

// CreateApplication creates the given Application.
func CreateApplication(ctx context.Context, db sqlx.Queryer, item *Application) error {
	if err := item.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	err := sqlx.Get(db, &item.ID, `
		insert into application (
			name,
			description,
			organization_id,
			service_profile_id,
			payload_codec,
			payload_encoder_script,
			payload_decoder_script
		) values ($1, $2, $3, $4, $5, $6, $7) returning id`,
		item.Name,
		item.Description,
		item.OrganizationID,
		item.ServiceProfileID,
		item.PayloadCodec,
		item.PayloadEncoderScript,
		item.PayloadDecoderScript,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	log.WithFields(log.Fields{
		"id":     item.ID,
		"name":   item.Name,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("application created")

	return nil
}

// GetApplication returns the Application for the given id.
func GetApplication(ctx context.Context, db sqlx.Queryer, id int64) (Application, error) {
	var app Application
	err := sqlx.Get(db, &app, "select * from application where id = $1", id)
	if err != nil {
		return app, handlePSQLError(Select, err, "select error")
	}

	return app, nil
}

// GetApplicationCount returns the total number of applications.
func GetApplicationCount(ctx context.Context, db sqlx.Queryer, search string) (int, error) {
	var count int
	if search != "" {
		search = "%" + search + "%"
	}

	err := sqlx.Get(db, &count, `
		select
			count(*)
		from application
		where
			$1 = ''
			or ($1 != '' and name ilike $1)`,
		search,
	)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetApplicationCountForUser returns the total number of applications
// available for the given user.
// When an organizationID is given, the results will be filtered by this
// organization ID.
func GetApplicationCountForUser(ctx context.Context, db sqlx.Queryer, username string, organizationID int64, search string) (int, error) {
	var count int
	if search != "" {
		search = "%" + search + "%"
	}

	err := sqlx.Get(db, &count, `
		select
			count(a.*)
		from application a
		inner join organization_user ou
			on a.organization_id = ou.organization_id
		inner join "user" u
			on u.id = ou.user_id
		where
			u.username = $1
			and u.is_active = true
			and (
				$2 = 0
				or a.organization_id = $2
			)
			and (
				$3 = ''
				or ($3 != '' and a.name ilike $3)
			)
	`, username, organizationID, search)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetApplicationCountForOrganizationID returns the total number of
// applications for the given organization.
func GetApplicationCountForOrganizationID(ctx context.Context, db sqlx.Queryer, organizationID int64, search string) (int, error) {
	var count int
	if search != "" {
		search = "%" + search + "%"
	}

	err := sqlx.Get(db, &count, `
		select
			count(*)
		from application
		where
			organization_id = $1
			and (
				$2 = ''
				or ($2 != '' and name ilike $2)
			)`,
		organizationID,
		search,
	)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetApplications returns a slice of applications, sorted by name and
// respecting the given limit and offset.
func GetApplications(ctx context.Context, db sqlx.Queryer, limit, offset int, search string) ([]ApplicationListItem, error) {
	var apps []ApplicationListItem
	if search != "" {
		search = "%" + search + "%"
	}

	err := sqlx.Select(db, &apps, `
		select
			a.*,
			sp.name as service_profile_name
		from application a
		inner join service_profile sp
			on sp.service_profile_id = a.service_profile_id
		where
			$3 = ''
			or ($3 != '' and a.name ilike $3)
		order by
			name
		limit $1
		offset $2`,
		limit,
		offset,
		search,
	)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return apps, nil
}

// GetApplicationsForUser returns a slice of application of which the given
// user is a member of.
func GetApplicationsForUser(ctx context.Context, db sqlx.Queryer, username string, organizationID int64, limit, offset int, search string) ([]ApplicationListItem, error) {
	var apps []ApplicationListItem
	if search != "" {
		search = "%" + search + "%"
	}

	err := sqlx.Select(db, &apps, `
		select
			a.*,
			sp.name as service_profile_name
		from application a
		inner join service_profile sp
			on sp.service_profile_id = a.service_profile_id
		inner join organization_user ou
			on a.organization_id = ou.organization_id
		inner join "user" u
			on u.id = ou.user_id
		where
			u.username = $1
			and u.is_active = true
			and (
				$2 = 0
				or a.organization_id = $2
			)
			and (
				$5 = ''
				or ($5 != '' and a.name ilike $5)
			)
		order by a.name
		limit $3 offset $4
	`, username, organizationID, limit, offset, search)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}

	return apps, nil
}

// GetApplicationsForOrganizationID returns a slice of applications for the given
// organization.
func GetApplicationsForOrganizationID(ctx context.Context, db sqlx.Queryer, organizationID int64, limit, offset int, search string) ([]ApplicationListItem, error) {
	var apps []ApplicationListItem
	if search != "" {
		search = "%" + search + "%"
	}

	err := sqlx.Select(db, &apps, `
		select
			a.*,
			sp.name as service_profile_name
		from application a
		inner join service_profile sp
			on sp.service_profile_id = a.service_profile_id
		where
			a.organization_id = $1
			and (
				$4 = ''
				or ($4 != '' and a.name ilike $4)
			)
		order by a.name
		limit $2 offset $3`,
		organizationID,
		limit,
		offset,
		search,
	)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}

	return apps, nil
}

// UpdateApplication updates the given Application.
func UpdateApplication(ctx context.Context, db sqlx.Execer, item Application) error {
	if err := item.Validate(); err != nil {
		return fmt.Errorf("validate application error: %s", err)
	}

	res, err := db.Exec(`
		update application
		set
			name = $2,
			description = $3,
			organization_id = $4,
			service_profile_id = $5,
			payload_codec = $6,
			payload_encoder_script = $7,
			payload_decoder_script = $8
		where id = $1`,
		item.ID,
		item.Name,
		item.Description,
		item.OrganizationID,
		item.ServiceProfileID,
		item.PayloadCodec,
		item.PayloadEncoderScript,
		item.PayloadDecoderScript,
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
		"id":     item.ID,
		"name":   item.Name,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("application updated")

	return nil
}

// DeleteApplication deletes the Application matching the given ID.
func DeleteApplication(ctx context.Context, db sqlx.Ext, id int64) error {
	err := DeleteAllDevicesForApplicationID(ctx, db, id)
	if err != nil {
		return errors.Wrap(err, "delete all nodes error")
	}

	res, err := db.Exec("delete from application where id = $1", id)
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
	}).Info("application deleted")

	return nil
}

// DeleteAllApplicationsForOrganizationID deletes all applications
// given an organization id.
func DeleteAllApplicationsForOrganizationID(ctx context.Context, db sqlx.Ext, organizationID int64) error {
	var apps []Application
	err := sqlx.Select(db, &apps, "select * from application where organization_id = $1", organizationID)
	if err != nil {
		return handlePSQLError(Select, err, "select error")
	}

	for _, app := range apps {
		err = DeleteApplication(ctx, db, app.ID)
		if err != nil {
			return errors.Wrap(err, "delete application error")
		}
	}

	return nil
}
