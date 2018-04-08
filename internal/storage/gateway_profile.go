package storage

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/loraserver/api/ns"
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
	GatewayProfileID string         `db:"gateway_profile_id"`
	NetworkServerID  int64          `db:"network_server_id"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
	Name             string         `db:"name"`
	Channels         []int          `db:"-"`
	ExtraChannels    []ExtraChannel `db:"-"`
}

// GatewayProfileMeta defines the gateway-profile meta record.
type GatewayProfileMeta struct {
	GatewayProfileID string    `db:"gateway_profile_id"`
	NetworkServerID  int64     `db:"network_server_id"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
	Name             string    `db:"name"`
}

// CreateGatewayProfile creates the given gateway-profile.
// This will create the gateway-profile at the network-server side and will
// create a local reference record.
func CreateGatewayProfile(db sqlx.Ext, gp *GatewayProfile) error {
	now := time.Now()
	gp.GatewayProfileID = uuid.NewV4().String()
	gp.CreatedAt = now
	gp.UpdatedAt = now

	_, err := db.Exec(`
		insert into gateway_profile (
			gateway_profile_id,
			network_server_id,
			created_at,
			updated_at,
			name
		) values ($1, $2, $3, $4, $5)`,

		gp.GatewayProfileID,
		gp.NetworkServerID,
		gp.CreatedAt,
		gp.UpdatedAt,
		gp.Name,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	req := ns.CreateGatewayProfileRequest{
		GatewayProfile: &ns.GatewayProfile{
			GatewayProfileID: gp.GatewayProfileID,
		},
	}

	for _, c := range gp.Channels {
		req.GatewayProfile.Channels = append(req.GatewayProfile.Channels, uint32(c))
	}

	for _, ec := range gp.ExtraChannels {
		c := ns.GatewayProfileExtraChannel{
			Frequency: uint32(ec.Frequency),
			Bandwidth: uint32(ec.Bandwidth),
			Bitrate:   uint32(ec.Bitrate),
		}

		switch ec.Modulation {
		case ModulationFSK:
			c.Modulation = ns.Modulation_FSK
		default:
			c.Modulation = ns.Modulation_LORA
		}

		for _, sf := range ec.SpreadingFactors {
			c.SpreadingFactors = append(c.SpreadingFactors, uint32(sf))
		}

		req.GatewayProfile.ExtraChannels = append(req.GatewayProfile.ExtraChannels, &c)
	}

	n, err := GetNetworkServer(db, gp.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.CreateGatewayProfile(context.Background(), &req)
	if err != nil {
		return handleGrpcError(err, "create gateway-profile error")
	}

	log.WithFields(log.Fields{
		"gateway_profile_id": gp.GatewayProfileID,
	}).Info("gateway-profile created")

	return nil
}

// GetGatewayProfile returns the gateway-profile matching the given id.
func GetGatewayProfile(db sqlx.Queryer, id string) (GatewayProfile, error) {
	var gp GatewayProfile
	err := sqlx.Get(db, &gp, `
		select *
		from gateway_profile
		where
			gateway_profile_id = $1`,
		id,
	)
	if err != nil {
		return gp, handlePSQLError(Select, err, "select error")
	}

	n, err := GetNetworkServer(db, gp.NetworkServerID)
	if err != nil {
		return gp, errors.Wrap(err, "get network-server error")
	}

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return gp, errors.Wrap(err, "get network-server client error")
	}

	resp, err := nsClient.GetGatewayProfile(context.Background(), &ns.GetGatewayProfileRequest{
		GatewayProfileID: id,
	})
	if err != nil {
		return gp, handleGrpcError(err, "get gateway-profile error")
	}

	for _, c := range resp.GatewayProfile.Channels {
		gp.Channels = append(gp.Channels, int(c))
	}

	for _, ec := range resp.GatewayProfile.ExtraChannels {
		c := ExtraChannel{
			Frequency: int(ec.Frequency),
			Bandwidth: int(ec.Bandwidth),
			Bitrate:   int(ec.Bitrate),
		}

		switch ec.Modulation {
		case ns.Modulation_FSK:
			c.Modulation = ModulationFSK
		default:
			c.Modulation = ModulationLoRa
		}

		for _, sf := range ec.SpreadingFactors {
			c.SpreadingFactors = append(c.SpreadingFactors, int(sf))
		}

		gp.ExtraChannels = append(gp.ExtraChannels, c)
	}

	return gp, nil
}

// UpdateGatewayProfile updates the given gateway-profile.
func UpdateGatewayProfile(db sqlx.Ext, gp *GatewayProfile) error {
	gp.UpdatedAt = time.Now()

	res, err := db.Exec(`
		update gateway_profile
		set
			updated_at = $2,
			network_server_id = $3,
			name = $4
		where
			gateway_profile_id = $1`,
		gp.GatewayProfileID,
		gp.UpdatedAt,
		gp.NetworkServerID,
		gp.Name,
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

	n, err := GetNetworkServer(db, gp.NetworkServerID)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	req := ns.UpdateGatewayProfileRequest{
		GatewayProfile: &ns.GatewayProfile{
			GatewayProfileID: gp.GatewayProfileID,
		},
	}

	for _, c := range gp.Channels {
		req.GatewayProfile.Channels = append(req.GatewayProfile.Channels, uint32(c))
	}

	for _, ec := range gp.ExtraChannels {
		c := ns.GatewayProfileExtraChannel{
			Frequency: uint32(ec.Frequency),
			Bandwidth: uint32(ec.Bandwidth),
			Bitrate:   uint32(ec.Bitrate),
		}

		switch ec.Modulation {
		case ModulationFSK:
			c.Modulation = ns.Modulation_FSK
		default:
			c.Modulation = ns.Modulation_LORA
		}

		for _, sf := range ec.SpreadingFactors {
			c.SpreadingFactors = append(c.SpreadingFactors, uint32(sf))
		}

		req.GatewayProfile.ExtraChannels = append(req.GatewayProfile.ExtraChannels, &c)
	}

	_, err = nsClient.UpdateGatewayProfile(context.Background(), &req)
	if err != nil {
		handleGrpcError(err, "update gateway-profile error")
	}

	return nil
}

// DeleteGatewayProfile deletes the gateway-profile matching the given id.
func DeleteGatewayProfile(db sqlx.Ext, id string) error {
	n, err := GetNetworkServerForGatewayProfileID(db, id)
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

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.DeleteGatewayProfile(context.Background(), &ns.DeleteGatewayProfileRequest{
		GatewayProfileID: id,
	})
	if err != nil {
		return handleGrpcError(err, "delete gateway-profile error")
	}

	return nil
}

// GetGatewayProfileCount returns the total number of gateway-profiles.
func GetGatewayProfileCount(db sqlx.Queryer) (int, error) {
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
func GetGatewayProfileCountForNetworkServerID(db sqlx.Queryer, networkServerID int64) (int, error) {
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
func GetGatewayProfiles(db sqlx.Queryer, limit, offset int) ([]GatewayProfileMeta, error) {
	var gps []GatewayProfileMeta
	err := sqlx.Select(db, &gps, `
		select
			*
		from gateway_profile
		order by name
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
func GetGatewayProfilesForNetworkServerID(db sqlx.Queryer, networkServerID int64, limit, offset int) ([]GatewayProfileMeta, error) {
	var gps []GatewayProfileMeta
	err := sqlx.Select(db, &gps, `
		select
			*
		from gateway_profile
		where
			network_server_id = $1
		order by name
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
