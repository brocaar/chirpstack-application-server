package storage

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
)

// Modulations
const (
	ModulationFSK  = "FSK"
	ModulationLoRa = "LORA"
)

// ExtraChannel defines an extra channel for the gateway-profile.
type ExtraChannel struct {
	Modulation       string
	Frequency        int
	Bandwidth        int
	Bitrate          int
	SpreadingFactors []int
}

// GatewayProfile defines a gateway-profile.
type GatewayProfile struct {
	NetworkServerID int64             `db:"network_server_id"`
	CreatedAt       time.Time         `db:"created_at"`
	UpdatedAt       time.Time         `db:"updated_at"`
	Name            string            `db:"name"`
	GatewayProfile  ns.GatewayProfile `db:"-"`
}

// GatewayProfileMeta defines the gateway-profile meta record.
type GatewayProfileMeta struct {
	GatewayProfileID  uuid.UUID     `db:"gateway_profile_id"`
	NetworkServerID   int64         `db:"network_server_id"`
	NetworkServerName string        `db:"network_server_name"`
	CreatedAt         time.Time     `db:"created_at"`
	UpdatedAt         time.Time     `db:"updated_at"`
	Name              string        `db:"name"`
	StatsInterval     time.Duration `db:"stats_interval"`
}

// CreateGatewayProfile creates the given gateway-profile.
// This will create the gateway-profile at the network-server side and will
// create a local reference record.
func CreateGatewayProfile(ctx context.Context, db sqlx.Ext, gp *GatewayProfile) error {
	gpID, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid v4 error")
	}

	now := time.Now()

	gp.GatewayProfile.Id = gpID.Bytes()
	gp.CreatedAt = now
	gp.UpdatedAt = now

	var statsInterval time.Duration
	if gp.GatewayProfile.StatsInterval != nil {
		statsInterval, err = ptypes.Duration(gp.GatewayProfile.StatsInterval)
		if err != nil {
			return errors.Wrap(err, "stats interval error")
		}
	}

	_, err = db.Exec(`
		insert into gateway_profile (
			gateway_profile_id,
			network_server_id,
			created_at,
			updated_at,
			name,
			stats_interval
		) values ($1, $2, $3, $4, $5, $6)`,

		gpID,
		gp.NetworkServerID,
		gp.CreatedAt,
		gp.UpdatedAt,
		gp.Name,
		statsInterval,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	n, err := GetNetworkServer(ctx, db, gp.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.CreateGatewayProfile(ctx, &ns.CreateGatewayProfileRequest{
		GatewayProfile: &gp.GatewayProfile,
	})
	if err != nil {
		return errors.Wrap(err, "create gateway-profile error")
	}

	log.WithFields(log.Fields{
		"id":     gpID,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("gateway-profile created")

	return nil
}

// GetGatewayProfile returns the gateway-profile matching the given id.
func GetGatewayProfile(ctx context.Context, db sqlx.Queryer, id uuid.UUID) (GatewayProfile, error) {
	var gp GatewayProfile
	err := sqlx.Get(db, &gp, `
		select
			network_server_id,
			name,
			created_at,
			updated_at
		from gateway_profile
		where
			gateway_profile_id = $1`,
		id,
	)
	if err != nil {
		return gp, handlePSQLError(Select, err, "select error")
	}

	n, err := GetNetworkServer(ctx, db, gp.NetworkServerID)
	if err != nil {
		return gp, errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return gp, errors.Wrap(err, "get network-server client error")
	}

	resp, err := nsClient.GetGatewayProfile(ctx, &ns.GetGatewayProfileRequest{
		Id: id.Bytes(),
	})
	if err != nil {
		return gp, errors.Wrap(err, "get gateway-profile error")
	}

	if resp.GatewayProfile == nil {
		return gp, errors.New("gateway_profile must not be nil")
	}

	gp.GatewayProfile = *resp.GatewayProfile

	return gp, nil
}

// UpdateGatewayProfile updates the given gateway-profile.
func UpdateGatewayProfile(ctx context.Context, db sqlx.Ext, gp *GatewayProfile) error {
	gp.UpdatedAt = time.Now()
	gpID, err := uuid.FromBytes(gp.GatewayProfile.Id)
	if err != nil {
		return errors.Wrap(err, "uuid from bytes error")
	}

	var statsInterval time.Duration
	if gp.GatewayProfile.StatsInterval != nil {
		statsInterval, err = ptypes.Duration(gp.GatewayProfile.StatsInterval)
		if err != nil {
			return errors.Wrap(err, "stats interval error")
		}
	}

	res, err := db.Exec(`
		update gateway_profile
		set
			updated_at = $2,
			network_server_id = $3,
			name = $4,
			stats_interval = $5
		where
			gateway_profile_id = $1`,
		gpID,
		gp.UpdatedAt,
		gp.NetworkServerID,
		gp.Name,
		statsInterval,
	)
	if err != nil {
		return handlePSQLError(Update, err, "update gateway-profile error")
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	n, err := GetNetworkServer(ctx, db, gp.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.UpdateGatewayProfile(context.Background(), &ns.UpdateGatewayProfileRequest{
		GatewayProfile: &gp.GatewayProfile,
	})
	if err != nil {
		return errors.Wrap(err, "update gateway-profile error")
	}

	return nil
}

// DeleteGatewayProfile deletes the gateway-profile matching the given id.
func DeleteGatewayProfile(ctx context.Context, db sqlx.Ext, id uuid.UUID) error {
	n, err := GetNetworkServerForGatewayProfileID(ctx, db, id)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	res, err := db.Exec(`
		delete from gateway_profile
		where
			gateway_profile_id = $1`,
		id,
	)
	if err != nil {
		return handlePSQLError(Delete, err, "delete gateway-profile error")
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.DeleteGatewayProfile(ctx, &ns.DeleteGatewayProfileRequest{
		Id: id.Bytes(),
	})
	if err != nil {
		return errors.Wrap(err, "delete gateway-profile error")
	}

	return nil
}

// GetGatewayProfileCount returns the total number of gateway-profiles.
func GetGatewayProfileCount(ctx context.Context, db sqlx.Queryer) (int, error) {
	var count int
	err := sqlx.Get(db, &count, `
		select
			count(*)
		from gateway_profile`)
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}

	return count, nil
}

// GetGatewayProfileCountForNetworkServerID returns the total number of
// gateway-profiles given a network-server ID.
func GetGatewayProfileCountForNetworkServerID(ctx context.Context, db sqlx.Queryer, networkServerID int64) (int, error) {
	var count int
	err := sqlx.Get(db, &count, `
		select
			count(*)
		from gateway_profile
		where
			network_server_id = $1`,
		networkServerID,
	)
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}

	return count, nil
}

// GetGatewayProfiles returns a slice of gateway-profiles.
func GetGatewayProfiles(ctx context.Context, db sqlx.Queryer, limit, offset int) ([]GatewayProfileMeta, error) {
	var gps []GatewayProfileMeta
	err := sqlx.Select(db, &gps, `
		select
			gp.*,
			n.name as network_server_name
		from
			gateway_profile gp
		inner join
			network_server n
		on
			n.id = gp.network_server_id
		order by
			name
		limit $1 offset $2`,
		limit,
		offset,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return gps, nil
}

// GetGatewayProfilesForNetworkServerID returns a slice of gateway-profiles
// for the given network-server ID.
func GetGatewayProfilesForNetworkServerID(ctx context.Context, db sqlx.Queryer, networkServerID int64, limit, offset int) ([]GatewayProfileMeta, error) {
	var gps []GatewayProfileMeta
	err := sqlx.Select(db, &gps, `
		select
			gp.*,
			n.name as network_server_name
		from
			gateway_profile gp
		inner join
			network_server n
		on
			n.id = gp.network_server_id
		where
			network_server_id = $1
		order by
			name
		limit $2 offset $3`,
		networkServerID,
		limit,
		offset,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return gps, nil
}
