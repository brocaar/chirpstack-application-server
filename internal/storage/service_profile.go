package storage

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
)

// ServiceProfile defines the service-profile.
type ServiceProfile struct {
	NetworkServerID int64             `db:"network_server_id"`
	OrganizationID  int64             `db:"organization_id"`
	CreatedAt       time.Time         `db:"created_at"`
	UpdatedAt       time.Time         `db:"updated_at"`
	Name            string            `db:"name"`
	ServiceProfile  ns.ServiceProfile `db:"-"`
}

// ServiceProfileMeta defines the service-profile meta record.
type ServiceProfileMeta struct {
	ServiceProfileID  uuid.UUID `db:"service_profile_id"`
	NetworkServerID   int64     `db:"network_server_id"`
	OrganizationID    int64     `db:"organization_id"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
	Name              string    `db:"name"`
	NetworkServerName string    `db:"network_server_name"`
}

// Validate validates the service-profile data.
func (sp ServiceProfile) Validate() error {
	if strings.TrimSpace(sp.Name) == "" || len(sp.Name) > 100 {
		return ErrServiceProfileInvalidName
	}
	return nil
}

// CreateServiceProfile creates the given service-profile.
func CreateServiceProfile(ctx context.Context, db sqlx.Ext, sp *ServiceProfile) error {
	if err := sp.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	spID, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid v4 error")
	}

	now := time.Now()
	sp.CreatedAt = now
	sp.UpdatedAt = now
	sp.ServiceProfile.Id = spID.Bytes()

	_, err = db.Exec(`
		insert into service_profile (
			service_profile_id,
			network_server_id,
			organization_id,
			created_at,
			updated_at,
			name
		) values ($1, $2, $3, $4, $5, $6)`,
		spID,
		sp.NetworkServerID,
		sp.OrganizationID,
		sp.CreatedAt,
		sp.UpdatedAt,
		sp.Name,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	n, err := GetNetworkServer(ctx, db, sp.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.CreateServiceProfile(ctx, &ns.CreateServiceProfileRequest{
		ServiceProfile: &sp.ServiceProfile,
	})
	if err != nil {
		return errors.Wrap(err, "create service-profile error")
	}

	log.WithFields(log.Fields{
		"id":     spID,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("service-profile created")
	return nil
}

// GetServiceProfile returns the service-profile matching the given id.
func GetServiceProfile(ctx context.Context, db sqlx.Queryer, id uuid.UUID, localOnly bool) (ServiceProfile, error) {
	var sp ServiceProfile
	row := db.QueryRowx(`
		select
			network_server_id,
			organization_id,
			created_at,
			updated_at,
			name
		from service_profile
		where
			service_profile_id = $1`,
		id,
	)
	if err := row.Err(); err != nil {
		return sp, handlePSQLError(Select, err, "select error")
	}

	err := row.Scan(&sp.NetworkServerID, &sp.OrganizationID, &sp.CreatedAt, &sp.UpdatedAt, &sp.Name)
	if err != nil {
		return sp, handlePSQLError(Scan, err, "scan error")
	}

	if localOnly {
		return sp, nil
	}

	n, err := GetNetworkServer(ctx, db, sp.NetworkServerID)
	if err != nil {
		return sp, errors.Wrap(err, "get network-server errror")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return sp, errors.Wrap(err, "get network-server client error")
	}

	resp, err := nsClient.GetServiceProfile(ctx, &ns.GetServiceProfileRequest{
		Id: id.Bytes(),
	})
	if err != nil {
		return sp, errors.Wrap(err, "get service-profile error")
	}

	if resp.ServiceProfile == nil {
		return sp, errors.New("service_profile must not be nil")
	}

	sp.ServiceProfile = *resp.ServiceProfile

	return sp, nil
}

// UpdateServiceProfile updates the given service-profile.
func UpdateServiceProfile(ctx context.Context, db sqlx.Ext, sp *ServiceProfile) error {
	if err := sp.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
	if err != nil {
		return errors.Wrap(err, "uuid from bytes error")
	}

	sp.UpdatedAt = time.Now()
	res, err := db.Exec(`
		update service_profile
		set
			updated_at = $2,
			name = $3
		where service_profile_id = $1`,
		spID,
		sp.UpdatedAt,
		sp.Name,
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

	n, err := GetNetworkServer(ctx, db, sp.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.UpdateServiceProfile(ctx, &ns.UpdateServiceProfileRequest{
		ServiceProfile: &sp.ServiceProfile,
	})
	if err != nil {
		return errors.Wrap(err, "update service-profile error")
	}

	log.WithFields(log.Fields{
		"id":     spID,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("service-profile updated")

	return nil
}

// DeleteServiceProfile deletes the service-profile matching the given id.
func DeleteServiceProfile(ctx context.Context, db sqlx.Ext, id uuid.UUID) error {
	n, err := GetNetworkServerForServiceProfileID(ctx, db, id)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	res, err := db.Exec("delete from service_profile where service_profile_id = $1", id)
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

	_, err = nsClient.DeleteServiceProfile(ctx, &ns.DeleteServiceProfileRequest{
		Id: id.Bytes(),
	})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return errors.Wrap(err, "delete service-profile error")
	}

	log.WithFields(log.Fields{
		"id":     id,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("service-profile deleted")

	return nil
}

// ServiceProfileFilters provides filters for filtering service profiles.
type ServiceProfileFilters struct {
	UserID          int64 `db:"user_id"`
	OrganizationID  int64 `db:"organization_id"`
	NetworkServerID int64 `db:"network_server_id"`

	// Limit and Offset are added for convenience so that this struct can
	// be given as the arguments.
	Limit  int `db:"limit"`
	Offset int `db:"offset"`
}

// SQL returns the SQL filters.
func (f ServiceProfileFilters) SQL() string {
	var filters []string

	if f.UserID != 0 {
		filters = append(filters, "u.id = :user_id")
	}

	if f.OrganizationID != 0 {
		filters = append(filters, "sp.organization_id = :organization_id")
	}

	if f.NetworkServerID != 0 {
		filters = append(filters, "sp.network_server_id = :network_server_id")
	}

	if len(filters) == 0 {
		return ""
	}

	return "where " + strings.Join(filters, " and ")
}

// GetServiceProfileCount returns the total number of service-profiles.
func GetServiceProfileCount(ctx context.Context, db sqlx.Queryer, filters ServiceProfileFilters) (int, error) {
	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			count(distinct sp.*)
		from
			service_profile sp
		left join organization_user ou
			on sp.organization_id = ou.organization_id
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

// GetServiceProfiles returns a slice of service-profiles.
func GetServiceProfiles(ctx context.Context, db sqlx.Queryer, filters ServiceProfileFilters) ([]ServiceProfileMeta, error) {
	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			sp.*,
			ns.name as network_server_name
		from
			service_profile sp
		inner join network_server ns
			on sp.network_server_id = ns.id
		left join organization_user ou
			on sp.organization_id = ou.organization_id
		left join "user" u
			on ou.user_id = u.id
	`+filters.SQL()+`
		group by
			sp.service_profile_id,
			sp.name,
			network_server_name
		order by
			sp.name
		limit :limit
		offset :offset
	`, filters)
	if err != nil {
		return nil, errors.Wrap(err, "named query error")
	}

	var sps []ServiceProfileMeta
	err = sqlx.Select(db, &sps, query, args...)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return sps, nil
}

// DeleteAllServiceProfilesForOrganizationID deletes all service-profiles
// given an organization id.
func DeleteAllServiceProfilesForOrganizationID(ctx context.Context, db sqlx.Ext, organizationID int64) error {
	var sps []ServiceProfileMeta
	err := sqlx.Select(db, &sps, "select * from service_profile where organization_id = $1", organizationID)
	if err != nil {
		return handlePSQLError(Select, err, "select error")
	}

	for _, sp := range sps {
		err = DeleteServiceProfile(ctx, db, sp.ServiceProfileID)
		if err != nil {
			return errors.Wrap(err, "delete service-profile error")
		}
	}

	return nil
}
