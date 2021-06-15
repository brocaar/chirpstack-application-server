package storage

import (
	"context"
	"strings"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

// NetworkServer defines the information to connect to a network-server.
type NetworkServer struct {
	ID                          int64     `db:"id"`
	CreatedAt                   time.Time `db:"created_at"`
	UpdatedAt                   time.Time `db:"updated_at"`
	Name                        string    `db:"name"`
	Server                      string    `db:"server"`
	CACert                      string    `db:"ca_cert"`
	TLSCert                     string    `db:"tls_cert"`
	TLSKey                      string    `db:"tls_key"`
	RoutingProfileCACert        string    `db:"routing_profile_ca_cert"`
	RoutingProfileTLSCert       string    `db:"routing_profile_tls_cert"`
	RoutingProfileTLSKey        string    `db:"routing_profile_tls_key"`
	GatewayDiscoveryEnabled     bool      `db:"gateway_discovery_enabled"`
	GatewayDiscoveryInterval    int       `db:"gateway_discovery_interval"`
	GatewayDiscoveryTXFrequency int       `db:"gateway_discovery_tx_frequency"`
	GatewayDiscoveryDR          int       `db:"gateway_discovery_dr"`
}

// Validate validates the network-server data.
func (ns NetworkServer) Validate() error {
	if strings.TrimSpace(ns.Name) == "" || len(ns.Name) > 100 {
		return ErrNetworkServerInvalidName
	}

	if ns.GatewayDiscoveryEnabled && ns.GatewayDiscoveryInterval <= 0 {
		return ErrInvalidGatewayDiscoveryInterval
	}
	return nil
}

// CreateNetworkServer creates the given network-server.
func CreateNetworkServer(ctx context.Context, db sqlx.Queryer, n *NetworkServer) error {
	if err := n.Validate(); err != nil {
		return errors.Wrap(err, "validation error")
	}

	now := time.Now()
	n.CreatedAt = now
	n.UpdatedAt = now

	err := sqlx.Get(db, &n.ID, `
		insert into network_server (
			created_at,
			updated_at,
			name,
			server,
			ca_cert,
			tls_cert,
			tls_key,
			routing_profile_ca_cert,
			routing_profile_tls_cert,
			routing_profile_tls_key,
			gateway_discovery_enabled,
			gateway_discovery_interval,
			gateway_discovery_tx_frequency,
			gateway_discovery_dr
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		returning id`,
		n.CreatedAt,
		n.UpdatedAt,
		n.Name,
		n.Server,
		n.CACert,
		n.TLSCert,
		n.TLSKey,
		n.RoutingProfileCACert,
		n.RoutingProfileTLSCert,
		n.RoutingProfileTLSKey,
		n.GatewayDiscoveryEnabled,
		n.GatewayDiscoveryInterval,
		n.GatewayDiscoveryTXFrequency,
		n.GatewayDiscoveryDR,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	rpID, err := uuid.FromString(config.C.ApplicationServer.ID)
	if err != nil {
		return errors.Wrap(err, "uuid from string error")
	}

	_, err = nsClient.CreateRoutingProfile(ctx, &ns.CreateRoutingProfileRequest{
		RoutingProfile: &ns.RoutingProfile{
			Id:      rpID.Bytes(),
			AsId:    config.C.ApplicationServer.API.PublicHost,
			CaCert:  n.RoutingProfileCACert,
			TlsCert: n.RoutingProfileTLSCert,
			TlsKey:  n.RoutingProfileTLSKey,
		},
	})
	if err != nil {
		return errors.Wrap(err, "create routing-profile error")
	}

	log.WithFields(log.Fields{
		"id":     n.ID,
		"name":   n.Name,
		"server": n.Server,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("network-server created")
	return nil
}

// GetNetworkServer returns the network-server matching the given id.
func GetNetworkServer(ctx context.Context, db sqlx.Queryer, id int64) (NetworkServer, error) {
	var ns NetworkServer
	err := sqlx.Get(db, &ns, "select * from network_server where id = $1", id)
	if err != nil {
		return ns, handlePSQLError(Select, err, "select error")
	}

	return ns, nil
}

// UpdateNetworkServer updates the given network-server.
func UpdateNetworkServer(ctx context.Context, db sqlx.Execer, n *NetworkServer) error {
	if err := n.Validate(); err != nil {
		return errors.Wrap(err, "validation error")
	}

	n.UpdatedAt = time.Now()

	res, err := db.Exec(`
		update network_server
		set
			updated_at = $2,
			name = $3,
			server = $4,
			ca_cert = $5,
			tls_cert = $6,
			tls_key = $7,
			routing_profile_ca_cert = $8,
			routing_profile_tls_cert = $9,
			routing_profile_tls_key = $10,
			gateway_discovery_enabled = $11,
			gateway_discovery_interval = $12,
			gateway_discovery_tx_frequency = $13,
			gateway_discovery_dr = $14
		where id = $1`,
		n.ID,
		n.UpdatedAt,
		n.Name,
		n.Server,
		n.CACert,
		n.TLSCert,
		n.TLSKey,
		n.RoutingProfileCACert,
		n.RoutingProfileTLSCert,
		n.RoutingProfileTLSKey,
		n.GatewayDiscoveryEnabled,
		n.GatewayDiscoveryInterval,
		n.GatewayDiscoveryTXFrequency,
		n.GatewayDiscoveryDR,
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

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	rpID, err := uuid.FromString(config.C.ApplicationServer.ID)
	if err != nil {
		return errors.Wrap(err, "uuid from string error")
	}

	_, err = nsClient.UpdateRoutingProfile(ctx, &ns.UpdateRoutingProfileRequest{
		RoutingProfile: &ns.RoutingProfile{
			Id:      rpID.Bytes(),
			AsId:    config.C.ApplicationServer.API.PublicHost,
			CaCert:  n.RoutingProfileCACert,
			TlsCert: n.RoutingProfileTLSCert,
			TlsKey:  n.RoutingProfileTLSKey,
		},
	})
	if err != nil {
		return errors.Wrap(err, "update routing-profile error")
	}

	log.WithFields(log.Fields{
		"id":     n.ID,
		"name":   n.Name,
		"server": n.Server,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("network-server updated")
	return nil
}

// DeleteNetworkServer deletes the network-server matching the given id.
func DeleteNetworkServer(ctx context.Context, db sqlx.Ext, id int64) error {
	n, err := GetNetworkServer(ctx, db, id)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	res, err := db.Exec("delete from network_server where id = $1", id)
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

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	rpID, err := uuid.FromString(config.C.ApplicationServer.ID)
	if err != nil {
		return errors.Wrap(err, "uuid from string error")
	}

	_, err = nsClient.DeleteRoutingProfile(ctx, &ns.DeleteRoutingProfileRequest{
		Id: rpID.Bytes(),
	})
	if err != nil {
		return errors.Wrap(err, "delete routing-profile error")
	}

	log.WithFields(log.Fields{
		"id":     id,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("network-server deleted")
	return nil
}

// NetworkServerFilters provides filters for filtering network-servers.
type NetworkServerFilters struct {
	OrganizationID int64 `db:"organization_id"`

	// Limit and Offset are added for convenience so that this struct can
	// be given as the arguments.
	Limit  int `db:"limit"`
	Offset int `db:"offset"`
}

// SQL returns the SQL filters.
func (f NetworkServerFilters) SQL() string {
	var filters []string

	if f.OrganizationID != 0 {
		filters = append(filters, "sp.organization_id = :organization_id")
	}

	if len(filters) == 0 {
		return ""
	}

	return "where " + strings.Join(filters, " and ")
}

// GetNetworkServerCount returns the total number of network-servers.
func GetNetworkServerCount(ctx context.Context, db sqlx.Queryer, filters NetworkServerFilters) (int, error) {
	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			count(distinct ns.id)
		from
			network_server ns
		left join service_profile sp
			on ns.id = sp.network_server_id
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

// GetNetworkServers returns a slice of network-servers.
func GetNetworkServers(ctx context.Context, db sqlx.Queryer, filters NetworkServerFilters) ([]NetworkServer, error) {
	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			distinct ns.*
		from
			network_server ns
		left join service_profile sp
			on ns.id = sp.network_server_id
	`+filters.SQL()+`
		order by ns.name
		limit :limit
		offset :offset
	`, filters)
	if err != nil {
		return nil, errors.Wrap(err, "named query error")
	}

	var nss []NetworkServer
	err = sqlx.Select(db, &nss, query, args...)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return nss, nil
}

// GetNetworkServerForDevEUI returns the network-server for the given DevEUI.
func GetNetworkServerForDevEUI(ctx context.Context, db sqlx.Queryer, devEUI lorawan.EUI64) (NetworkServer, error) {
	var n NetworkServer
	err := sqlx.Get(db, &n, `
		select
			ns.*
		from
			network_server ns
		inner join device_profile dp
			on dp.network_server_id = ns.id
		inner join device d
			on d.device_profile_id = dp.device_profile_id
		where
			d.dev_eui = $1`,
		devEUI,
	)
	if err != nil {
		return n, handlePSQLError(Select, err, "select error")
	}
	return n, nil
}

// GetNetworkServerForDeviceProfileID returns the network-server for the given
// device-profile id.
func GetNetworkServerForDeviceProfileID(ctx context.Context, db sqlx.Queryer, id uuid.UUID) (NetworkServer, error) {
	var n NetworkServer
	err := sqlx.Get(db, &n, `
		select
			ns.*
		from
			network_server ns
		inner join device_profile dp
			on dp.network_server_id = ns.id
		where
			dp.device_profile_id = $1`,
		id,
	)
	if err != nil {
		return n, handlePSQLError(Select, err, "select error")
	}
	return n, nil
}

// GetNetworkServerForServiceProfileID returns the network-server for the given
// service-profile id.
func GetNetworkServerForServiceProfileID(ctx context.Context, db sqlx.Queryer, id uuid.UUID) (NetworkServer, error) {
	var n NetworkServer
	err := sqlx.Get(db, &n, `
		select
			ns.*
		from
			network_server ns
		inner join service_profile sp
			on sp.network_server_id = ns.id
		where
			sp.service_profile_id = $1`,
		id,
	)
	if err != nil {
		return n, handlePSQLError(Select, err, "select error")
	}
	return n, nil
}

// GetNetworkServerForGatewayMAC returns the network-server for a given
// gateway mac.
func GetNetworkServerForGatewayMAC(ctx context.Context, db sqlx.Queryer, mac lorawan.EUI64) (NetworkServer, error) {
	var n NetworkServer
	err := sqlx.Get(db, &n, `
		select
			ns.*
		from network_server ns
		inner join gateway gw
			on gw.network_server_id = ns.id
		where
			gw.mac = $1`,
		mac[:],
	)
	if err != nil {
		return n, handlePSQLError(Select, err, "select error")
	}
	return n, nil
}

// GetNetworkServerForGatewayProfileID returns the network-server for the given
// gateway-profile id.
func GetNetworkServerForGatewayProfileID(ctx context.Context, db sqlx.Queryer, id uuid.UUID) (NetworkServer, error) {
	var n NetworkServer
	err := sqlx.Get(db, &n, `
		select
			ns.*
		from
			network_server ns
		inner join gateway_profile gp
			on gp.network_server_id = ns.id
		where
			gp.gateway_profile_id = $1`,
		id,
	)
	if err != nil {
		return n, handlePSQLError(Select, err, "select errror")
	}
	return n, nil
}

// GetNetworkServerForMulticastGroupID returns the network-server for the given
// multicast-group id.
func GetNetworkServerForMulticastGroupID(ctx context.Context, db sqlx.Queryer, id uuid.UUID) (NetworkServer, error) {
	var n NetworkServer
	err := sqlx.Get(db, &n, `
		select
			ns.*
		from
			network_server ns
		inner join service_profile sp
			on sp.network_server_id = ns.id
		inner join application a
			on a.service_profile_id = sp.service_profile_id
		inner join multicast_group mg
			on mg.application_id = a.id
		where
			mg.id = $1
	`, id)
	if err != nil {
		return n, handlePSQLError(Select, err, "select error")
	}
	return n, nil
}

// GetNetworkServerForApplicationID returns the network-server for the given
// application ID.
func GetNetworkServerForApplicationID(ctx context.Context, db sqlx.Queryer, id int64) (NetworkServer, error) {
	var n NetworkServer
	err := sqlx.Get(db, &n, `
		select
			ns.*
		from
			network_server ns
		inner join service_profile sp
			on sp.network_server_id = ns.id
		inner join application a
			on a.service_profile_id = sp.service_profile_id
		where
			a.id = $1
	`, id)
	if err != nil {
		return n, handlePSQLError(Select, err, "select error")
	}
	return n, nil
}
