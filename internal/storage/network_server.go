package storage

import (
	"context"
	"time"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/loraserver/api/ns"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// NetworkServer defines the information to connect to a network-server.
type NetworkServer struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Name      string    `db:"name"`
	Server    string    `db:"server"`
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
			server
		) values ($1, $2, $3, $4)
		returning id`,
		n.CreatedAt,
		n.UpdatedAt,
		n.Name,
		n.Server,
	)
	if err != nil {
		return handlePSQLError(err, "insert error")
	}

	_, err = common.NetworkServer.CreateRoutingProfile(context.Background(), &ns.CreateRoutingProfileRequest{
		RoutingProfile: &ns.RoutingProfile{
			RoutingProfileID: common.ApplicationServerID,
			AsID:             common.ApplicationServerServer,
		},
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
		return ns, handlePSQLError(err, "select error")
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
			server = $4
		where id = $1`,
		n.ID,
		n.UpdatedAt,
		n.Name,
		n.Server,
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

	_, err = common.NetworkServer.UpdateRoutingProfile(context.Background(), &ns.UpdateRoutingProfileRequest{
		RoutingProfile: &ns.RoutingProfile{
			RoutingProfileID: common.ApplicationServerID,
			AsID:             common.ApplicationServerServer,
		},
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
func DeleteNetworkServer(db sqlx.Execer, id int64) error {
	res, err := db.Exec("delete from network_server where id = $1", id)
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

	_, err = common.NetworkServer.DeleteRoutingProfile(context.Background(), &ns.DeleteRoutingProfileRequest{
		RoutingProfileID: common.ApplicationServerID,
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
		return 0, handlePSQLError(err, "select error")
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
			count (ns.*)
		from
			network_server ns
		inner join service_profile sp
			on sp.network_server_id = ns.id
		where
			sp.organization_id = $1`,
		organizationID,
	)
	if err != nil {
		return 0, handlePSQLError(err, "select error")
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
		return nil, handlePSQLError(err, "select error")
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
		order by name
		limit $2 offset $3`,
		organizationID,
		limit,
		offset,
	)
	if err != nil {
		return nil, handlePSQLError(err, "select error")
	}

	return nss, nil
}
