package storage

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	uuid "github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
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
	MQTTTLSCert          []byte     `db:"mqtt_tls_cert"`
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
			payload_decoder_script,
			mqtt_tls_cert
		) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id`,
		item.Name,
		item.Description,
		item.OrganizationID,
		item.ServiceProfileID,
		item.PayloadCodec,
		item.PayloadEncoderScript,
		item.PayloadDecoderScript,
		item.MQTTTLSCert,
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

// ApplicationFilters provides filters for filtering applications.
type ApplicationFilters struct {
	UserID         int64  `db:"user_id"`
	OrganizationID int64  `db:"organization_id"`
	Search         string `db:"search"`

	// Limit and Offset are added for convenience so that this struct can
	// be given as the arguments.
	Limit  int `db:"limit"`
	Offset int `db:"offset"`
}

// SQL returns the SQL filters.
func (f ApplicationFilters) SQL() string {
	var filters []string

	if f.UserID != 0 {
		filters = append(filters, "u.id = :user_id")
	}

	if f.OrganizationID != 0 {
		filters = append(filters, "a.organization_id = :organization_id")
	}

	if f.Search != "" {
		filters = append(filters, "(a.name ilike :search)")
	}

	if len(filters) == 0 {
		return ""
	}

	return "where " + strings.Join(filters, " and ")
}

// GetApplicationCount returns the total number of applications.
func GetApplicationCount(ctx context.Context, db sqlx.Queryer, filters ApplicationFilters) (int, error) {
	if filters.Search != "" {
		filters.Search = "%" + filters.Search + "%"
	}

	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			count(distinct a.*)
		from
			application a
		left join organization_user ou
			on a.organization_id = ou.organization_id
		left join "user" u
			on ou.user_id = u.id
	`+filters.SQL(), filters)
	if err != nil {
		return 0, errors.Wrap(err, "named query error")
	}

	var count int
	err = sqlx.Get(db, &count, query, args...)
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}

	return count, nil
}

// GetApplications returns a slice of applications, sorted by name and
// respecting the given limit and offset.
func GetApplications(ctx context.Context, db sqlx.Queryer, filters ApplicationFilters) ([]ApplicationListItem, error) {
	if filters.Search != "" {
		filters.Search = "%" + filters.Search + "%"
	}

	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			a.*,
			sp.name as service_profile_name
		from
			application a
		inner join service_profile sp
			on a.service_profile_id = sp.service_profile_id
		left join organization_user ou
			on a.organization_id = ou.organization_id
		left join "user" u
			on ou.user_id = u.id
	`+filters.SQL()+`
		group by
			a.id,
			sp.name
		order by
			a.name
		limit :limit
		offset :offset
	`, filters)
	if err != nil {
		return nil, errors.Wrap(err, "named query error")
	}

	var apps []ApplicationListItem
	err = sqlx.Select(db, &apps, query, args...)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
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
			payload_decoder_script = $8,
			mqtt_tls_cert = $9
		where id = $1`,
		item.ID,
		item.Name,
		item.Description,
		item.OrganizationID,
		item.ServiceProfileID,
		item.PayloadCodec,
		item.PayloadEncoderScript,
		item.PayloadDecoderScript,
		item.MQTTTLSCert,
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
