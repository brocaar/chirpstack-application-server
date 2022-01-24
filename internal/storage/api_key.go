package storage

import (
	"context"
	"strings"
	"time"

	uuid "github.com/gofrs/uuid"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-application-server/internal/logging"
)

// APIKey represents an API key.
type APIKey struct {
	ID             uuid.UUID `db:"id"`
	CreatedAt      time.Time `db:"created_at"`
	Name           string    `db:"name"`
	IsAdmin        bool      `db:"is_admin"`
	OrganizationID *int64    `db:"organization_id"`
	ApplicationID  *int64    `db:"application_id"`
}

// Validate validates the given API Key data.
func (a APIKey) Validate() error {
	if strings.TrimSpace(a.Name) == "" || len(a.Name) > 100 {
		return ErrAPIKeyInvalidName
	}
	return nil
}

// CreateAPIKey creates the given API key and returns the JWT.
func CreateAPIKey(ctx context.Context, db sqlx.Ext, a *APIKey) (string, error) {
	if err := a.Validate(); err != nil {
		return "", errors.Wrap(err, "validate error")
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", errors.Wrap(err, "new uuid error")
	}

	a.ID = id
	a.CreatedAt = time.Now()

	_, err = db.Exec(`
		insert into api_key (
			id,
			created_at,
			name,
			is_admin,
			organization_id,
			application_id
		) values ($1, $2, $3, $4, $5, $6)`,
		a.ID,
		a.CreatedAt,
		a.Name,
		a.IsAdmin,
		a.OrganizationID,
		a.ApplicationID,
	)
	if err != nil {
		return "", handlePSQLError(Insert, err, "insert error")
	}

	log.WithFields(log.Fields{
		"ctx_id": ctx.Value(logging.ContextIDKey),
		"id":     a.ID,
	}).Info("storage: api-key created")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":        "as",
		"aud":        "as",
		"nbf":        time.Now().Unix(),
		"sub":        "api_key",
		"api_key_id": a.ID.String(),
	})

	jwt, err := token.SignedString(jwtsecret)
	if err != nil {
		return jwt, errors.Wrap(err, "sign jwt token error")
	}

	return jwt, nil
}

// GetAPIKey returns the API key for the given ID.
func GetAPIKey(ctx context.Context, db sqlx.Queryer, id uuid.UUID) (APIKey, error) {
	var a APIKey

	err := sqlx.Get(db, &a, `
		select
			*
		from
			api_key
		where
			id = $1`,
		id,
	)
	if err != nil {
		return a, handlePSQLError(Select, err, "select error")
	}

	return a, nil
}

// DeleteAPIKey deletes the API key for the given ID.
func DeleteAPIKey(ctx context.Context, db sqlx.Ext, id uuid.UUID) error {
	res, err := db.Exec(`
		delete
		from
			api_key
		where
			id = $1`,
		id,
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
		"ctx_id": ctx.Value(logging.ContextIDKey),
		"id":     id,
	}).Info("storage: api-key deleted")
	return nil
}

// APIKeyFilters provides filters for getting the API keys.
type APIKeyFilters struct {
	IsAdmin        bool   `db:"is_admin"`
	OrganizationID *int64 `db:"organization_id"`
	ApplicationID  *int64 `db:"application_id"`

	// Limit and Offset are added for convenience so that this struct can
	// be given as the arguments.
	Limit  int `db:"limit"`
	Offset int `db:"offset"`
}

// SQL returns the filters as SQL.
func (f APIKeyFilters) SQL() string {
	var filters []string

	filters = append(filters, "is_admin = :is_admin")

	if f.OrganizationID != nil {
		filters = append(filters, "organization_id = :organization_id")
	}

	if f.ApplicationID != nil {
		filters = append(filters, "application_id = :application_id")
	}

	return "where " + strings.Join(filters, " and ")
}

// GetAPIKeyCount returns the number of API keys.
func GetAPIKeyCount(ctx context.Context, db sqlx.Queryer, filters APIKeyFilters) (int, error) {
	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			count(*)
		from
			api_key
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

// GetAPIKeys returns a slice of API keys.
func GetAPIKeys(ctx context.Context, db sqlx.Queryer, filters APIKeyFilters) ([]APIKey, error) {
	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			*
		from
			api_key
	`+filters.SQL()+`
		order by
			name
		limit :limit
		offset :offset
	`, filters)
	if err != nil {
		return nil, errors.Wrap(err, "named query error")
	}

	var keys []APIKey
	err = sqlx.Select(db, &keys, query, args...)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return keys, nil
}
