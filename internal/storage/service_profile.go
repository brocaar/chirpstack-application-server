package storage

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/common"
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

// Validate validates the service-profile data.
func (sp ServiceProfile) Validate() error {
	return nil
}

// CreateServiceProfile creates the given service-profile.
func CreateServiceProfile(db sqlx.Execer, sp *ServiceProfile) error {
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
		return handlePSQLError(err, "insert error")
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

	_, err = common.NetworkServer.CreateServiceProfile(context.Background(), &req)
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
		return sp, handlePSQLError(err, "select error")
	}

	err := row.Scan(&sp.ServiceProfile.ServiceProfileID, &sp.NetworkServerID, &sp.OrganizationID, &sp.CreatedAt, &sp.UpdatedAt, &sp.Name)
	if err != nil {
		return sp, handlePSQLError(err, "scan error")
	}

	resp, err := common.NetworkServer.GetServiceProfile(context.Background(), &ns.GetServiceProfileRequest{
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
func UpdateServiceProfile(db sqlx.Execer, sp *ServiceProfile) error {
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
		return handlePSQLError(err, "update error")
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

	_, err = common.NetworkServer.UpdateServiceProfile(context.Background(), &req)
	if err != nil {
		return handleGrpcError(err, "update service-profile error")
	}

	log.WithField("service_profile_id", sp.ServiceProfile.ServiceProfileID).Info("service-profile updated")

	return nil
}

// DeleteServiceProfile deletes the service-profile matching the given id.
func DeleteServiceProfile(db sqlx.Execer, id string) error {
	res, err := db.Exec("delete from service_profile where service_profile_id = $1", id)
	if err != nil {
		return handlePSQLError(err, "delete error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	_, err = common.NetworkServer.DeleteServiceProfile(context.Background(), &ns.DeleteServiceProfileRequest{
		ServiceProfileID: id,
	})
	if err != nil {
		return handleGrpcError(err, "delete service-profile error")
	}

	log.WithField("service_profile_id", id).Info("service-profile deleted")

	return nil
}
