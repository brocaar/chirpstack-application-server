package storage

import (
	"context"
	"time"

	"github.com/brocaar/lorawan"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/brocaar/loraserver/api/ns"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// NetworkServer defines the information to connect to a network-server.
type NetworkServer struct {
	ID                    int64     `db:"id"`
	CreatedAt             time.Time `db:"created_at"`
	UpdatedAt             time.Time `db:"updated_at"`
	Name                  string    `db:"name"`
	Server                string    `db:"server"`
	CACert                string    `db:"ca_cert"`
	TLSCert               string    `db:"tls_cert"`
	TLSKey                string    `db:"tls_key"`
	RoutingProfileCACert  string    `db:"routing_profile_ca_cert"`
	RoutingProfileTLSCert string    `db:"routing_profile_tls_cert"`
	RoutingProfileTLSKey  string    `db:"routing_profile_tls_key"`
}

// Validate validates the network-server data.
func (ns NetworkServer) Validate() error {
	return nil
}

// CreateNetworkServer creates the given network-server.
func CreateNetworkServer(db sqlx.Queryer, n *NetworkServer) error {
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
			routing_profile_tls_key
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.CreateRoutingProfile(context.Background(), &ns.CreateRoutingProfileRequest{
		RoutingProfile: &ns.RoutingProfile{
			RoutingProfileID: config.C.ApplicationServer.ID,
			AsID:             config.C.ApplicationServer.API.PublicHost,
		},
		CaCert:  n.RoutingProfileCACert,
		TlsCert: n.RoutingProfileTLSCert,
		TlsKey:  n.RoutingProfileTLSKey,
	})
	if err != nil {
		log.WithError(err).Error("network-server create routing-profile api error")
		return handleGrpcError(err, "create routing-profile error")
	}

	log.WithFields(log.Fields{
		"id":     n.ID,
		"name":   n.Name,
		"server": n.Server,
	}).Info("network-server created")
	return nil
}

// GetNetworkServer returns the network-server matching the given id.
func GetNetworkServer(db sqlx.Queryer, id int64) (NetworkServer, error) {
	var ns NetworkServer
	err := sqlx.Get(db, &ns, "select * from network_server where id = $1", id)
	if err != nil {
		return ns, handlePSQLError(Select, err, "select error")
	}

	return ns, nil
}

// UpdateNetworkServer updates the given network-server.
func UpdateNetworkServer(db sqlx.Execer, n *NetworkServer) error {
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
			routing_profile_tls_key = $10
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

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.UpdateRoutingProfile(context.Background(), &ns.UpdateRoutingProfileRequest{
		RoutingProfile: &ns.RoutingProfile{
			RoutingProfileID: config.C.ApplicationServer.ID,
			AsID:             config.C.ApplicationServer.API.PublicHost,
		},
		CaCert:  n.RoutingProfileCACert,
		TlsCert: n.RoutingProfileTLSCert,
		TlsKey:  n.RoutingProfileTLSKey,
	})
	if err != nil {
		log.WithError(err).Error("network-server update routing-profile api error")
		return handleGrpcError(err, "update routing-profile error")
	}

	log.WithFields(log.Fields{
		"id":     n.ID,
		"name":   n.Name,
		"server": n.Server,
	}).Info("network-server updated")
	return nil
}

// DeleteNetworkServer deletes the network-server matching the given id.
func DeleteNetworkServer(db sqlx.Ext, id int64) error {
	n, err := GetNetworkServer(db, id)
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

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.DeleteRoutingProfile(context.Background(), &ns.DeleteRoutingProfileRequest{
		RoutingProfileID: config.C.ApplicationServer.ID,
	})
	if err != nil {
		log.WithError(err).Error("network-server delete routing-profile api error")
		return handleGrpcError(err, "delete routing-profile error")
	}

	log.WithField("id", id).Info("network-server deleted")
	return nil
}

// GetNetworkServerCount returns the total number of network-servers.
func GetNetworkServerCount(db sqlx.Queryer) (int, error) {
	var count int
	err := sqlx.Get(db, &count, "select count(*) from network_server")
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}

	return count, nil
}

// GetNetworkServerCountForOrganizationID returns the total number of
// network-servers accessible for the given organization id.
// A network-server is accessible for an organization when it is used by one
// of its service-profiles.
func GetNetworkServerCountForOrganizationID(db sqlx.Queryer, organizationID int64) (int, error) {
	var count int
	err := sqlx.Get(db, &count, `
		select
			count (distinct ns.id)
		from
			network_server ns
		inner join service_profile sp
			on sp.network_server_id = ns.id
		where
			sp.organization_id = $1`,
		organizationID,
	)
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}
	return count, nil
}

// GetNetworkServers returns a slice of network-servers.
func GetNetworkServers(db sqlx.Queryer, limit, offset int) ([]NetworkServer, error) {
	var nss []NetworkServer
	err := sqlx.Select(db, &nss, `
		select *
		from network_server
		order by name
		limit $1 offset $2`,
		limit,
		offset,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return nss, nil
}

// GetNetworkServersForOrganizationID returns a slice of network-server
// accessible for the given organization id.
// A network-server is accessible for an organization when it is used by one
// of its service-profiles.
func GetNetworkServersForOrganizationID(db sqlx.Queryer, organizationID int64, limit, offset int) ([]NetworkServer, error) {
	var nss []NetworkServer
	err := sqlx.Select(db, &nss, `
		select
			ns.*
		from
			network_server ns
		inner join service_profile sp
			on sp.network_server_id = ns.id
		where
			sp.organization_id = $1
		group by ns.id
		order by name
		limit $2 offset $3`,
		organizationID,
		limit,
		offset,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return nss, nil
}

// GetNetworkServerForDevEUI returns the network-server for the given DevEUI.
func GetNetworkServerForDevEUI(db sqlx.Queryer, devEUI lorawan.EUI64) (NetworkServer, error) {
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
func GetNetworkServerForDeviceProfileID(db sqlx.Queryer, id string) (NetworkServer, error) {
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
func GetNetworkServerForServiceProfileID(db sqlx.Queryer, id string) (NetworkServer, error) {
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
func GetNetworkServerForGatewayMAC(db sqlx.Queryer, mac lorawan.EUI64) (NetworkServer, error) {
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
