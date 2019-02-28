package storage

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/loraserver/api/ns"
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
	ServiceProfileID uuid.UUID `db:"service_profile_id"`
	NetworkServerID  int64     `db:"network_server_id"`
	OrganizationID   int64     `db:"organization_id"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
	Name             string    `db:"name"`
}

// Validate validates the service-profile data.
func (sp ServiceProfile) Validate() error {
	return nil
}

// CreateServiceProfile creates the given service-profile.
func CreateServiceProfile(db sqlx.Ext, sp *ServiceProfile) error {
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

	n, err := GetNetworkServer(db, sp.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.CreateServiceProfile(context.Background(), &ns.CreateServiceProfileRequest{
		ServiceProfile: &sp.ServiceProfile,
	})
	if err != nil {
		return handleGrpcError(err, "create service-profile error")
	}

	log.WithField("id", spID).Info("service-profile created")
	return nil
}

// GetServiceProfile returns the service-profile matching the given id.
func GetServiceProfile(db sqlx.Queryer, id uuid.UUID, localOnly bool) (ServiceProfile, error) {
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

	n, err := GetNetworkServer(db, sp.NetworkServerID)
	if err != nil {
		return sp, errors.Wrap(err, "get network-server errror")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return sp, errors.Wrap(err, "get network-server client error")
	}

	resp, err := nsClient.GetServiceProfile(context.Background(), &ns.GetServiceProfileRequest{
		Id: id.Bytes(),
	})
	if err != nil {
		return sp, handleGrpcError(err, "get service-profile error")
	}

	if resp.ServiceProfile == nil {
		return sp, errors.New("service_profile must not be nil")
	}

	sp.ServiceProfile = *resp.ServiceProfile

	return sp, nil
}

// UpdateServiceProfile updates the given service-profile.
func UpdateServiceProfile(db sqlx.Ext, sp *ServiceProfile) error {
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

	n, err := GetNetworkServer(db, sp.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.UpdateServiceProfile(context.Background(), &ns.UpdateServiceProfileRequest{
		ServiceProfile: &sp.ServiceProfile,
	})
	if err != nil {
		return handleGrpcError(err, "update service-profile error")
	}

	log.WithField("id", spID).Info("service-profile updated")

	return nil
}

// DeleteServiceProfile deletes the service-profile matching the given id.
func DeleteServiceProfile(db sqlx.Ext, id uuid.UUID) error {
	n, err := GetNetworkServerForServiceProfileID(db, id)
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

	_, err = nsClient.DeleteServiceProfile(context.Background(), &ns.DeleteServiceProfileRequest{
		Id: id.Bytes(),
	})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return handleGrpcError(err, "delete service-profile error")
	}

	log.WithField("id", id).Info("service-profile deleted")

	return nil
}

// GetServiceProfileCount returns the total number of service-profiles.
func GetServiceProfileCount(db sqlx.Queryer) (int, error) {
	var count int
	err := sqlx.Get(db, &count, "select count(*) from service_profile")
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}
	return count, nil
}

// GetServiceProfileCountForOrganizationID returns the total number of
// service-profiles for the given organization id.
func GetServiceProfileCountForOrganizationID(db sqlx.Queryer, organizationID int64) (int, error) {
	var count int
	err := sqlx.Get(db, &count, "select count(*) from service_profile where organization_id = $1", organizationID)
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}
	return count, nil
}

// GetServiceProfileCountForUser returns the total number of service-profiles
// for the given username.
func GetServiceProfileCountForUser(db sqlx.Queryer, username string) (int, error) {
	var count int
	err := sqlx.Get(db, &count, `
		select
			count(sp.*)
		from service_profile sp
		inner join organization o
			on o.id = sp.organization_id
		inner join organization_user ou
			on ou.organization_id = o.id
		inner join "user" u
			on u.id = ou.user_id
		where
			u.username = $1`,
		username,
	)
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}
	return count, nil
}

// GetServiceProfiles returns a slice of service-profiles.
func GetServiceProfiles(db sqlx.Queryer, limit, offset int) ([]ServiceProfileMeta, error) {
	var sps []ServiceProfileMeta
	err := sqlx.Select(db, &sps, `
		select *
		from service_profile
		order by name
		limit $1 offset $2`,
		limit,
		offset,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return sps, nil
}

// GetServiceProfilesForOrganizationID returns a slice of service-profiles
// for the given organization id.
func GetServiceProfilesForOrganizationID(db sqlx.Queryer, organizationID int64, limit, offset int) ([]ServiceProfileMeta, error) {
	var sps []ServiceProfileMeta
	err := sqlx.Select(db, &sps, `
		select *
		from service_profile
		where
			organization_id = $1
		order by name
		limit $2 offset $3`,
		organizationID,
		limit,
		offset,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return sps, nil
}

// GetServiceProfilesForUser returns a slice of service-profile for the given
// username.
func GetServiceProfilesForUser(db sqlx.Queryer, username string, limit, offset int) ([]ServiceProfileMeta, error) {
	var sps []ServiceProfileMeta
	err := sqlx.Select(db, &sps, `
		select
			sp.*
		from service_profile sp
		inner join organization o
			on o.id = sp.organization_id
		inner join organization_user ou
			on ou.organization_id = o.id
		inner join "user" u
			on u.id = ou.user_id
		where
			u.username = $1
		order by sp.name
		limit $2 offset $3`,
		username,
		limit,
		offset,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return sps, nil
}

// DeleteAllServiceProfilesForOrganizationID deletes all service-profiles
// given an organization id.
func DeleteAllServiceProfilesForOrganizationID(db sqlx.Ext, organizationID int64) error {
	var sps []ServiceProfileMeta
	err := sqlx.Select(db, &sps, "select * from service_profile where organization_id = $1", organizationID)
	if err != nil {
		return handlePSQLError(Select, err, "select error")
	}

	for _, sp := range sps {
		err = DeleteServiceProfile(db, sp.ServiceProfileID)
		if err != nil {
			return errors.Wrap(err, "delete service-profile error")
		}
	}

	return nil
}
