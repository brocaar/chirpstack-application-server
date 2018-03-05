package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/brocaar/loraserver/api/ns"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var gatewayNameRegexp = regexp.MustCompile(`^[\w-]+$`)

// Gateway represents a gateway.
type Gateway struct {
	MAC             lorawan.EUI64 `db:"mac"`
	CreatedAt       time.Time     `db:"created_at"`
	UpdatedAt       time.Time     `db:"updated_at"`
	Name            string        `db:"name"`
	Description     string        `db:"description"`
	OrganizationID  int64         `db:"organization_id"`
	Ping            bool          `db:"ping"`
	LastPingID      *int64        `db:"last_ping_id"`
	LastPingSentAt  *time.Time    `db:"last_ping_sent_at"`
	NetworkServerID int64         `db:"network_server_id"`
}

// GatewayPing represents a gateway ping.
type GatewayPing struct {
	ID         int64         `db:"id"`
	CreatedAt  time.Time     `db:"created_at"`
	GatewayMAC lorawan.EUI64 `db:"gateway_mac"`
	Frequency  int           `db:"frequency"`
	DR         int           `db:"dr"`
}

// GatewayPingRX represents a ping received by one of the gateways.
type GatewayPingRX struct {
	ID         int64         `db:"id"`
	PingID     int64         `db:"ping_id"`
	CreatedAt  time.Time     `db:"created_at"`
	GatewayMAC lorawan.EUI64 `db:"gateway_mac"`
	ReceivedAt *time.Time    `db:"received_at"`
	RSSI       int           `db:"rssi"`
	LoRaSNR    float64       `db:"lora_snr"`
	Location   GPSPoint      `db:"location"`
	Altitude   float64       `db:"altitude"`
}

// GPSPoint contains a GPS point.
type GPSPoint struct {
	Latitude  float64
	Longitude float64
}

// Value implements the driver.Valuer interface.
func (l GPSPoint) Value() (driver.Value, error) {
	return fmt.Sprintf("(%s,%s)", strconv.FormatFloat(l.Latitude, 'f', -1, 64), strconv.FormatFloat(l.Longitude, 'f', -1, 64)), nil
}

// Scan implements the sql.Scanner interface.
func (l *GPSPoint) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", src)
	}

	_, err := fmt.Sscanf(string(b), "(%f,%f)", &l.Latitude, &l.Longitude)
	return err
}

// Validate validates the gateway data.
func (g Gateway) Validate() error {
	if !gatewayNameRegexp.MatchString(g.Name) {
		return ErrGatewayInvalidName
	}
	return nil
}

// CreateGateway creates the given Gateway.
func CreateGateway(db sqlx.Execer, gw *Gateway) error {
	if err := gw.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	now := time.Now()
	gw.CreatedAt = now
	gw.UpdatedAt = now

	_, err := db.Exec(`
		insert into gateway (
			mac,
			created_at,
			updated_at,
			name,
			description,
			organization_id,
			ping,
			last_ping_id,
			last_ping_sent_at,
			network_server_id
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		gw.MAC[:],
		gw.CreatedAt,
		gw.UpdatedAt,
		gw.Name,
		gw.Description,
		gw.OrganizationID,
		gw.Ping,
		gw.LastPingID,
		gw.LastPingSentAt,
		gw.NetworkServerID,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	log.WithFields(log.Fields{
		"mac":  gw.MAC,
		"name": gw.Name,
	}).Info("gateway created")
	return nil
}

// UpdateGateway updates the given Gateway.
func UpdateGateway(db sqlx.Execer, gw *Gateway) error {
	if err := gw.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	now := time.Now()

	res, err := db.Exec(`
		update gateway
			set updated_at = $2,
			name = $3,
			description = $4,
			organization_id = $5,
			ping = $6,
			last_ping_id = $7,
			last_ping_sent_at = $8,
			network_server_id = $9
		where
			mac = $1`,
		gw.MAC[:],
		now,
		gw.Name,
		gw.Description,
		gw.OrganizationID,
		gw.Ping,
		gw.LastPingID,
		gw.LastPingSentAt,
		gw.NetworkServerID,
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

	gw.UpdatedAt = now
	log.WithFields(log.Fields{
		"mac":  gw.MAC,
		"name": gw.Name,
	}).Info("gateway updated")
	return nil
}

// DeleteGateway deletes the gateway matching the given MAC.
func DeleteGateway(db sqlx.Ext, mac lorawan.EUI64) error {
	n, err := GetNetworkServerForGatewayMAC(db, mac)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	res, err := db.Exec("delete from gateway where mac = $1", mac[:])
	if err != nil {
		return errors.Wrap(err, "delete error")
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

	_, err = nsClient.DeleteGateway(context.Background(), &ns.DeleteGatewayRequest{
		Mac: mac[:],
	})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return errors.Wrap(err, "delete gateway error")
	}

	log.WithField("mac", mac).Info("gateway deleted")
	return nil
}

// GetGateway returns the gateway for the given mac.
func GetGateway(db sqlx.Queryer, mac lorawan.EUI64, forUpdate bool) (Gateway, error) {
	var fu string
	if forUpdate {
		fu = " for update"
	}

	var gw Gateway
	err := sqlx.Get(db, &gw, "select * from gateway where mac = $1"+fu, mac[:])
	if err != nil {
		if err == sql.ErrNoRows {
			return gw, ErrDoesNotExist
		}
	}
	return gw, nil
}

// GetGatewayCount returns the total number of gateways.
func GetGatewayCount(db sqlx.Queryer) (int, error) {
	var count int
	err := sqlx.Get(db, &count, "select count(*) from gateway")
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetGateways returns a slice of gateways sorted by name.
func GetGateways(db sqlx.Queryer, limit, offset int) ([]Gateway, error) {
	var gws []Gateway
	err := sqlx.Select(db, &gws, `
		select *
		from gateway
		order by name
		limit $1 offset $2`,
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return gws, nil
}

// GetGatewayCountForOrganizationID returns the total number of gateways
// given an organization ID.
func GetGatewayCountForOrganizationID(db sqlx.Queryer, organizationID int64) (int, error) {
	var count int
	err := sqlx.Get(db, &count, `
		select count(*)
		from gateway
		where
			organization_id = $1`,
		organizationID,
	)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetGatewaysForOrganizationID returns a slice of gateways sorted by name
// for the given organization ID.
func GetGatewaysForOrganizationID(db sqlx.Queryer, organizationID int64, limit, offset int) ([]Gateway, error) {
	var gws []Gateway
	err := sqlx.Select(db, &gws, `
		select *
		from gateway
		where
			organization_id = $1
		order by name
		limit $2 offset $3`,
		organizationID,
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return gws, nil
}

// GetGatewayCountForUser returns the total number of gateways to which the
// given user has access.
func GetGatewayCountForUser(db sqlx.Queryer, username string) (int, error) {
	var count int
	err := sqlx.Get(db, &count, `
		select count(g.*)
		from gateway g
		inner join organization o
			on o.id = g.organization_id
		inner join organization_user ou
			on ou.organization_id = o.id
		inner join "user" u
			on u.id = ou.user_id
		where
			u.username = $1`,
		username,
	)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetGatewaysForUser returns a slice of gateways sorted by name to which the
// given user has access.
func GetGatewaysForUser(db sqlx.Queryer, username string, limit, offset int) ([]Gateway, error) {
	var gws []Gateway
	err := sqlx.Select(db, &gws, `
		select g.*
		from gateway g
		inner join organization o
			on o.id = g.organization_id
		inner join organization_user ou
			on ou.organization_id = o.id
		inner join "user" u
			on u.id = ou.user_id
		where
			u.username = $1
		order by g.name
		limit $2 offset $3`,
		username,
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return gws, nil
}

// CreateGatewayPing creates the given gateway ping.
func CreateGatewayPing(db sqlx.Queryer, ping *GatewayPing) error {
	ping.CreatedAt = time.Now()

	err := sqlx.Get(db, &ping.ID, `
		insert into gateway_ping (
			created_at,
			gateway_mac,
			frequency,
			dr
		) values ($1, $2, $3, $4)
		returning id`,
		ping.CreatedAt,
		ping.GatewayMAC[:],
		ping.Frequency,
		ping.DR,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	log.WithFields(log.Fields{
		"gateway_mac": ping.GatewayMAC,
		"frequency":   ping.Frequency,
		"dr":          ping.DR,
		"id":          ping.ID,
	}).Info("gateway ping created")

	return nil
}

// GetGatewayPing returns the ping matching the given id.
func GetGatewayPing(db sqlx.Queryer, id int64) (GatewayPing, error) {
	var ping GatewayPing
	err := sqlx.Get(db, &ping, "select * from gateway_ping where id = $1", id)
	if err != nil {
		return ping, handlePSQLError(Select, err, "select error")
	}

	return ping, nil
}

// CreateGatewayPingRX creates the received ping.
func CreateGatewayPingRX(db sqlx.Queryer, rx *GatewayPingRX) error {
	rx.CreatedAt = time.Now()

	err := sqlx.Get(db, &rx.ID, `
		insert into gateway_ping_rx (
			ping_id,
			created_at,
			gateway_mac,
			received_at,
			rssi,
			lora_snr,
			location,
			altitude
		) values ($1, $2, $3, $4, $5, $6, $7, $8)
		returning id`,
		rx.PingID,
		rx.CreatedAt,
		rx.GatewayMAC[:],
		rx.ReceivedAt,
		rx.RSSI,
		rx.LoRaSNR,
		rx.Location,
		rx.Altitude,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	return nil
}

// DeleteAllGatewaysForOrganizationID deletes all gateways for a given
// organization id.
func DeleteAllGatewaysForOrganizationID(db sqlx.Ext, organizationID int64) error {
	var gws []Gateway
	err := sqlx.Select(db, &gws, "select * from gateway where organization_id = $1", organizationID)
	if err != nil {
		return handlePSQLError(Select, err, "select error")
	}

	for _, gw := range gws {
		err = DeleteGateway(db, gw.MAC)
		if err != nil {
			return errors.Wrap(err, "delete gateway error")
		}
	}

	return nil
}

// GetGatewayPingRXForPingID returns the received gateway pings for the given
// ping ID.
func GetGatewayPingRXForPingID(db sqlx.Queryer, pingID int64) ([]GatewayPingRX, error) {
	var rx []GatewayPingRX

	err := sqlx.Select(db, &rx, "select * from gateway_ping_rx where ping_id = $1", pingID)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return rx, nil
}

// GetLastGatewayPingAndRX returns the last gateway ping and RX for the given
// gateway MAC.
func GetLastGatewayPingAndRX(db sqlx.Queryer, mac lorawan.EUI64) (GatewayPing, []GatewayPingRX, error) {
	gw, err := GetGateway(db, mac, false)
	if err != nil {
		return GatewayPing{}, nil, errors.Wrap(err, "get gateway error")
	}

	if gw.LastPingID == nil {
		return GatewayPing{}, nil, ErrDoesNotExist
	}

	ping, err := GetGatewayPing(db, *gw.LastPingID)
	if err != nil {
		return GatewayPing{}, nil, errors.Wrap(err, "get gateway ping error")
	}

	rx, err := GetGatewayPingRXForPingID(db, ping.ID)
	if err != nil {
		return GatewayPing{}, nil, errors.Wrap(err, "get gateway ping rx for ping id error")
	}

	return ping, rx, nil
}
