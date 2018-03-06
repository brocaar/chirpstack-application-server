package storage

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan/backend"
)

// ServiceProfile defines the service-profile.
type ServiceProfile struct {
	NetworkServerID int64                  `db:"network_server_id"`
	OrganizationID  int64                  `db:"organization_id"`
	CreatedAt       time.Time              `db:"created_at"`
	UpdatedAt       time.Time              `db:"updated_at"`
	Name            string                 `db:"name"`
	ServiceProfile  backend.ServiceProfile `db:"-"`
}

// ServiceProfileMeta defines the service-profile meta record.
type ServiceProfileMeta struct {
	ServiceProfileID string    `db:"service_profile_id"`
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

	now := time.Now()
	sp.CreatedAt = now
	sp.UpdatedAt = now
	sp.ServiceProfile.ServiceProfileID = uuid.NewV4().String()

	_, err := db.Exec(`
		insert into service_profile (
			service_profile_id,
			network_server_id,
			organization_id,
			created_at,
			updated_at,
			name
		) values ($1, $2, $3, $4, $5, $6)`,
		sp.ServiceProfile.ServiceProfileID,
		sp.NetworkServerID,
		sp.OrganizationID,
		sp.CreatedAt,
		sp.UpdatedAt,
		sp.Name,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	req := ns.CreateServiceProfileRequest{
		ServiceProfile: &ns.ServiceProfile{
			ServiceProfileID:       sp.ServiceProfile.ServiceProfileID,
			UlRate:                 uint32(sp.ServiceProfile.ULRate),
			UlBucketSize:           uint32(sp.ServiceProfile.ULBucketSize),
			DlRate:                 uint32(sp.ServiceProfile.DLRate),
			DlBucketSize:           uint32(sp.ServiceProfile.DLBucketSize),
			AddGWMetadata:          sp.ServiceProfile.AddGWMetadata,
			DevStatusReqFreq:       uint32(sp.ServiceProfile.DevStatusReqFreq),
			ReportDevStatusBattery: sp.ServiceProfile.ReportDevStatusBattery,
			ReportDevStatusMargin:  sp.ServiceProfile.ReportDevStatusMargin,
			DrMin:          uint32(sp.ServiceProfile.DRMin),
			DrMax:          uint32(sp.ServiceProfile.DRMax),
			ChannelMask:    []byte(sp.ServiceProfile.ChannelMask),
			PrAllowed:      sp.ServiceProfile.PRAllowed,
			HrAllowed:      sp.ServiceProfile.HRAllowed,
			RaAllowed:      sp.ServiceProfile.RAAllowed,
			NwkGeoLoc:      sp.ServiceProfile.NwkGeoLoc,
			TargetPER:      uint32(sp.ServiceProfile.TargetPER),
			MinGWDiversity: uint32(sp.ServiceProfile.MinGWDiversity),
		},
	}

	switch sp.ServiceProfile.ULRatePolicy {
	case backend.Drop:
		req.ServiceProfile.UlRatePolicy = ns.RatePolicy_DROP
	case backend.Mark:
		req.ServiceProfile.UlRatePolicy = ns.RatePolicy_MARK
	}

	switch sp.ServiceProfile.DLRatePolicy {
	case backend.Drop:
		req.ServiceProfile.DlRatePolicy = ns.RatePolicy_DROP
	case backend.Mark:
		req.ServiceProfile.DlRatePolicy = ns.RatePolicy_MARK
	}

	n, err := GetNetworkServer(db, sp.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.CreateServiceProfile(context.Background(), &req)
	if err != nil {
		return handleGrpcError(err, "create service-profile error")
	}

	log.WithField("service_profile_id", sp.ServiceProfile.ServiceProfileID).Info("service-profile created")
	return nil
}

// GetServiceProfile returns the service-profile matching the given id.
func GetServiceProfile(db sqlx.Queryer, id string) (ServiceProfile, error) {
	var sp ServiceProfile
	row := db.QueryRowx(`
		select 
			service_profile_id,
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

	err := row.Scan(&sp.ServiceProfile.ServiceProfileID, &sp.NetworkServerID, &sp.OrganizationID, &sp.CreatedAt, &sp.UpdatedAt, &sp.Name)
	if err != nil {
		return sp, handlePSQLError(Scan, err, "scan error")
	}

	n, err := GetNetworkServer(db, sp.NetworkServerID)
	if err != nil {
		return sp, errors.Wrap(err, "get network-server errror")
	}

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return sp, errors.Wrap(err, "get network-server client error")
	}

	resp, err := nsClient.GetServiceProfile(context.Background(), &ns.GetServiceProfileRequest{
		ServiceProfileID: id,
	})
	if err != nil {
		return sp, handleGrpcError(err, "get service-profile error")
	}

	sp.ServiceProfile = backend.ServiceProfile{
		ServiceProfileID:       resp.ServiceProfile.ServiceProfileID,
		ULRate:                 int(resp.ServiceProfile.UlRate),
		ULBucketSize:           int(resp.ServiceProfile.UlBucketSize),
		DLRate:                 int(resp.ServiceProfile.DlRate),
		DLBucketSize:           int(resp.ServiceProfile.DlBucketSize),
		AddGWMetadata:          resp.ServiceProfile.AddGWMetadata,
		DevStatusReqFreq:       int(resp.ServiceProfile.DevStatusReqFreq),
		ReportDevStatusBattery: resp.ServiceProfile.ReportDevStatusBattery,
		ReportDevStatusMargin:  resp.ServiceProfile.ReportDevStatusMargin,
		DRMin:          int(resp.ServiceProfile.DrMin),
		DRMax:          int(resp.ServiceProfile.DrMax),
		ChannelMask:    backend.HEXBytes(resp.ServiceProfile.ChannelMask),
		PRAllowed:      resp.ServiceProfile.PrAllowed,
		HRAllowed:      resp.ServiceProfile.HrAllowed,
		RAAllowed:      resp.ServiceProfile.RaAllowed,
		NwkGeoLoc:      resp.ServiceProfile.NwkGeoLoc,
		TargetPER:      backend.Percentage(resp.ServiceProfile.TargetPER),
		MinGWDiversity: int(resp.ServiceProfile.MinGWDiversity),
	}

	switch resp.ServiceProfile.UlRatePolicy {
	case ns.RatePolicy_MARK:
		sp.ServiceProfile.ULRatePolicy = backend.Mark
	case ns.RatePolicy_DROP:
		sp.ServiceProfile.ULRatePolicy = backend.Drop
	}

	switch resp.ServiceProfile.DlRatePolicy {
	case ns.RatePolicy_MARK:
		sp.ServiceProfile.DLRatePolicy = backend.Mark
	case ns.RatePolicy_DROP:
		sp.ServiceProfile.DLRatePolicy = backend.Drop
	}

	return sp, nil
}

// UpdateServiceProfile updates the given service-profile.
func UpdateServiceProfile(db sqlx.Ext, sp *ServiceProfile) error {
	if err := sp.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	sp.UpdatedAt = time.Now()
	res, err := db.Exec(`
		update service_profile
		set
			updated_at = $2,
			name = $3
		where service_profile_id = $1`,
		sp.ServiceProfile.ServiceProfileID,
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

	req := ns.UpdateServiceProfileRequest{
		ServiceProfile: &ns.ServiceProfile{
			ServiceProfileID:       sp.ServiceProfile.ServiceProfileID,
			UlRate:                 uint32(sp.ServiceProfile.ULRate),
			UlBucketSize:           uint32(sp.ServiceProfile.ULBucketSize),
			DlRate:                 uint32(sp.ServiceProfile.DLRate),
			DlBucketSize:           uint32(sp.ServiceProfile.DLBucketSize),
			AddGWMetadata:          sp.ServiceProfile.AddGWMetadata,
			DevStatusReqFreq:       uint32(sp.ServiceProfile.DevStatusReqFreq),
			ReportDevStatusBattery: sp.ServiceProfile.ReportDevStatusBattery,
			ReportDevStatusMargin:  sp.ServiceProfile.ReportDevStatusMargin,
			DrMin:          uint32(sp.ServiceProfile.DRMin),
			DrMax:          uint32(sp.ServiceProfile.DRMax),
			ChannelMask:    []byte(sp.ServiceProfile.ChannelMask),
			PrAllowed:      sp.ServiceProfile.PRAllowed,
			HrAllowed:      sp.ServiceProfile.HRAllowed,
			RaAllowed:      sp.ServiceProfile.RAAllowed,
			NwkGeoLoc:      sp.ServiceProfile.NwkGeoLoc,
			TargetPER:      uint32(sp.ServiceProfile.TargetPER),
			MinGWDiversity: uint32(sp.ServiceProfile.MinGWDiversity),
		},
	}

	switch sp.ServiceProfile.ULRatePolicy {
	case backend.Drop:
		req.ServiceProfile.UlRatePolicy = ns.RatePolicy_DROP
	case backend.Mark:
		req.ServiceProfile.UlRatePolicy = ns.RatePolicy_MARK
	}

	switch sp.ServiceProfile.DLRatePolicy {
	case backend.Drop:
		req.ServiceProfile.DlRatePolicy = ns.RatePolicy_DROP
	case backend.Mark:
		req.ServiceProfile.DlRatePolicy = ns.RatePolicy_MARK
	}

	n, err := GetNetworkServer(db, sp.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.UpdateServiceProfile(context.Background(), &req)
	if err != nil {
		return handleGrpcError(err, "update service-profile error")
	}

	log.WithField("service_profile_id", sp.ServiceProfile.ServiceProfileID).Info("service-profile updated")

	return nil
}

// DeleteServiceProfile deletes the service-profile matching the given id.
func DeleteServiceProfile(db sqlx.Ext, id string) error {
	n, err := GetNetworkServerForServiceProfileID(db, id)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
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
		ServiceProfileID: id,
	})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return handleGrpcError(err, "delete service-profile error")
	}

	log.WithField("service_profile_id", id).Info("service-profile deleted")

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
