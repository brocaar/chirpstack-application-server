package storage

import (
	"context"
	"strings"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq/hstore"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

// Device defines a LoRaWAN device.
type Device struct {
	DevEUI                    lorawan.EUI64     `db:"dev_eui"`
	CreatedAt                 time.Time         `db:"created_at"`
	UpdatedAt                 time.Time         `db:"updated_at"`
	LastSeenAt                *time.Time        `db:"last_seen_at"`
	ApplicationID             int64             `db:"application_id"`
	DeviceProfileID           uuid.UUID         `db:"device_profile_id"`
	Name                      string            `db:"name"`
	Description               string            `db:"description"`
	SkipFCntCheck             bool              `db:"-"`
	ReferenceAltitude         float64           `db:"-"`
	DeviceStatusBattery       *float32          `db:"device_status_battery"`
	DeviceStatusMargin        *int              `db:"device_status_margin"`
	DeviceStatusExternalPower bool              `db:"device_status_external_power_source"`
	DR                        *int              `db:"dr"`
	Latitude                  *float64          `db:"latitude"`
	Longitude                 *float64          `db:"longitude"`
	Altitude                  *float64          `db:"altitude"`
	DevAddr                   lorawan.DevAddr   `db:"dev_addr"`
	AppSKey                   lorawan.AES128Key `db:"app_s_key"`
	Variables                 hstore.Hstore     `db:"variables"`
	Tags                      hstore.Hstore     `db:"tags"`
	IsDisabled                bool              `db:"-"`
}

// DeviceListItem defines the Device as list item.
type DeviceListItem struct {
	Device
	DeviceProfileName string `db:"device_profile_name"`
}

// Validate validates the device data.
func (d Device) Validate() error {
	return nil
}

// DeviceKeys defines the keys for a LoRaWAN device.
type DeviceKeys struct {
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
	DevEUI    lorawan.EUI64     `db:"dev_eui"`
	NwkKey    lorawan.AES128Key `db:"nwk_key"`
	AppKey    lorawan.AES128Key `db:"app_key"`
	JoinNonce int               `db:"join_nonce"`
}

// DevicesActiveInactive holds the active and inactive counts.
type DevicesActiveInactive struct {
	NeverSeenCount uint32 `db:"never_seen_count"`
	ActiveCount    uint32 `db:"active_count"`
	InactiveCount  uint32 `db:"inactive_count"`
}

// DevicesDataRates holds the device counts by data-rate.
type DevicesDataRates map[uint32]uint32

// CreateDevice creates the given device.
func CreateDevice(ctx context.Context, db sqlx.Ext, d *Device) error {
	if err := d.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	now := time.Now()
	d.CreatedAt = now
	d.UpdatedAt = now

	_, err := db.Exec(`
        insert into device (
            dev_eui,
            created_at,
            updated_at,
            application_id,
            device_profile_id,
            name,
			description,
			device_status_battery,
			device_status_margin,
			device_status_external_power_source,
			last_seen_at,
			latitude,
			longitude,
			altitude,
			dr,
			variables,
			tags,
			dev_addr,
			app_s_key
        ) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		d.DevEUI[:],
		d.CreatedAt,
		d.UpdatedAt,
		d.ApplicationID,
		d.DeviceProfileID,
		d.Name,
		d.Description,
		d.DeviceStatusBattery,
		d.DeviceStatusMargin,
		d.DeviceStatusExternalPower,
		d.LastSeenAt,
		d.Latitude,
		d.Longitude,
		d.Altitude,
		d.DR,
		d.Variables,
		d.Tags,
		d.DevAddr[:],
		d.AppSKey,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	app, err := GetApplication(ctx, db, d.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get application error")
	}

	n, err := GetNetworkServerForDevEUI(ctx, db, d.DevEUI)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.CreateDevice(ctx, &ns.CreateDeviceRequest{
		Device: &ns.Device{
			DevEui:            d.DevEUI[:],
			DeviceProfileId:   d.DeviceProfileID.Bytes(),
			ServiceProfileId:  app.ServiceProfileID.Bytes(),
			RoutingProfileId:  applicationServerID.Bytes(),
			SkipFCntCheck:     d.SkipFCntCheck,
			ReferenceAltitude: d.ReferenceAltitude,
			IsDisabled:        d.IsDisabled,
		},
	})
	if err != nil {
		return errors.Wrap(err, "create device error")
	}

	log.WithFields(log.Fields{
		"dev_eui": d.DevEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("device created")

	return nil
}

// GetDevice returns the device matching the given DevEUI.
// When forUpdate is set to true, then db must be a db transaction.
// When localOnly is set to true, no call to the network-server is made to
// retrieve additional device data.
func GetDevice(ctx context.Context, db sqlx.Queryer, devEUI lorawan.EUI64, forUpdate, localOnly bool) (Device, error) {
	var fu string
	if forUpdate {
		fu = " for update"
	}

	var d Device
	err := sqlx.Get(db, &d, "select * from device where dev_eui = $1"+fu, devEUI[:])
	if err != nil {
		return d, handlePSQLError(Select, err, "select error")
	}

	if localOnly {
		return d, nil
	}

	n, err := GetNetworkServerForDevEUI(ctx, db, d.DevEUI)
	if err != nil {
		return d, errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return d, errors.Wrap(err, "get network-server client error")
	}

	resp, err := nsClient.GetDevice(ctx, &ns.GetDeviceRequest{
		DevEui: d.DevEUI[:],
	})
	if err != nil {
		return d, err
	}

	if resp.Device != nil {
		d.SkipFCntCheck = resp.Device.SkipFCntCheck
		d.ReferenceAltitude = resp.Device.ReferenceAltitude
		d.IsDisabled = resp.Device.IsDisabled
	}

	return d, nil
}

// DeviceFilters provide filters that can be used to filter on devices.
// Note that empty values are not used as filter.
type DeviceFilters struct {
	OrganizationID   int64         `db:"organization_id"`
	ApplicationID    int64         `db:"application_id"`
	MulticastGroupID uuid.UUID     `db:"multicast_group_id"`
	ServiceProfileID uuid.UUID     `db:"service_profile_id"`
	Search           string        `db:"search"`
	Tags             hstore.Hstore `db:"tags"`

	// Limit and Offset are added for convenience so that this struct can
	// be given as the arguments.
	Limit  int `db:"limit"`
	Offset int `db:"offset"`
}

// SQL returns the SQL filter.
func (f DeviceFilters) SQL() string {
	var filters []string

	if f.OrganizationID != 0 {
		filters = append(filters, "a.organization_id = :organization_id")
	}

	if f.ApplicationID != 0 {
		filters = append(filters, "d.application_id = :application_id")
	}

	if f.MulticastGroupID != uuid.Nil {
		filters = append(filters, "dmg.multicast_group_id = :multicast_group_id")
	}

	if f.ServiceProfileID != uuid.Nil {
		filters = append(filters, "a.service_profile_id = :service_profile_id")
	}

	if f.Search != "" {
		filters = append(filters, "(d.name ilike :search or encode(d.dev_eui, 'hex') ilike :search)")
	}

	if len(f.Tags.Map) != 0 {
		filters = append(filters, "d.tags @> :tags")
	}

	if len(filters) == 0 {
		return ""
	}

	return "where " + strings.Join(filters, " and ")
}

// GetDeviceCount returns the number of devices.
func GetDeviceCount(ctx context.Context, db sqlx.Queryer, filters DeviceFilters) (int, error) {
	if filters.Search != "" {
		filters.Search = "%" + filters.Search + "%"
	}

	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			count(distinct d.*)
		from device d
		inner join application a
			on d.application_id = a.id
		left join device_multicast_group dmg
			on d.dev_eui = dmg.dev_eui
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

// GetDevices returns a slice of devices.
func GetDevices(ctx context.Context, db sqlx.Queryer, filters DeviceFilters) ([]DeviceListItem, error) {
	if filters.Search != "" {
		filters.Search = "%" + filters.Search + "%"
	}

	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			distinct d.*,
			dp.name as device_profile_name
		from
			device d
		inner join device_profile dp
			on dp.device_profile_id = d.device_profile_id
		inner join application a
			on d.application_id = a.id
		left join device_multicast_group dmg
			on d.dev_eui = dmg.dev_eui
		`+filters.SQL()+`
		order by
			d.name
		limit :limit
		offset :offset
	`, filters)
	if err != nil {
		return nil, errors.Wrap(err, "named query error")
	}

	var devices []DeviceListItem
	err = sqlx.Select(db, &devices, query, args...)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return devices, nil
}

// UpdateDevice updates the given device.
// When localOnly is set, it will not update the device on the network-server.
func UpdateDevice(ctx context.Context, db sqlx.Ext, d *Device, localOnly bool) error {
	if err := d.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	d.UpdatedAt = time.Now()

	res, err := db.Exec(`
        update device
        set
            updated_at = $2,
            application_id = $3,
            device_profile_id = $4,
            name = $5,
			description = $6,
			device_status_battery = $7,
			device_status_margin = $8,
			last_seen_at = $9,
			latitude = $10,
			longitude = $11,
			altitude = $12,
			device_status_external_power_source = $13,
			dr = $14,
			variables = $15,
			tags = $16,
			dev_addr = $17,
			app_s_key = $18
        where
            dev_eui = $1`,
		d.DevEUI[:],
		d.UpdatedAt,
		d.ApplicationID,
		d.DeviceProfileID,
		d.Name,
		d.Description,
		d.DeviceStatusBattery,
		d.DeviceStatusMargin,
		d.LastSeenAt,
		d.Latitude,
		d.Longitude,
		d.Altitude,
		d.DeviceStatusExternalPower,
		d.DR,
		d.Variables,
		d.Tags,
		d.DevAddr,
		d.AppSKey,
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

	// update the device on the network-server
	if !localOnly {
		app, err := GetApplication(ctx, db, d.ApplicationID)
		if err != nil {
			return errors.Wrap(err, "get application error")
		}

		n, err := GetNetworkServerForDevEUI(ctx, db, d.DevEUI)
		if err != nil {
			return errors.Wrap(err, "get network-server error")
		}

		nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
		if err != nil {
			return errors.Wrap(err, "get network-server client error")
		}

		rpID, err := uuid.FromString(config.C.ApplicationServer.ID)
		if err != nil {
			return errors.Wrap(err, "uuid from string error")
		}

		_, err = nsClient.UpdateDevice(ctx, &ns.UpdateDeviceRequest{
			Device: &ns.Device{
				DevEui:            d.DevEUI[:],
				DeviceProfileId:   d.DeviceProfileID.Bytes(),
				ServiceProfileId:  app.ServiceProfileID.Bytes(),
				RoutingProfileId:  rpID.Bytes(),
				SkipFCntCheck:     d.SkipFCntCheck,
				ReferenceAltitude: d.ReferenceAltitude,
				IsDisabled:        d.IsDisabled,
			},
		})
		if err != nil {
			return errors.Wrap(err, "update device error")
		}
	}

	log.WithFields(log.Fields{
		"dev_eui": d.DevEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("device updated")

	return nil
}

// UpdateDeviceLastSeenAndDR updates the device last-seen timestamp and data-rate.
func UpdateDeviceLastSeenAndDR(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64, ts time.Time, dr int) error {
	res, err := db.Exec(`
		update device
		set
			last_seen_at = $2,
			dr = $3
		where
			dev_eui = $1`,
		devEUI[:],
		ts,
		dr,
	)
	if err != nil {
		return handlePSQLError(Update, err, "update last-seen and dr error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("device last-seen and dr updated")

	return nil
}

// UpdateDeviceActivation updates the device address and the AppSKey.
func UpdateDeviceActivation(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64, devAddr lorawan.DevAddr, appSKey lorawan.AES128Key) error {
	res, err := db.Exec(`
		update device
		set
			dev_addr = $2,
			app_s_key = $3
		where
			dev_eui = $1`,
		devEUI[:],
		devAddr[:],
		appSKey[:],
	)
	if err != nil {
		return handlePSQLError(Update, err, "update last-seen and dr error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	log.WithFields(log.Fields{
		"dev_eui":  devEUI,
		"dev_addr": devAddr,
		"ctx_id":   ctx.Value(logging.ContextIDKey),
	}).Info("device activation updated")

	return nil
}

// DeleteDevice deletes the device matching the given DevEUI.
func DeleteDevice(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64) error {
	n, err := GetNetworkServerForDevEUI(ctx, db, devEUI)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}

	res, err := db.Exec("delete from device where dev_eui = $1", devEUI[:])
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

	_, err = nsClient.DeleteDevice(ctx, &ns.DeleteDeviceRequest{
		DevEui: devEUI[:],
	})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return errors.Wrap(err, "delete device error")
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("device deleted")

	return nil
}

// CreateDeviceKeys creates the keys for the given device.
func CreateDeviceKeys(ctx context.Context, db sqlx.Execer, dc *DeviceKeys) error {
	now := time.Now()
	dc.CreatedAt = now
	dc.UpdatedAt = now

	_, err := db.Exec(`
        insert into device_keys (
            created_at,
            updated_at,
            dev_eui,
			nwk_key,
			app_key,
			join_nonce
        ) values ($1, $2, $3, $4, $5, $6)`,
		dc.CreatedAt,
		dc.UpdatedAt,
		dc.DevEUI[:],
		dc.NwkKey[:],
		dc.AppKey[:],
		dc.JoinNonce,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	log.WithFields(log.Fields{
		"dev_eui": dc.DevEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("device-keys created")

	return nil
}

// GetDeviceKeys returns the device-keys for the given DevEUI.
func GetDeviceKeys(ctx context.Context, db sqlx.Queryer, devEUI lorawan.EUI64) (DeviceKeys, error) {
	var dc DeviceKeys

	err := sqlx.Get(db, &dc, "select * from device_keys where dev_eui = $1", devEUI[:])
	if err != nil {
		return dc, handlePSQLError(Select, err, "select error")
	}

	return dc, nil
}

// UpdateDeviceKeys updates the given device-keys.
func UpdateDeviceKeys(ctx context.Context, db sqlx.Execer, dc *DeviceKeys) error {
	dc.UpdatedAt = time.Now()

	res, err := db.Exec(`
        update device_keys
        set
            updated_at = $2,
			nwk_key = $3,
			app_key = $4,
			join_nonce = $5
        where
            dev_eui = $1`,
		dc.DevEUI[:],
		dc.UpdatedAt,
		dc.NwkKey[:],
		dc.AppKey[:],
		dc.JoinNonce,
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
		"dev_eui": dc.DevEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("device-keys updated")

	return nil
}

// DeleteDeviceKeys deletes the device-keys for the given DevEUI.
func DeleteDeviceKeys(ctx context.Context, db sqlx.Execer, devEUI lorawan.EUI64) error {
	res, err := db.Exec("delete from device_keys where dev_eui = $1", devEUI[:])
	if err != nil {
		return handlePSQLError(Delete, err, "delete error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected errro")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("device-keys deleted")

	return nil
}

// DeleteAllDevicesForApplicationID deletes all devices given an application id.
func DeleteAllDevicesForApplicationID(ctx context.Context, db sqlx.Ext, applicationID int64) error {
	var devs []Device
	err := sqlx.Select(db, &devs, "select * from device where application_id = $1", applicationID)
	if err != nil {
		return handlePSQLError(Select, err, "select error")
	}

	for _, dev := range devs {
		err = DeleteDevice(ctx, db, dev.DevEUI)
		if err != nil {
			return errors.Wrap(err, "delete device error")
		}
	}

	return nil
}

// EnqueueDownlinkPayload adds the downlink payload to the network-server
// device-queue.
func EnqueueDownlinkPayload(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64, confirmed bool, fPort uint8, data []byte) (uint32, error) {
	// get network-server and network-server api client
	n, err := GetNetworkServerForDevEUI(ctx, db, devEUI)
	if err != nil {
		return 0, errors.Wrap(err, "get network-server error")
	}
	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return 0, errors.Wrap(err, "get network-server client error")
	}

	// get fCnt to use for encrypting and enqueueing
	resp, err := nsClient.GetNextDownlinkFCntForDevEUI(context.Background(), &ns.GetNextDownlinkFCntForDevEUIRequest{
		DevEui: devEUI[:],
	})
	if err != nil {
		return 0, errors.Wrap(err, "get next downlink fcnt for deveui error")
	}

	// get device
	d, err := GetDevice(ctx, db, devEUI, false, true)
	if err != nil {
		return 0, errors.Wrap(err, "get device error")
	}

	// encrypt payload
	b, err := lorawan.EncryptFRMPayload(d.AppSKey, false, d.DevAddr, resp.FCnt, data)
	if err != nil {
		return 0, errors.Wrap(err, "encrypt frmpayload error")
	}

	// enqueue device-queue item
	_, err = nsClient.CreateDeviceQueueItem(ctx, &ns.CreateDeviceQueueItemRequest{
		Item: &ns.DeviceQueueItem{
			DevAddr:    d.DevAddr[:],
			DevEui:     devEUI[:],
			FrmPayload: b,
			FCnt:       resp.FCnt,
			FPort:      uint32(fPort),
			Confirmed:  confirmed,
		},
	})
	if err != nil {
		return 0, errors.Wrap(err, "create device-queue item error")
	}

	log.WithFields(log.Fields{
		"f_cnt":     resp.FCnt,
		"dev_eui":   devEUI,
		"confirmed": confirmed,
	}).Info("downlink device-queue item handled")

	return resp.FCnt, nil
}

// GetDevicesActiveInactive returns the active / inactive devices.
func GetDevicesActiveInactive(ctx context.Context, db sqlx.Queryer, organizationID int64) (DevicesActiveInactive, error) {
	var out DevicesActiveInactive
	err := sqlx.Get(db, &out, `
		with device_active_inactive as (
			select
				make_interval(secs => dp.uplink_interval / 1000000000) * 1.5 as uplink_interval,
				d.last_seen_at as last_seen_at
			from
				device d
			inner join device_profile dp
				on d.device_profile_id = dp.device_profile_id
			inner join application a
				on d.application_id = a.id
			where
				$1 = 0 or a.organization_id = $1
		)
		select
			coalesce(sum(case when last_seen_at is null then 1 end), 0) as never_seen_count,
			coalesce(sum(case when (now() - uplink_interval) > last_seen_at then 1 end), 0) as inactive_count,
			coalesce(sum(case when (now() - uplink_interval) <= last_seen_at then 1 end), 0) as active_count
		from
			device_active_inactive
	`, organizationID)
	if err != nil {
		return out, errors.Wrap(err, "get device active/inactive count error")
	}

	return out, nil
}

// GetDevicesDataRates returns the device counts by data-rate.
func GetDevicesDataRates(ctx context.Context, db sqlx.Queryer, organizationID int64) (DevicesDataRates, error) {
	out := make(DevicesDataRates)

	rows, err := db.Queryx(`
		select
			d.dr,
			count(1)
		from
			device d
		inner join application a
			on d.application_id = a.id
		where
			($1 = 0 or a.organization_id = $1)
			and d.dr is not null
		group by d.dr
	`, organizationID)
	if err != nil {
		return out, errors.Wrap(err, "get device count per data-rate error")
	}

	for rows.Next() {
		var dr, count uint32

		if err := rows.Scan(&dr, &count); err != nil {
			return out, errors.Wrap(err, "scan row error")
		}

		out[dr] = count
	}

	return out, nil
}
