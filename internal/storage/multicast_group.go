package storage

import (
	"context"
	"strings"
	"time"

	"github.com/brocaar/lorawan"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/loraserver/api/ns"
)

// MulticastGroup defines the multicast-group.
type MulticastGroup struct {
	CreatedAt        time.Time         `db:"created_at"`
	UpdatedAt        time.Time         `db:"updated_at"`
	Name             string            `db:"name"`
	MCAppSKey        lorawan.AES128Key `db:"mc_app_s_key"`
	MCKey            lorawan.AES128Key `db:"mc_key"`
	FCnt             uint32            `db:"f_cnt"`
	ServiceProfileID uuid.UUID         `db:"service_profile_id"`
	MulticastGroup   ns.MulticastGroup `db:"-"`
}

// MulticastGroupListItem defines the multicast-group for listing.
type MulticastGroupListItem struct {
	ID                 uuid.UUID `db:"id"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
	Name               string    `db:"name"`
	ServiceProfileID   uuid.UUID `db:"service_profile_id"`
	ServiceProfileName string    `db:"service_profile_name"`
}

// CreateMulticastGroup creates the given multicast-group.
func CreateMulticastGroup(db sqlx.Ext, mg *MulticastGroup) error {
	mgID, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid v4 error")
	}

	now := time.Now()
	mg.MulticastGroup.Id = mgID.Bytes()
	mg.CreatedAt = now
	mg.UpdatedAt = now

	_, err = db.Exec(`
		insert into multicast_group (
			id,
			created_at,
			updated_at,
			name,
			service_profile_id,
			mc_app_s_key,
			mc_key,
			f_cnt
		) values ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		mgID,
		mg.CreatedAt,
		mg.UpdatedAt,
		mg.Name,
		mg.ServiceProfileID,
		mg.MCAppSKey,
		mg.MCKey,
		mg.FCnt,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	nsClient, err := getNSClientForServiceProfile(db, mg.ServiceProfileID)
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.CreateMulticastGroup(context.Background(), &ns.CreateMulticastGroupRequest{
		MulticastGroup: &mg.MulticastGroup,
	})
	if err != nil {
		return handleGrpcError(err, "create multicast-group error")
	}

	log.WithFields(log.Fields{
		"id": mgID,
	}).Info("multicast-group created")

	return nil
}

// GetMulticastGroup returns the multicast-group given an id.
func GetMulticastGroup(db sqlx.Queryer, id uuid.UUID, forUpdate, localOnly bool) (MulticastGroup, error) {
	var fu string
	if forUpdate {
		fu = " for update"
	}

	var mg MulticastGroup

	err := sqlx.Get(db, &mg, `
		select
			created_at,
			updated_at,
			name,
			service_profile_id,
			mc_app_s_key,
			mc_key,
			f_cnt
		from
			multicast_group
		where
			id = $1
	`+fu, id)
	if err != nil {
		return mg, handlePSQLError(Select, err, "select error")
	}

	if localOnly {
		return mg, nil
	}

	nsClient, err := getNSClientForServiceProfile(db, mg.ServiceProfileID)
	if err != nil {
		return mg, errors.Wrap(err, "get network-server client error")
	}

	resp, err := nsClient.GetMulticastGroup(context.Background(), &ns.GetMulticastGroupRequest{
		Id: id.Bytes(),
	})
	if err != nil {
		return mg, handleGrpcError(err, "get multicast-group error")
	}

	if resp.MulticastGroup == nil {
		return mg, errors.New("multicast_group must not be nil")
	}

	mg.MulticastGroup = *resp.MulticastGroup

	return mg, nil
}

// UpdateMulticastGroup updates the given multicast-group.
func UpdateMulticastGroup(db sqlx.Ext, mg *MulticastGroup) error {
	mgID, err := uuid.FromBytes(mg.MulticastGroup.Id)
	if err != nil {
		return errors.Wrap(err, "uuid from bytes error")
	}

	mg.UpdatedAt = time.Now()
	res, err := db.Exec(`
		update
			multicast_group
		set
			updated_at = $2,
			name = $3,
			mc_app_s_key = $4,
			mc_key = $5,
			f_cnt = $6
		where
			id = $1
	`,
		mgID,
		mg.UpdatedAt,
		mg.Name,
		mg.MCAppSKey,
		mg.MCKey,
		mg.FCnt,
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

	nsClient, err := getNSClientForServiceProfile(db, mg.ServiceProfileID)
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.UpdateMulticastGroup(context.Background(), &ns.UpdateMulticastGroupRequest{
		MulticastGroup: &mg.MulticastGroup,
	})
	if err != nil {
		return handleGrpcError(err, "update multicast-group error")
	}

	log.WithFields(log.Fields{
		"id": mgID,
	}).Info("multicast-group updated")

	return nil
}

// DeleteMulticastGroup deletes a multicast-group given an id.
func DeleteMulticastGroup(db sqlx.Ext, id uuid.UUID) error {
	nsClient, err := getNSClientForMulticastGroup(db, id)
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	res, err := db.Exec(`
		delete
		from
			multicast_group
		where
			id = $1
	`, id)
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

	_, err = nsClient.DeleteMulticastGroup(context.Background(), &ns.DeleteMulticastGroupRequest{
		Id: id.Bytes(),
	})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return handleGrpcError(err, "delete multicast-group error")
	}

	log.WithFields(log.Fields{
		"id": id,
	}).Info("multicast-group deleted")

	return nil
}

// MulticastGroupFilters provide filters that can be used to filter on
// multicast-groups. Note that empty values are not used as filters.
type MulticastGroupFilters struct {
	OrganizationID   int64         `db:"organization_id"`
	ServiceProfileID uuid.UUID     `db:"service_profile_id"`
	DevEUI           lorawan.EUI64 `db:"dev_eui"`
	Search           string        `db:"search"`

	// Limit and Offset are added for convenience so that this struct can
	// be given as the arguments.
	Limit  int `db:"limit"`
	Offset int `db:"offset"`
}

// SQL returns the SQL filter.
func (f MulticastGroupFilters) SQL() string {
	var filters []string
	var nilEUI lorawan.EUI64

	if f.OrganizationID != 0 {
		filters = append(filters, "o.id = :organization_id")
	}
	if f.ServiceProfileID != uuid.Nil {
		filters = append(filters, "mg.service_profile_id = :service_profile_id")
	}
	if f.DevEUI != nilEUI {
		filters = append(filters, "dmg.dev_eui = :dev_eui")
	}
	if f.Search != "" {
		filters = append(filters, "mg.name ilike :search")
	}

	if len(filters) == 0 {
		return ""
	}

	return "where " + strings.Join(filters, " and ")
}

// GetMulticastGroupCount returns the total number of multicast-groups given
// the provided filters. Note that empty values are not used as filters.
func GetMulticastGroupCount(db sqlx.Queryer, filters MulticastGroupFilters) (int, error) {
	if filters.Search != "" {
		filters.Search = "%" + filters.Search + "%"
	}

	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			count(distinct mg.*)
		from
			multicast_group mg
		inner join service_profile sp
			on sp.service_profile_id = mg.service_profile_id
		inner join organization o
			on o.id = sp.organization_id
		left join device_multicast_group dmg
			on mg.id = dmg.multicast_group_id
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

// GetMulticastGroups returns a slice of multicast-groups, given the privded
// filters. Note that empty values are not used as filters.
func GetMulticastGroups(db sqlx.Queryer, filters MulticastGroupFilters) ([]MulticastGroupListItem, error) {
	if filters.Search != "" {
		filters.Search = "%" + filters.Search + "%"
	}

	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			distinct mg.id,
			mg.created_at,
			mg.updated_at,
			mg.name,
			mg.service_profile_id,
			sp.name as service_profile_name
		from
			multicast_group mg
		inner join service_profile sp
			on sp.service_profile_id = mg.service_profile_id
		inner join organization o
			on o.id = sp.organization_id
		left join device_multicast_group dmg
			on mg.id = dmg.multicast_group_id
	`+filters.SQL()+`
		order by
			mg.name
		limit :limit
		offset :offset
	`, filters)
	if err != nil {
		return nil, errors.Wrap(err, "named query error")
	}

	var mgs []MulticastGroupListItem
	err = sqlx.Select(db, &mgs, query, args...)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return mgs, nil
}

// AddDeviceToMulticastGroup adds the given device to the given multicast-group.
// It is recommended that db is a transaction.
func AddDeviceToMulticastGroup(db sqlx.Ext, multicastGroupID uuid.UUID, devEUI lorawan.EUI64) error {
	_, err := db.Exec(`
		insert into device_multicast_group (
			dev_eui,
			multicast_group_id,
			created_at
		) values ($1, $2, $3)
	`, devEUI, multicastGroupID, time.Now())
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	nsClient, err := getNSClientForMulticastGroup(db, multicastGroupID)
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.AddDeviceToMulticastGroup(context.Background(), &ns.AddDeviceToMulticastGroupRequest{
		DevEui:           devEUI[:],
		MulticastGroupId: multicastGroupID.Bytes(),
	})
	if err != nil {
		return handleGrpcError(err, "add device to multicast-group error")
	}

	log.WithFields(log.Fields{
		"dev_eui":            devEUI,
		"multicast_group_id": multicastGroupID,
	}).Info("device added to multicast-group")

	return nil
}

// RemoveDeviceFromMulticastGroup removes the given device from the given
// multicast-group.
func RemoveDeviceFromMulticastGroup(db sqlx.Ext, multicastGroupID uuid.UUID, devEUI lorawan.EUI64) error {
	res, err := db.Exec(`
		delete from
			device_multicast_group
		where
			dev_eui = $1
			and multicast_group_id = $2
	`, devEUI, multicastGroupID)
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

	nsClient, err := getNSClientForMulticastGroup(db, multicastGroupID)
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	_, err = nsClient.RemoveDeviceFromMulticastGroup(context.Background(), &ns.RemoveDeviceFromMulticastGroupRequest{
		DevEui:           devEUI[:],
		MulticastGroupId: multicastGroupID.Bytes(),
	})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return handleGrpcError(err, "remove device from multicast-group error")
	}

	log.WithFields(log.Fields{
		"dev_eui":            devEUI,
		"multicast_group_id": multicastGroupID,
	}).Info("Device removed from multicast-group")

	return nil
}

// GetDeviceCountForMulticastGroup returns the number of devices for the given
// multicast-group.
func GetDeviceCountForMulticastGroup(db sqlx.Queryer, multicastGroup uuid.UUID) (int, error) {
	var count int

	err := sqlx.Get(db, &count, `
		select
			count(*)
		from
			device_multicast_group
		where
			multicast_group_id = $1
	`, multicastGroup)
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}

	return count, nil
}

// GetDevicesForMulticastGroup returns a slice of devices for the given
// multicast-group.
func GetDevicesForMulticastGroup(db sqlx.Queryer, multicastGroupID uuid.UUID, limit, offset int) ([]DeviceListItem, error) {
	var devices []DeviceListItem

	err := sqlx.Select(db, &devices, `
		select
			d.*,
			dp.name as device_profile_name
		from
			device d
		inner join device_profile dp
			on dp.device_profile_id = d.device_profile_id
		inner join device_multicast_group dmg
			on dmg.dev_eui = d.dev_eui
		where
			dmg.multicast_group_id = $1
		order by
			d.name
		limit $2
		offset $3
	`, multicastGroupID, limit, offset)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return devices, nil
}
