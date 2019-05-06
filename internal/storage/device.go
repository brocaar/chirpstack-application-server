package storage

import (
	"context"
	"strings"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

// Device defines a LoRaWAN device.
type Device struct {
	DevEUI                    lorawan.EUI64 `db:"dev_eui"`
	CreatedAt                 time.Time     `db:"created_at"`
	UpdatedAt                 time.Time     `db:"updated_at"`
	LastSeenAt                *time.Time    `db:"last_seen_at"`
	ApplicationID             int64         `db:"application_id"`
	DeviceProfileID           uuid.UUID     `db:"device_profile_id"`
	Name                      string        `db:"name"`
	Description               string        `db:"description"`
	SkipFCntCheck             bool          `db:"-"`
	ReferenceAltitude         float64       `db:"-"`
	DeviceStatusBattery       *float32      `db:"device_status_battery"`
	DeviceStatusMargin        *int          `db:"device_status_margin"`
	DeviceStatusExternalPower bool          `db:"device_status_external_power_source"`
	DR                        *int          `db:"dr"`
	Latitude                  *float64      `db:"latitude"`
	Longitude                 *float64      `db:"longitude"`
	Altitude                  *float64      `db:"altitude"`
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
	GenAppKey lorawan.AES128Key `db:"gen_app_key"`
	JoinNonce int               `db:"join_nonce"`
}

// DeviceActivation defines the device-activation for a LoRaWAN device.
type DeviceActivation struct {
	ID        int64             `db:"id"`
	CreatedAt time.Time         `db:"created_at"`
	DevEUI    lorawan.EUI64     `db:"dev_eui"`
	DevAddr   lorawan.DevAddr   `db:"dev_addr"`
	AppSKey   lorawan.AES128Key `db:"app_s_key"`
}

// CreateDevice creates the given device.
func CreateDevice(db sqlx.Ext, d *Device) error {
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
			dr
        ) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
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
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	app, err := GetApplication(db, d.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get application error")
	}

	n, err := GetNetworkServerForDevEUI(db, d.DevEUI)
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

	_, err = nsClient.CreateDevice(context.Background(), &ns.CreateDeviceRequest{
		Device: &ns.Device{
			DevEui:            d.DevEUI[:],
			DeviceProfileId:   d.DeviceProfileID.Bytes(),
			ServiceProfileId:  app.ServiceProfileID.Bytes(),
			RoutingProfileId:  rpID.Bytes(),
			SkipFCntCheck:     d.SkipFCntCheck,
			ReferenceAltitude: d.ReferenceAltitude,
		},
	})
	if err != nil {
		log.WithError(err).Error("network-server create device api error")
		return handleGrpcError(err, "create device error")
	}

	log.WithFields(log.Fields{
		"dev_eui": d.DevEUI,
	}).Info("device created")

	return nil
}

// GetDevice returns the device matching the given DevEUI.
// When forUpdate is set to true, then db must be a db transaction.
// When localOnly is set to true, no call to the network-server is made to
// retrieve additional device data.
func GetDevice(db sqlx.Queryer, devEUI lorawan.EUI64, forUpdate, localOnly bool) (Device, error) {
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

	n, err := GetNetworkServerForDevEUI(db, d.DevEUI)
	if err != nil {
		return d, errors.Wrap(err, "get network-server error")
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return d, errors.Wrap(err, "get network-server client error")
	}

	resp, err := nsClient.GetDevice(context.Background(), &ns.GetDeviceRequest{
		DevEui: d.DevEUI[:],
	})
	if err != nil {
		return d, err
	}

	if resp.Device != nil {
		d.SkipFCntCheck = resp.Device.SkipFCntCheck
		d.ReferenceAltitude = resp.Device.ReferenceAltitude
	}

	return d, nil
}

// DeviceFilters provide filters that can be used to filter on devices.
// Note that empty values are not used as filter.
type DeviceFilters struct {
	ApplicationID    int64     `db:"application_id"`
	MulticastGroupID uuid.UUID `db:"multicast_group_id"`
	ServiceProfileID uuid.UUID `db:"service_profile_id"`
	Search           string    `db:"search"`

	// Limit and Offset are added for convenience so that this struct can
	// be given as the arguments.
	Limit  int `db:"limit"`
	Offset int `db:"offset"`
}

// SQL returns the SQL filter.
func (f DeviceFilters) SQL() string {
	var filters []string

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

	if len(filters) == 0 {
		return ""
	}

	return "where " + strings.Join(filters, " and ")
}

// GetDeviceCount returns the number of devices.
func GetDeviceCount(db sqlx.Queryer, filters DeviceFilters) (int, error) {
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
func GetDevices(db sqlx.Queryer, filters DeviceFilters) ([]DeviceListItem, error) {
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
func UpdateDevice(db sqlx.Ext, d *Device, localOnly bool) error {
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
			dr = $14
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
		app, err := GetApplication(db, d.ApplicationID)
		if err != nil {
			return errors.Wrap(err, "get application error")
		}

		n, err := GetNetworkServerForDevEUI(db, d.DevEUI)
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

		_, err = nsClient.UpdateDevice(context.Background(), &ns.UpdateDeviceRequest{
			Device: &ns.Device{
				DevEui:            d.DevEUI[:],
				DeviceProfileId:   d.DeviceProfileID.Bytes(),
				ServiceProfileId:  app.ServiceProfileID.Bytes(),
				RoutingProfileId:  rpID.Bytes(),
				SkipFCntCheck:     d.SkipFCntCheck,
				ReferenceAltitude: d.ReferenceAltitude,
			},
		})
		if err != nil {
			log.WithError(err).Error("network-server update device api error")
			return handleGrpcError(err, "update device error")
		}
	}

	log.WithFields(log.Fields{
		"dev_eui": d.DevEUI,
	}).Info("device updated")

	return nil
}

// DeleteDevice deletes the device matching the given DevEUI.
func DeleteDevice(db sqlx.Ext, devEUI lorawan.EUI64) error {
	n, err := GetNetworkServerForDevEUI(db, devEUI)
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

	_, err = nsClient.DeleteDevice(context.Background(), &ns.DeleteDeviceRequest{
		DevEui: devEUI[:],
	})
	if err != nil && grpc.Code(err) != codes.NotFound {
		log.WithError(err).Error("network-server delete device api error")
		return handleGrpcError(err, "delete device error")
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
	}).Info("device deleted")

	return nil
}

// CreateDeviceKeys creates the keys for the given device.
func CreateDeviceKeys(db sqlx.Execer, dc *DeviceKeys) error {
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
			join_nonce,
			gen_app_key
        ) values ($1, $2, $3, $4, $5, $6, $7)`,
		dc.CreatedAt,
		dc.UpdatedAt,
		dc.DevEUI[:],
		dc.NwkKey[:],
		dc.AppKey[:],
		dc.JoinNonce,
		dc.GenAppKey[:],
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	log.WithFields(log.Fields{
		"dev_eui": dc.DevEUI,
	}).Info("device-keys created")

	return nil
}

// GetDeviceKeys returns the device-keys for the given DevEUI.
func GetDeviceKeys(db sqlx.Queryer, devEUI lorawan.EUI64) (DeviceKeys, error) {
	var dc DeviceKeys

	err := sqlx.Get(db, &dc, "select * from device_keys where dev_eui = $1", devEUI[:])
	if err != nil {
		return dc, handlePSQLError(Select, err, "select error")
	}

	return dc, nil
}

// UpdateDeviceKeys updates the given device-keys.
func UpdateDeviceKeys(db sqlx.Execer, dc *DeviceKeys) error {
	dc.UpdatedAt = time.Now()

	res, err := db.Exec(`
        update device_keys
        set
            updated_at = $2,
			nwk_key = $3,
			app_key = $4,
			join_nonce = $5,
			gen_app_key = $6
        where
            dev_eui = $1`,
		dc.DevEUI[:],
		dc.UpdatedAt,
		dc.NwkKey[:],
		dc.AppKey[:],
		dc.JoinNonce,
		dc.GenAppKey[:],
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
	}).Info("device-keys updated")

	return nil
}

// DeleteDeviceKeys deletes the device-keys for the given DevEUI.
func DeleteDeviceKeys(db sqlx.Execer, devEUI lorawan.EUI64) error {
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

	log.WithField("dev_eui", devEUI).Info("device-keys deleted")

	return nil
}

// CreateDeviceActivation creates the given device-activation.
func CreateDeviceActivation(db sqlx.Queryer, da *DeviceActivation) error {
	da.CreatedAt = time.Now()

	err := sqlx.Get(db, &da.ID, `
        insert into device_activation (
            created_at,
            dev_eui,
            dev_addr,
			app_s_key
        ) values ($1, $2, $3, $4)
        returning id`,
		da.CreatedAt,
		da.DevEUI[:],
		da.DevAddr[:],
		da.AppSKey[:],
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	log.WithFields(log.Fields{
		"id":      da.ID,
		"dev_eui": da.DevEUI,
	}).Info("device-activation created")

	return nil
}

// GetLastDeviceActivationForDevEUI returns the most recent device-activation for the given DevEUI.
func GetLastDeviceActivationForDevEUI(db sqlx.Queryer, devEUI lorawan.EUI64) (DeviceActivation, error) {
	var da DeviceActivation

	err := sqlx.Get(db, &da, `
        select *
        from device_activation
        where
            dev_eui = $1
        order by
            created_at desc
        limit 1`,
		devEUI[:],
	)
	if err != nil {
		return da, handlePSQLError(Select, err, "select error")
	}

	return da, nil
}

// DeleteAllDevicesForApplicationID deletes all devices given an application id.
func DeleteAllDevicesForApplicationID(db sqlx.Ext, applicationID int64) error {
	var devs []Device
	err := sqlx.Select(db, &devs, "select * from device where application_id = $1", applicationID)
	if err != nil {
		return handlePSQLError(Select, err, "select error")
	}

	for _, dev := range devs {
		err = DeleteDevice(db, dev.DevEUI)
		if err != nil {
			return errors.Wrap(err, "delete device error")
		}
	}

	return nil
}
