package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

// FUOTADeploymentState defines the fuota deployment state.
type FUOTADeploymentState string

// FUOTA deployment states.
const (
	FUOTADeploymentMulticastCreate        FUOTADeploymentState = "MC_CREATE"
	FUOTADeploymentMulticastSetup         FUOTADeploymentState = "MC_SETUP"
	FUOTADeploymentFragmentationSessSetup FUOTADeploymentState = "FRAG_SESS_SETUP"
	FUOTADeploymentMulticastSessCSetup    FUOTADeploymentState = "MC_SESS_C_SETUP"
	FUOTADeploymentEnqueue                FUOTADeploymentState = "ENQUEUE"
	FUOTADeploymentStatusRequest          FUOTADeploymentState = "STATUS_REQUEST"
	FUOTADeploymentSetDeviceStatus        FUOTADeploymentState = "SET_DEVICE_STATUS"
	FUOTADeploymentCleanup                FUOTADeploymentState = "CLEANUP"
	FUOTADeploymentDone                   FUOTADeploymentState = "DONE"
)

// FUOTADeploymentDeviceState defines the fuota deployment device state.
type FUOTADeploymentDeviceState string

// FUOTA deployment device states.
const (
	FUOTADeploymentDevicePending FUOTADeploymentDeviceState = "PENDING"
	FUOTADeploymentDeviceSuccess FUOTADeploymentDeviceState = "SUCCESS"
	FUOTADeploymentDeviceError   FUOTADeploymentDeviceState = "ERROR"
)

// FUOTADeploymentGroupType defines the group-type.
type FUOTADeploymentGroupType string

// FUOTA deployment group types.
const (
	FUOTADeploymentGroupTypeB FUOTADeploymentGroupType = "B"
	FUOTADeploymentGroupTypeC FUOTADeploymentGroupType = "C"
)

// FUOTADeployment defiles a firmware update over the air deployment.
type FUOTADeployment struct {
	ID                  uuid.UUID                `db:"id"`
	CreatedAt           time.Time                `db:"created_at"`
	UpdatedAt           time.Time                `db:"updated_at"`
	Name                string                   `db:"name"`
	MulticastGroupID    *uuid.UUID               `db:"multicast_group_id"`
	GroupType           FUOTADeploymentGroupType `db:"group_type"`
	DR                  int                      `db:"dr"`
	Frequency           int                      `db:"frequency"`
	PingSlotPeriod      int                      `db:"ping_slot_period"`
	FragmentationMatrix uint8                    `db:"fragmentation_matrix"`
	Descriptor          [4]byte                  `db:"descriptor"`
	Payload             []byte                   `db:"payload"`
	FragSize            int                      `db:"frag_size"`
	Redundancy          int                      `db:"redundancy"`
	BlockAckDelay       int                      `db:"block_ack_delay"`
	MulticastTimeout    int                      `db:"multicast_timeout"`
	State               FUOTADeploymentState     `db:"state"`
	UnicastTimeout      time.Duration            `db:"unicast_timeout"`
	NextStepAfter       time.Time                `db:"next_step_after"`
}

// FUOTADeploymentListItem defines a FUOTA deployment item for listing.
type FUOTADeploymentListItem struct {
	ID            uuid.UUID            `db:"id"`
	CreatedAt     time.Time            `db:"created_at"`
	UpdatedAt     time.Time            `db:"updated_at"`
	Name          string               `db:"name"`
	State         FUOTADeploymentState `db:"state"`
	NextStepAfter time.Time            `db:"next_step_after"`
}

// FUOTADeploymentDevice defines the device record of a FUOTA deployment.
type FUOTADeploymentDevice struct {
	FUOTADeploymentID uuid.UUID                  `db:"fuota_deployment_id"`
	DevEUI            lorawan.EUI64              `db:"dev_eui"`
	CreatedAt         time.Time                  `db:"created_at"`
	UpdatedAt         time.Time                  `db:"updated_at"`
	State             FUOTADeploymentDeviceState `db:"state"`
	ErrorMessage      string                     `db:"error_message"`
}

// FUOTADeploymentDeviceListItem defines the Device as FUOTA deployment list item.
type FUOTADeploymentDeviceListItem struct {
	CreatedAt         time.Time                  `db:"created_at"`
	UpdatedAt         time.Time                  `db:"updated_at"`
	FUOTADeploymentID uuid.UUID                  `db:"fuota_deployment_id"`
	DevEUI            lorawan.EUI64              `db:"dev_eui"`
	DeviceName        string                     `db:"device_name"`
	State             FUOTADeploymentDeviceState `db:"state"`
	ErrorMessage      string                     `db:"error_message"`
}

// FUOTADeploymentFilter provides filters that can be used to filter on
// FUOTA deployments. Note that empty values are not used as filters.
type FUOTADeploymentFilters struct {
	DevEUI        lorawan.EUI64 `db:"dev_eui"`
	ApplicationID int64         `db:"application_id"`

	// Limit and Offset are added for convenience so that this struct can
	// be given as the arguments.
	Limit  int `db:"limit"`
	Offset int `db:"offset"`
}

// SQL returns the SQL filter.
func (f FUOTADeploymentFilters) SQL() string {
	var filters []string
	var nullDevEUI lorawan.EUI64

	if f.DevEUI != nullDevEUI {
		filters = append(filters, "fdd.dev_eui = :dev_eui")
	}

	if f.ApplicationID != 0 {
		filters = append(filters, "d.application_id = :application_id")
	}

	if len(filters) == 0 {
		return ""
	}

	return "where " + strings.Join(filters, " and ")
}

// CreateFUOTADeploymentForDevice creates and initializes a FUOTA deployment
// for the given device.
func CreateFUOTADeploymentForDevice(ctx context.Context, db sqlx.Ext, fd *FUOTADeployment, devEUI lorawan.EUI64) error {
	now := time.Now()
	var err error
	fd.ID, err = uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "new uuid error")
	}

	fd.CreatedAt = now
	fd.UpdatedAt = now
	fd.NextStepAfter = now
	if fd.State == "" {
		fd.State = FUOTADeploymentMulticastCreate
	}

	_, err = db.Exec(`
		insert into fuota_deployment (
			id,
			created_at,
			updated_at,
			name,
			multicast_group_id,

			fragmentation_matrix,
			descriptor,
			payload,
			state,
			next_step_after,
			unicast_timeout,
			frag_size,
			redundancy,
			block_ack_delay,
			multicast_timeout,
			group_type,
			dr,
			frequency,
			ping_slot_period
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		fd.ID,
		fd.CreatedAt,
		fd.UpdatedAt,
		fd.Name,
		fd.MulticastGroupID,
		[]byte{fd.FragmentationMatrix},
		fd.Descriptor[:],
		fd.Payload,
		fd.State,
		fd.NextStepAfter,
		fd.UnicastTimeout,
		fd.FragSize,
		fd.Redundancy,
		fd.BlockAckDelay,
		fd.MulticastTimeout,
		fd.GroupType,
		fd.DR,
		fd.Frequency,
		fd.PingSlotPeriod,
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	_, err = db.Exec(`
		insert into fuota_deployment_device (
			fuota_deployment_id,
			dev_eui,
			created_at,
			updated_at,
			state,
			error_message
		) values ($1, $2, $3, $4, $5, $6)`,
		fd.ID,
		devEUI,
		now,
		now,
		FUOTADeploymentDevicePending,
		"",
	)
	if err != nil {
		return handlePSQLError(Insert, err, "insert error")
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"id":      fd.ID,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("fuota deploymented created for device")

	return nil
}

// GetFUOTADeployment returns the FUOTA deployment for the given ID.
func GetFUOTADeployment(ctx context.Context, db sqlx.Ext, id uuid.UUID, forUpdate bool) (FUOTADeployment, error) {
	var fu string
	if forUpdate {
		fu = " for update"
	}

	row := db.QueryRowx(`
		select
			id,
			created_at,
			updated_at,
			name,
			multicast_group_id,
			fragmentation_matrix,
			descriptor,
			payload,
			state,
			next_step_after,
			unicast_timeout,
			frag_size,
			redundancy,
			block_ack_delay,
			multicast_timeout,
			group_type,
			dr,
			frequency,
			ping_slot_period
		from
			fuota_deployment
		where
			id = $1`+fu,
		id,
	)

	return scanFUOTADeployment(row)
}

// GetPendingFUOTADeployments returns the pending FUOTA deployments.
func GetPendingFUOTADeployments(ctx context.Context, db sqlx.Ext, batchSize int) ([]FUOTADeployment, error) {
	var out []FUOTADeployment

	rows, err := db.Queryx(`
		select
			id,
			created_at,
			updated_at,
			name,
			multicast_group_id,
			fragmentation_matrix,
			descriptor,
			payload,
			state,
			next_step_after,
			unicast_timeout,
			frag_size,
			redundancy,
			block_ack_delay,
			multicast_timeout,
			group_type,
			dr,
			frequency,
			ping_slot_period
		from
			fuota_deployment
		where
			state != $1
			and next_step_after <= $2
		limit $3
		for update
		skip locked`,
		FUOTADeploymentDone,
		time.Now(),
		batchSize,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}
	defer rows.Close()

	for rows.Next() {
		item, err := scanFUOTADeployment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	return out, nil
}

// UpdateFUOTADeployment updates the given FUOTA deployment.
func UpdateFUOTADeployment(ctx context.Context, db sqlx.Ext, fd *FUOTADeployment) error {
	fd.UpdatedAt = time.Now()

	res, err := db.Exec(`
		update fuota_deployment
		set
			updated_at = $2,
			name = $3,
			multicast_group_id = $4,
			fragmentation_matrix = $5,
			descriptor = $6,
			payload = $7,
			state = $8,
			next_step_after = $9,
			unicast_timeout = $10,
			frag_size = $11,
			redundancy = $12,
			block_ack_delay = $13,
			multicast_timeout = $14,
			group_type = $15,
			dr = $16,
			frequency = $17,
			ping_slot_period = $18
		where
			id = $1`,
		fd.ID,
		fd.UpdatedAt,
		fd.Name,
		fd.MulticastGroupID,
		[]byte{fd.FragmentationMatrix},
		fd.Descriptor[:],
		fd.Payload,
		fd.State,
		fd.NextStepAfter,
		fd.UnicastTimeout,
		fd.FragSize,
		fd.Redundancy,
		fd.BlockAckDelay,
		fd.MulticastTimeout,
		fd.GroupType,
		fd.DR,
		fd.Frequency,
		fd.PingSlotPeriod,
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
		"id":     fd.ID,
		"state":  fd.State,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("fuota deployment updated")

	return nil
}

// GetFUOTADeploymentCount returns the number of FUOTA deployments.
func GetFUOTADeploymentCount(ctx context.Context, db sqlx.Queryer, filters FUOTADeploymentFilters) (int, error) {
	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			count(distinct fd.*)
		from
			fuota_deployment fd
		inner join
			fuota_deployment_device fdd
		on
			fd.id = fdd.fuota_deployment_id
		inner join
			device d
		on
			fdd.dev_eui = d.dev_eui
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

// GetFUOTADeployments returns a slice of fuota deployments.
func GetFUOTADeployments(ctx context.Context, db sqlx.Queryer, filters FUOTADeploymentFilters) ([]FUOTADeploymentListItem, error) {
	query, args, err := sqlx.BindNamed(sqlx.DOLLAR, `
		select
			distinct fd.id,
			fd.created_at,
			fd.updated_at,
			fd.name,
			fd.state,
			fd.next_step_after
		from
			fuota_deployment fd
		inner join
			fuota_deployment_device fdd
		on
			fd.id = fdd.fuota_deployment_id
		inner join
			device d
		on
			fdd.dev_eui = d.dev_eui
	`+filters.SQL()+`
	order by
		fd.created_at desc
	limit :limit
	offset :offset
	`, filters)
	if err != nil {
		return nil, errors.Wrap(err, "named query error")
	}

	var items []FUOTADeploymentListItem
	if err = sqlx.Select(db, &items, query, args...); err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return items, nil
}

// GetFUOTADeploymentDevice returns the FUOTA deployment record for the given
// device.
func GetFUOTADeploymentDevice(ctx context.Context, db sqlx.Queryer, fuotaDeploymentID uuid.UUID, devEUI lorawan.EUI64) (FUOTADeploymentDevice, error) {
	var out FUOTADeploymentDevice
	err := sqlx.Get(db, &out, `
		select
			*
		from
			fuota_deployment_device
		where
			fuota_deployment_id = $1
			and dev_eui = $2`,
		fuotaDeploymentID,
		devEUI,
	)
	if err != nil {
		return out, handlePSQLError(Select, err, "select error")
	}
	return out, nil
}

// GetPendingFUOTADeploymentDevice returns the pending FUOTA deployment record
// for the given DevEUI.
func GetPendingFUOTADeploymentDevice(ctx context.Context, db sqlx.Queryer, devEUI lorawan.EUI64) (FUOTADeploymentDevice, error) {
	var out FUOTADeploymentDevice

	err := sqlx.Get(db, &out, `
		select
			*
		from
			fuota_deployment_device
		where
			dev_eui = $1
			and state = $2`,
		devEUI,
		FUOTADeploymentDevicePending,
	)
	if err != nil {
		return out, handlePSQLError(Select, err, "select error")
	}

	return out, nil
}

// UpdateFUOTADeploymentDevice updates the given fuota deployment device record.
func UpdateFUOTADeploymentDevice(ctx context.Context, db sqlx.Ext, fdd *FUOTADeploymentDevice) error {
	fdd.UpdatedAt = time.Now()

	res, err := db.Exec(`
		update
			fuota_deployment_device
		set
			updated_at = $3,
			state = $4,
			error_message = $5
		where
			dev_eui = $1
			and fuota_deployment_id = $2`,
		fdd.DevEUI,
		fdd.FUOTADeploymentID,
		fdd.UpdatedAt,
		fdd.State,
		fdd.ErrorMessage,
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
		"dev_eui":             fdd.DevEUI,
		"fuota_deployment_id": fdd.FUOTADeploymentID,
		"state":               fdd.State,
		"ctx_id":              ctx.Value(logging.ContextIDKey),
	}).Info("fuota deployment device updated")

	return nil
}

// GetFUOTADeploymentDeviceCount returns the device count for the given
// FUOTA deployment ID.
func GetFUOTADeploymentDeviceCount(ctx context.Context, db sqlx.Queryer, fuotaDeploymentID uuid.UUID) (int, error) {
	var count int
	err := sqlx.Get(db, &count, `
		select
			count(*)
		from
			fuota_deployment_device
		where
			fuota_deployment_id = $1`,
		fuotaDeploymentID,
	)
	if err != nil {
		return 0, handlePSQLError(Select, err, "select error")
	}

	return count, nil
}

// GetFUOTADeploymentDevices returns a slice of devices for the given FUOTA
// deployment ID.
func GetFUOTADeploymentDevices(ctx context.Context, db sqlx.Queryer, fuotaDeploymentID uuid.UUID, limit, offset int) ([]FUOTADeploymentDeviceListItem, error) {
	var out []FUOTADeploymentDeviceListItem

	err := sqlx.Select(db, &out, `
		select
			dd.created_at,
			dd.updated_at,
			dd.fuota_deployment_id,
			dd.dev_eui,
			d.name as device_name,
			dd.state,
			dd.error_message
		from
			fuota_deployment_device dd
		inner join
			device d
			on dd.dev_eui = d.dev_eui
		where
			dd.fuota_deployment_id = $3
		order by
			d.Name
		limit $1
		offset $2`,
		limit,
		offset,
		fuotaDeploymentID,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return out, nil
}

// GetServiceProfileIDForFUOTADeployment returns the service-profile ID for the given FUOTA deployment.
func GetServiceProfileIDForFUOTADeployment(ctx context.Context, db sqlx.Ext, fuotaDeploymentID uuid.UUID) (uuid.UUID, error) {
	var out uuid.UUID

	err := sqlx.Get(db, &out, `
		select
			a.service_profile_id
		from
			fuota_deployment_device fdd
		inner join
			device d
		on
			d.dev_eui = fdd.dev_eui
		inner join
			application a
		on
			a.id = d.application_id
		where
			fdd.fuota_deployment_id = $1
		limit 1`,
		fuotaDeploymentID,
	)
	if err != nil {
		return out, handlePSQLError(Select, err, "select error")
	}

	return out, nil
}

func scanFUOTADeployment(row sqlx.ColScanner) (FUOTADeployment, error) {
	var fd FUOTADeployment

	var fragmentationMatrix []byte
	var descriptor []byte

	err := row.Scan(
		&fd.ID,
		&fd.CreatedAt,
		&fd.UpdatedAt,
		&fd.Name,
		&fd.MulticastGroupID,
		&fragmentationMatrix,
		&descriptor,
		&fd.Payload,
		&fd.State,
		&fd.NextStepAfter,
		&fd.UnicastTimeout,
		&fd.FragSize,
		&fd.Redundancy,
		&fd.BlockAckDelay,
		&fd.MulticastTimeout,
		&fd.GroupType,
		&fd.DR,
		&fd.Frequency,
		&fd.PingSlotPeriod,
	)
	if err != nil {
		return fd, handlePSQLError(Select, err, "select error")
	}

	if len(fragmentationMatrix) != 1 {
		return fd, fmt.Errorf("FragmentationMatrix must have length 1, got: %d", len(fragmentationMatrix))
	}
	fd.FragmentationMatrix = fragmentationMatrix[0]

	if len(descriptor) != len(fd.Descriptor) {
		return fd, fmt.Errorf("Descriptor must have length: %d, got: %d", len(fd.Descriptor), len(descriptor))
	}
	copy(fd.Descriptor[:], descriptor)

	return fd, nil
}
