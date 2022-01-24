package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/lib/pq/hstore"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

var gatewayNameRegexp = regexp.MustCompile(`^[\w-]+$`)

// Gateway represents a gateway.
type Gateway struct {
	MAC              lorawan.EUI64 `db:"mac"`
	CreatedAt        time.Time     `db:"created_at"`
	UpdatedAt        time.Time     `db:"updated_at"`
	FirstSeenAt      *time.Time    `db:"first_seen_at"`
	LastSeenAt       *time.Time    `db:"last_seen_at"`
	Name             string        `db:"name"`
	Description      string        `db:"description"`
	OrganizationID   int64         `db:"organization_id"`
	Ping             bool          `db:"ping"`
	LastPingID       *int64        `db:"last_ping_id"`
	LastPingSentAt   *time.Time    `db:"last_ping_sent_at"`
	NetworkServerID  int64         `db:"network_server_id"`
	GatewayProfileID *uuid.UUID    `db:"gateway_profile_id"`
	ServiceProfileID *uuid.UUID    `db:"service_profile_id"`
	Latitude         float64       `db:"latitude"`
	Longitude        float64       `db:"longitude"`
	Altitude         float64       `db:"altitude"`
	Tags             hstore.Hstore `db:"tags"`
	Metadata         hstore.Hstore `db:"metadata"`
}

// GatewayListItem defines the gateway as list item.
type GatewayListItem struct {
	MAC               lorawan.EUI64 `db:"mac"`
	Name              string        `db:"name"`
	Description       string        `db:"description"`
	CreatedAt         time.Time     `db:"created_at"`
	UpdatedAt         time.Time     `db:"updated_at"`
	FirstSeenAt       *time.Time    `db:"first_seen_at"`
	LastSeenAt        *time.Time    `db:"last_seen_at"`
	OrganizationID    int64         `db:"organization_id"`
	NetworkServerID   int64         `db:"network_server_id"`
	Latitude          float64       `db:"latitude"`
	Longitude         float64       `db:"longitude"`
	Altitude          float64       `db:"altitude"`
	NetworkServerName string        `db:"network_server_name"`
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

// GatewaysActiveInactive holds the avtive and inactive counts.
type GatewaysActiveInactive struct {
	NeverSeenCount uint32 `db:"never_seen_count"`
	ActiveCount    uint32 `db:"active_count"`
	InactiveCount  uint32 `db:"inactive_count"`
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
func CreateGateway(ctx context.Context, db sqlx.Execer, gw *Gateway) error {
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
			network_server_id,
			gateway_profile_id,
			first_seen_at,
			last_seen_at,
			latitude,
			longitude,
			altitude,
			tags,
			metadata,
			service_profile_id
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
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
		gw.GatewayProfileID,
		gw.FirstSeenAt,
		gw.LastSeenAt,
		gw.Latitude,
		gw.Longitude,
		gw.Altitude,
		gw.Tags,
		gw.Metadata,
		gw.ServiceProfileID,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	log.WithFields(log.Fields{
		"id":     gw.MAC,
		"name":   gw.Name,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("gateway created")
	return nil
}

// UpdateGateway updates the given Gateway.
func UpdateGateway(ctx context.Context, db sqlx.Execer, gw *Gateway) error {
	if err := gw.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	gw.UpdatedAt = time.Now()

	res, err := db.Exec(`
		update gateway
			set updated_at = $2,
			name = $3,
			description = $4,
			organization_id = $5,
			ping = $6,
			last_ping_id = $7,
			last_ping_sent_at = $8,
			network_server_id = $9,
			gateway_profile_id = $10,
			first_seen_at = $11,
			last_seen_at = $12,
			latitude = $13,
			longitude = $14,
			altitude = $15,
			tags = $16,
			metadata = $17,
			service_profile_id = $18
		where
			mac = $1`,
		gw.MAC[:],
		gw.UpdatedAt,
		gw.Name,
		gw.Description,
		gw.OrganizationID,
		gw.Ping,
		gw.LastPingID,
		gw.LastPingSentAt,
		gw.NetworkServerID,
		gw.GatewayProfileID,
		gw.FirstSeenAt,
		gw.LastSeenAt,
		gw.Latitude,
		gw.Longitude,
		gw.Altitude,
		gw.Tags,
		gw.Metadata,
		gw.ServiceProfileID,
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
		"id":     gw.MAC,
		"name":   gw.Name,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("gateway updated")
	return nil
}

// DeleteGateway deletes the gateway matching the given MAC.
func DeleteGateway(ctx context.Context, db sqlx.Ext, mac lorawan.EUI64) error {
	n, err := GetNetworkServerForGatewayMAC(ctx, db, mac)
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

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.DeleteGateway(ctx, &ns.DeleteGatewayRequest{
		Id: mac[:],
	})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return errors.Wrap(err, "delete gateway error")
	}

	log.WithFields(log.Fields{
		"id":     mac,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("gateway deleted")
	return nil
}

// GetGateway returns the gateway for the given mac.
func GetGateway(ctx context.Context, db sqlx.Queryer, mac lorawan.EUI64, forUpdate bool) (Gateway, error) {
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

// GatewayFilters provides filters for filtering gateways.
type GatewayFilters struct {
	OrganizationID int64  `db:"organization_id"`
	UserID         int64  `db:"user_id"`
	Search         string `db:"search"`

	// Limit and Offset are added for convenience so that this struct can
	// be given as the arguments.
	Limit  int `db:"limit"`
	Offset int `db:"offset"`
}

// SQL returns the SQL filters.
func (f GatewayFilters) SQL() string {
	var filters []string

	if f.OrganizationID != 0 {
		filters = append(filters, "g.organization_id = :organization_id")
	}

	if f.UserID != 0 {
		filters = append(filters, "u.id = :user_id")
	}

	if f.Search != "" {
		filters = append(filters, "(g.name ilike :search or encode(g.mac, 'hex') ilike :search)")
	}

	if len(filters) == 0 {
		return ""
	}

	return "where " + strings.Join(filters, " and ")
}

// GetGatewayCount returns the total number of gateways.
func GetGatewayCount(ctx context.Context, db sqlx.Queryer, filters GatewayFilters) (int, error) {
	if filters.Search != "" {
		filters.Search = "%" + filters.Search + "%"
	}

	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			count(distinct g.*)
		from
			gateway g
		inner join organization o
			on o.id = g.organization_id
		left join organization_user ou
			on o.id = ou.organization_id
		left join "user" u
			on ou.user_id = u.id
	`+filters.SQL(), filters)
	if err != nil {
		return 0, errors.Wrap(err, "named query error")
	}

	var count int
	err = sqlx.Get(db, &count, query, args...)
	if err != nil {

		return 0, errors.Wrap(err, "named query error")
	}

	return count, nil
}

// GetGateways returns a slice of gateways sorted by name.
func GetGateways(ctx context.Context, db sqlx.Queryer, filters GatewayFilters) ([]GatewayListItem, error) {
	if filters.Search != "" {
		filters.Search = "%" + filters.Search + "%"
	}

	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			distinct g.mac,
			g.name,
			g.description,
			g.created_at,
			g.updated_at,
			g.first_seen_at,
			g.last_seen_at,
			g.organization_id,
			g.network_server_id,
			g.latitude,
			g.longitude,
			g.altitude,
			n.name as network_server_name
		from
			gateway g
		inner join organization o
			on o.id = g.organization_id
		inner join network_server n
			on n.id = g.network_server_id
		left join organization_user ou
			on o.id = ou.organization_id
		left join "user" u
			on ou.user_id = u.id
	`+filters.SQL()+`
		order by
			g.name
		limit :limit
		offset :offset
	`, filters)

	var gws []GatewayListItem
	err = sqlx.Select(db, &gws, query, args...)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return gws, nil
}

// GetGatewaysForMACs returns a map of gateways given a slice of MACs.
func GetGatewaysForMACs(ctx context.Context, db sqlx.Queryer, macs []lorawan.EUI64) (map[lorawan.EUI64]Gateway, error) {
	out := make(map[lorawan.EUI64]Gateway)
	var macsB [][]byte
	for i := range macs {
		macsB = append(macsB, macs[i][:])
	}

	var gws []Gateway
	err := sqlx.Select(db, &gws, "select * from gateway where mac = any($1)", pq.ByteaArray(macsB))
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	if len(gws) != len(macs) {
		log.WithFields(log.Fields{
			"expected": len(macs),
			"returned": len(gws),
			"ctx_id":   ctx.Value(logging.ContextIDKey),
		}).Warning("requested number of gateways does not match returned")
	}

	for i := range gws {
		out[gws[i].MAC] = gws[i]
	}

	return out, nil
}

// CreateGatewayPing creates the given gateway ping.
func CreateGatewayPing(ctx context.Context, db sqlx.Queryer, ping *GatewayPing) error {
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
		"ctx_id":      ctx.Value(logging.ContextIDKey),
	}).Info("gateway ping created")

	return nil
}

// GetGatewayPing returns the ping matching the given id.
func GetGatewayPing(ctx context.Context, db sqlx.Queryer, id int64) (GatewayPing, error) {
	var ping GatewayPing
	err := sqlx.Get(db, &ping, "select * from gateway_ping where id = $1", id)
	if err != nil {
		return ping, handlePSQLError(Select, err, "select error")
	}

	return ping, nil
}

// CreateGatewayPingRX creates the received ping.
func CreateGatewayPingRX(ctx context.Context, db sqlx.Queryer, rx *GatewayPingRX) error {
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
func DeleteAllGatewaysForOrganizationID(ctx context.Context, db sqlx.Ext, organizationID int64) error {
	var gws []Gateway
	err := sqlx.Select(db, &gws, "select * from gateway where organization_id = $1", organizationID)
	if err != nil {
		return handlePSQLError(Select, err, "select error")
	}

	for _, gw := range gws {
		err = DeleteGateway(ctx, db, gw.MAC)
		if err != nil {
			return errors.Wrap(err, "delete gateway error")
		}
	}

	return nil
}

// GetGatewayPingRXForPingID returns the received gateway pings for the given
// ping ID.
func GetGatewayPingRXForPingID(ctx context.Context, db sqlx.Queryer, pingID int64) ([]GatewayPingRX, error) {
	var rx []GatewayPingRX

	err := sqlx.Select(db, &rx, "select * from gateway_ping_rx where ping_id = $1", pingID)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return rx, nil
}

// GetLastGatewayPingAndRX returns the last gateway ping and RX for the given
// gateway MAC.
func GetLastGatewayPingAndRX(ctx context.Context, db sqlx.Queryer, mac lorawan.EUI64) (GatewayPing, []GatewayPingRX, error) {
	gw, err := GetGateway(ctx, db, mac, false)
	if err != nil {
		return GatewayPing{}, nil, errors.Wrap(err, "get gateway error")
	}

	if gw.LastPingID == nil {
		return GatewayPing{}, nil, ErrDoesNotExist
	}

	ping, err := GetGatewayPing(ctx, db, *gw.LastPingID)
	if err != nil {
		return GatewayPing{}, nil, errors.Wrap(err, "get gateway ping error")
	}

	rx, err := GetGatewayPingRXForPingID(ctx, db, ping.ID)
	if err != nil {
		return GatewayPing{}, nil, errors.Wrap(err, "get gateway ping rx for ping id error")
	}

	return ping, rx, nil
}

// GetGatewaysActiveInactive returns the active / inactive gateways.
func GetGatewaysActiveInactive(ctx context.Context, db sqlx.Queryer, organizationID int64) (GatewaysActiveInactive, error) {
	var out GatewaysActiveInactive
	err := sqlx.Get(db, &out, `
		with gateway_active_inactive as (
			select
				g.last_seen_at as last_seen_at,
				make_interval(secs => coalesce(gp.stats_interval / 1000000000, 30)) * 1.5 as stats_interval
			from
				gateway g
			left join gateway_profile gp
				on g.gateway_profile_id = gp.gateway_profile_id
			where
				$1 = 0 or g.organization_id = $1
		)
		select
			coalesce(sum(case when last_seen_at is null then 1 end), 0) as never_seen_count,
			coalesce(sum(case when (now() - stats_interval) > last_seen_at then 1 end), 0) as inactive_count,
			coalesce(sum(case when (now() - stats_interval) <= last_seen_at then 1 end), 0) as active_count
		from
			gateway_active_inactive
	`, organizationID)
	if err != nil {
		return out, errors.Wrap(err, "get gateway active/inactive count error")
	}

	return out, nil
}
