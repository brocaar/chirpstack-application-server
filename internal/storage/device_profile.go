package storage

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/loraserver/api/ns"
)

// DeviceProfile defines the device-profile.
type DeviceProfile struct {
	NetworkServerID      int64            `db:"network_server_id"`
	OrganizationID       int64            `db:"organization_id"`
	CreatedAt            time.Time        `db:"created_at"`
	UpdatedAt            time.Time        `db:"updated_at"`
	Name                 string           `db:"name"`
	PayloadCodec         codec.Type       `db:"payload_codec"`
	PayloadEncoderScript string           `db:"payload_encoder_script"`
	PayloadDecoderScript string           `db:"payload_decoder_script"`
	DeviceProfile        ns.DeviceProfile `db:"-"`
}

// DeviceProfileMeta defines the device-profile meta record.
type DeviceProfileMeta struct {
	DeviceProfileID uuid.UUID `db:"device_profile_id"`
	NetworkServerID int64     `db:"network_server_id"`
	OrganizationID  int64     `db:"organization_id"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
	Name            string    `db:"name"`
}

// Validate validates the device-profile data.
func (dp DeviceProfile) Validate() error {
	if dp.Name == "" {
		return ErrDeviceProfileInvalidName
	}
	return nil
}

// CreateDeviceProfile creates the given device-profile.
// This will create the device-profile at the network-server side and will
// create a local reference record.
func CreateDeviceProfile(ctx context.Context, db sqlx.Ext, dp *DeviceProfile) error {
	if err := dp.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	dpID, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid v4 error")
	}

	now := time.Now()
	dp.DeviceProfile.Id = dpID.Bytes()
	dp.CreatedAt = now
	dp.UpdatedAt = now

	_, err = db.Exec(`
        insert into device_profile (
            device_profile_id,
            network_server_id,
            organization_id,
            created_at,
            updated_at,
            name,
			payload_codec,
			payload_encoder_script,
			payload_decoder_script
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		dpID,
		dp.NetworkServerID,
		dp.OrganizationID,
		dp.CreatedAt,
		dp.UpdatedAt,
		dp.Name,
		dp.PayloadCodec,
		dp.PayloadEncoderScript,
		dp.PayloadDecoderScript,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	n, err := GetNetworkServer(ctx, db, dp.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.CreateDeviceProfile(ctx, &ns.CreateDeviceProfileRequest{
		DeviceProfile: &dp.DeviceProfile,
	})
	if err != nil {
		return errors.Wrap(err, "create device-profile errror")
	}

	log.WithFields(log.Fields{
		"id":     dpID,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("device-profile created")

	return nil
}

// GetDeviceProfile returns the device-profile matching the given id.
// When forUpdate is set to true, then db must be a db transaction.
// When localOnly is set to true, no call to the network-server is made to
// retrieve additional device data.
func GetDeviceProfile(ctx context.Context, db sqlx.Queryer, id uuid.UUID, forUpdate, localOnly bool) (DeviceProfile, error) {
	var fu string
	if forUpdate {
		fu = " for update"
	}

	var dp DeviceProfile

	row := db.QueryRowx(`
		select
			network_server_id,
			organization_id,
			created_at,
			updated_at,
			name,
			payload_codec,
			payload_encoder_script,
			payload_decoder_script
		from device_profile
		where
			device_profile_id = $1`+fu,
		id,
	)
	if err := row.Err(); err != nil {
		return dp, handlePSQLError(Select, err, "select error")
	}

	err := row.Scan(
		&dp.NetworkServerID,
		&dp.OrganizationID,
		&dp.CreatedAt,
		&dp.UpdatedAt,
		&dp.Name,
		&dp.PayloadCodec,
		&dp.PayloadEncoderScript,
		&dp.PayloadDecoderScript,
	)
	if err != nil {
		return dp, handlePSQLError(Scan, err, "scan error")
	}

	if localOnly {
		return dp, nil
	}

	n, err := GetNetworkServer(ctx, db, dp.NetworkServerID)
	if err != nil {
		return dp, errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return dp, errors.Wrap(err, "get network-server client error")
	}

	resp, err := nsClient.GetDeviceProfile(ctx, &ns.GetDeviceProfileRequest{
		Id: id.Bytes(),
	})
	if err != nil {
		return dp, errors.Wrap(err, "get device-profile error")
	}
	if resp.DeviceProfile == nil {
		return dp, errors.New("device_profile must not be nil")
	}

	dp.DeviceProfile = *resp.DeviceProfile

	return dp, nil
}

// UpdateDeviceProfile updates the given device-profile.
func UpdateDeviceProfile(ctx context.Context, db sqlx.Ext, dp *DeviceProfile) error {
	if err := dp.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
	if err != nil {
		return errors.Wrap(err, "uuid from bytes error")
	}

	n, err := GetNetworkServer(ctx, db, dp.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.UpdateDeviceProfile(ctx, &ns.UpdateDeviceProfileRequest{
		DeviceProfile: &dp.DeviceProfile,
	})
	if err != nil {
		return errors.Wrap(err, "update device-profile error")
	}

	dp.UpdatedAt = time.Now()

	res, err := db.Exec(`
        update device_profile
        set
            updated_at = $2,
            name = $3,
			payload_codec = $4,
			payload_encoder_script = $5,
			payload_decoder_script = $6
		where device_profile_id = $1`,
		dpID,
		dp.UpdatedAt,
		dp.Name,
		dp.PayloadCodec,
		dp.PayloadEncoderScript,
		dp.PayloadDecoderScript,
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
		"id":     dpID,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("device-profile updated")

	return nil
}

// DeleteDeviceProfile deletes the device-profile matching the given id.
func DeleteDeviceProfile(ctx context.Context, db sqlx.Ext, id uuid.UUID) error {
	n, err := GetNetworkServerForDeviceProfileID(ctx, db, id)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	res, err := db.Exec("delete from device_profile where device_profile_id = $1", id)
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

	_, err = nsClient.DeleteDeviceProfile(ctx, &ns.DeleteDeviceProfileRequest{
		Id: id.Bytes(),
	})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return errors.Wrap(err, "delete device-profile error")
	}

	log.WithFields(log.Fields{
		"id":     id,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("device-profile deleted")

	return nil
}

// GetDeviceProfileCount returns the total number of device-profiles.
func GetDeviceProfileCount(ctx context.Context, db sqlx.Queryer) (int, error) {
	var count int
	err := sqlx.Get(db, &count, "select count(*) from device_profile")
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}

	return count, nil
}

// GetDeviceProfileCountForOrganizationID returns the total number of
// device-profiles for the given organization id.
func GetDeviceProfileCountForOrganizationID(ctx context.Context, db sqlx.Queryer, organizationID int64) (int, error) {
	var count int
	err := sqlx.Get(db, &count, "select count(*) from device_profile where organization_id = $1", organizationID)
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}
	return count, nil
}

// GetDeviceProfileCountForUser returns the total number of device-profiles
// for the given username.
func GetDeviceProfileCountForUser(ctx context.Context, db sqlx.Queryer, username string) (int, error) {
	var count int
	err := sqlx.Get(db, &count, `
		select
			count(dp.*)
		from device_profile dp
		inner join organization o
			on o.id = dp.organization_id
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

// GetDeviceProfileCountForApplicationID returns the total number of
// device-profiles that can be used for the given application id (based
// on the service-profile of the application).
func GetDeviceProfileCountForApplicationID(ctx context.Context, db sqlx.Queryer, applicationID int64) (int, error) {
	var count int
	err := sqlx.Get(db, &count, `
		select
			count(dp.*)
		from device_profile dp
		inner join network_server ns
			on ns.id = dp.network_server_id
		inner join service_profile sp
			on sp.network_server_id = ns.id
		inner join application a
			on a.service_profile_id = sp.service_profile_id
		where
			a.id = $1
			and dp.organization_id = a.organization_id`,
		applicationID,
	)
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}
	return count, nil
}

// GetDeviceProfiles returns a slice of device-profiles.
func GetDeviceProfiles(ctx context.Context, db sqlx.Queryer, limit, offset int) ([]DeviceProfileMeta, error) {
	var dps []DeviceProfileMeta
	err := sqlx.Select(db, &dps, `
		select
			device_profile_id,
			network_server_id,
			organization_id,
			created_at,
			updated_at,
			name
		from device_profile
		order by name
		limit $1 offset $2`,
		limit,
		offset,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}
	return dps, nil
}

// GetDeviceProfilesForOrganizationID returns a slice of device-profiles
// for the given organization id.
func GetDeviceProfilesForOrganizationID(ctx context.Context, db sqlx.Queryer, organizationID int64, limit, offset int) ([]DeviceProfileMeta, error) {
	var dps []DeviceProfileMeta
	err := sqlx.Select(db, &dps, `
		select
			device_profile_id,
			network_server_id,
			organization_id,
			created_at,
			updated_at,
			name
		from device_profile
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
	return dps, nil
}

// GetDeviceProfilesForUser returns a slice of device-profiles for the given
// username.
func GetDeviceProfilesForUser(ctx context.Context, db sqlx.Queryer, username string, limit, offset int) ([]DeviceProfileMeta, error) {
	var dps []DeviceProfileMeta
	err := sqlx.Select(db, &dps, `
		select
			dp.device_profile_id,
			dp.network_server_id,
			dp.organization_id,
			dp.created_at,
			dp.updated_at,
			dp.name
		from
			device_profile dp
		inner join organization o
			on o.id = dp.organization_id
		inner join organization_user ou
			on ou.organization_id = o.id
		inner join "user" u
			on u.id = ou.user_id
		where
			u.username = $1
		order by dp.name
		limit $2 offset $3`,
		username,
		limit,
		offset,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return dps, nil
}

// GetDeviceProfilesForApplicationID returns a slice of device-profiles that
// can be used for the given application id (based on the service-profile
// of the application).
func GetDeviceProfilesForApplicationID(ctx context.Context, db sqlx.Queryer, applicationID int64, limit, offset int) ([]DeviceProfileMeta, error) {
	var dps []DeviceProfileMeta
	err := sqlx.Select(db, &dps, `
		select
			dp.device_profile_id,
			dp.network_server_id,
			dp.organization_id,
			dp.created_at,
			dp.updated_at,
			dp.name
		from
			device_profile dp
		inner join network_server ns
			on ns.id = dp.network_server_id
		inner join service_profile sp
			on sp.network_server_id = ns.id
		inner join application a
			on a.service_profile_id = sp.service_profile_id
		where
			a.id = $1
			and dp.organization_id = a.organization_id
		order by dp.name
		limit $2 offset $3`,
		applicationID,
		limit,
		offset,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}
	return dps, nil
}

// DeleteAllDeviceProfilesForOrganizationID deletes all device-profiles
// given an organization id.
func DeleteAllDeviceProfilesForOrganizationID(ctx context.Context, db sqlx.Ext, organizationID int64) error {
	var dps []DeviceProfileMeta
	err := sqlx.Select(db, &dps, `
		select
			device_profile_id,
			network_server_id,
			organization_id,
			created_at,
			updated_at,
			name
		from
			device_profile
		where
			organization_id = $1`,
		organizationID)
	if err != nil {
		return handlePSQLError(Select, err, "select error")
	}

	for _, dp := range dps {
		err = DeleteDeviceProfile(ctx, db, dp.DeviceProfileID)
		if err != nil {
			return errors.Wrap(err, "delete device-profile error")
		}
	}

	return nil
}
