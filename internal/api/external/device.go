package external

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq/hstore"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/api/external/auth"
	"github.com/brocaar/chirpstack-application-server/internal/api/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/eventlog"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

// DeviceAPI exports the Node related functions.
type DeviceAPI struct {
	validator auth.Validator
}

// NewDeviceAPI creates a new NodeAPI.
func NewDeviceAPI(validator auth.Validator) *DeviceAPI {
	return &DeviceAPI{
		validator: validator,
	}
}

// Create creates the given device.
func (a *DeviceAPI) Create(ctx context.Context, req *pb.CreateDeviceRequest) (*empty.Empty, error) {
	if req.Device == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "device must not be nil")
	}

	// decode DevEUI
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.Device.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// decode DeviceProfileID
	dpID, err := uuid.FromString(req.Device.DeviceProfileId)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// validate access
	if err := a.validator.Validate(ctx,
		auth.ValidateNodesAccess(req.Device.ApplicationId, auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// if Name is "", set it to the DevEUI
	if req.Device.Name == "" {
		req.Device.Name = req.Device.DevEui
	}

	// Validate that application and device-profile are under the same
	// organization ID.
	app, err := storage.GetApplication(ctx, storage.DB(), req.Device.ApplicationId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	dp, err := storage.GetDeviceProfile(ctx, storage.DB(), dpID, false, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	if app.OrganizationID != dp.OrganizationID {
		return nil, grpc.Errorf(codes.InvalidArgument, "device-profile and application must be under the same organization")
	}

	// Set Device struct.
	d := storage.Device{
		DevEUI:            devEUI,
		ApplicationID:     req.Device.ApplicationId,
		DeviceProfileID:   dpID,
		Name:              req.Device.Name,
		Description:       req.Device.Description,
		SkipFCntCheck:     req.Device.SkipFCntCheck,
		ReferenceAltitude: req.Device.ReferenceAltitude,
		IsDisabled:        req.Device.IsDisabled,
		Variables: hstore.Hstore{
			Map: make(map[string]sql.NullString),
		},
		Tags: hstore.Hstore{
			Map: make(map[string]sql.NullString),
		},
	}
	for k, v := range req.Device.Variables {
		d.Variables.Map[k] = sql.NullString{String: v, Valid: true}
	}
	for k, v := range req.Device.Tags {
		d.Tags.Map[k] = sql.NullString{String: v, Valid: true}
	}

	// A transaction is needed as:
	//  * A remote gRPC call is performed and in case of error, we want to
	//    rollback the transaction.
	//  * We want to lock the organization so that we can validate the
	//    max device count.
	err = storage.Transaction(func(tx sqlx.Ext) error {
		org, err := storage.GetOrganization(ctx, tx, app.OrganizationID, true)
		if err != nil {
			return err
		}

		// Validate max. device count when != 0.
		if org.MaxDeviceCount != 0 {
			count, err := storage.GetDeviceCount(ctx, tx, storage.DeviceFilters{OrganizationID: app.OrganizationID})
			if err != nil {
				return err
			}

			if count >= org.MaxDeviceCount {
				return storage.ErrOrganizationMaxDeviceCount
			}
		}

		return storage.CreateDevice(ctx, tx, &d)
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// Get returns the device matching the given DevEUI.
func (a *DeviceAPI) Get(ctx context.Context, req *pb.GetDeviceRequest) (*pb.GetDeviceResponse, error) {
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Read)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	d, err := storage.GetDevice(ctx, storage.DB(), eui, false, false)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp := pb.GetDeviceResponse{
		Device: &pb.Device{
			DevEui:            d.DevEUI.String(),
			Name:              d.Name,
			ApplicationId:     d.ApplicationID,
			Description:       d.Description,
			DeviceProfileId:   d.DeviceProfileID.String(),
			SkipFCntCheck:     d.SkipFCntCheck,
			ReferenceAltitude: d.ReferenceAltitude,
			IsDisabled:        d.IsDisabled,
			Variables:         make(map[string]string),
			Tags:              make(map[string]string),
		},

		DeviceStatusBattery: 256,
		DeviceStatusMargin:  256,
	}

	if d.DeviceStatusBattery != nil {
		resp.DeviceStatusBattery = uint32(*d.DeviceStatusBattery)
	}
	if d.DeviceStatusMargin != nil {
		resp.DeviceStatusMargin = int32(*d.DeviceStatusMargin)
	}
	if d.LastSeenAt != nil {
		resp.LastSeenAt, err = ptypes.TimestampProto(*d.LastSeenAt)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
	}

	if d.Latitude != nil && d.Longitude != nil && d.Altitude != nil {
		resp.Location = &common.Location{
			Latitude:  *d.Latitude,
			Longitude: *d.Longitude,
			Altitude:  *d.Altitude,
		}
	}

	for k, v := range d.Variables.Map {
		if v.Valid {
			resp.Device.Variables[k] = v.String
		}
	}

	for k, v := range d.Tags.Map {
		if v.Valid {
			resp.Device.Tags[k] = v.String
		}
	}

	return &resp, nil
}

// List lists the available applications.
func (a *DeviceAPI) List(ctx context.Context, req *pb.ListDeviceRequest) (*pb.ListDeviceResponse, error) {
	var err error
	var idFilter bool

	filters := storage.DeviceFilters{
		ApplicationID: req.ApplicationId,
		Search:        req.Search,
		Tags: hstore.Hstore{
			Map: make(map[string]sql.NullString),
		},
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}

	if req.MulticastGroupId != "" {
		filters.MulticastGroupID, err = uuid.FromString(req.MulticastGroupId)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
	}

	if req.ServiceProfileId != "" {
		filters.ServiceProfileID, err = uuid.FromString(req.ServiceProfileId)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
	}

	for k, v := range req.Tags {
		filters.Tags.Map[k] = sql.NullString{String: v, Valid: true}
	}

	if filters.ApplicationID != 0 {
		idFilter = true

		// validate that the client has access to the given application
		if err := a.validator.Validate(ctx,
			auth.ValidateApplicationAccess(req.ApplicationId, auth.Read),
		); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
		}

	}

	if filters.MulticastGroupID != uuid.Nil {
		idFilter = true

		// validate that the client has access to the given multicast-group
		if err := a.validator.Validate(ctx,
			auth.ValidateMulticastGroupAccess(auth.Read, filters.MulticastGroupID),
		); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
		}
	}

	if filters.ServiceProfileID != uuid.Nil {
		idFilter = true

		// validate that the client has access to the given service-profile
		if err := a.validator.Validate(ctx,
			auth.ValidateServiceProfileAccess(auth.Read, filters.ServiceProfileID),
		); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
		}
	}

	if !idFilter {
		user, err := a.validator.GetUser(ctx)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		if !user.IsAdmin {
			return nil, grpc.Errorf(codes.Unauthenticated, "client must be global admin for unfiltered request")
		}
	}

	count, err := storage.GetDeviceCount(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	devices, err := storage.GetDevices(ctx, storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return a.returnList(count, devices)
}

// Update updates the device matching the given DevEUI.
func (a *DeviceAPI) Update(ctx context.Context, req *pb.UpdateDeviceRequest) (*empty.Empty, error) {
	if req.Device == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "device must not be nil")
	}

	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.Device.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	dpID, err := uuid.FromString(req.Device.DeviceProfileId)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(devEUI, auth.Update)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app, err := storage.GetApplication(ctx, storage.DB(), req.Device.ApplicationId)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	dp, err := storage.GetDeviceProfile(ctx, storage.DB(), dpID, false, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if app.OrganizationID != dp.OrganizationID {
		return nil, grpc.Errorf(codes.InvalidArgument, "device-profile and application must be under the same organization")
	}

	err = storage.Transaction(func(tx sqlx.Ext) error {
		d, err := storage.GetDevice(ctx, tx, devEUI, true, false)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		// If the device is moved to a different application, validate that
		// the new application is assigned to the same service-profile.
		// This to guarantee that the new application is still on the same
		// network-server and is not assigned to a different organization.
		if req.Device.ApplicationId != d.ApplicationID {
			appOld, err := storage.GetApplication(ctx, tx, d.ApplicationID)
			if err != nil {
				return helpers.ErrToRPCError(err)
			}

			appNew, err := storage.GetApplication(ctx, tx, req.Device.ApplicationId)
			if err != nil {
				return helpers.ErrToRPCError(err)
			}

			if appOld.ServiceProfileID != appNew.ServiceProfileID {
				return grpc.Errorf(codes.InvalidArgument, "when moving a device from application A to B, both A and B must share the same service-profile")
			}
		}

		d.ApplicationID = req.Device.ApplicationId
		d.DeviceProfileID = dpID
		d.Name = req.Device.Name
		d.Description = req.Device.Description
		d.SkipFCntCheck = req.Device.SkipFCntCheck
		d.ReferenceAltitude = req.Device.ReferenceAltitude
		d.IsDisabled = req.Device.IsDisabled
		d.Variables = hstore.Hstore{
			Map: make(map[string]sql.NullString),
		}
		d.Tags = hstore.Hstore{
			Map: make(map[string]sql.NullString),
		}

		for k, v := range req.Device.Variables {
			d.Variables.Map[k] = sql.NullString{String: v, Valid: true}
		}

		for k, v := range req.Device.Tags {
			d.Tags.Map[k] = sql.NullString{String: v, Valid: true}
		}

		if err := storage.UpdateDevice(ctx, tx, &d, false); err != nil {
			return helpers.ErrToRPCError(err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// Delete deletes the node matching the given name.
func (a *DeviceAPI) Delete(ctx context.Context, req *pb.DeleteDeviceRequest) (*empty.Empty, error) {
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Delete)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// as this also performs a remote call to delete the node from the
	// network-server, wrap it in a transaction
	err := storage.Transaction(func(tx sqlx.Ext) error {
		return storage.DeleteDevice(ctx, tx, eui)
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// CreateKeys creates the given device-keys.
func (a *DeviceAPI) CreateKeys(ctx context.Context, req *pb.CreateDeviceKeysRequest) (*empty.Empty, error) {
	if req.DeviceKeys == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "device_keys must not be nil")
	}

	// appKey is not used for LoRaWAN 1.0
	var appKey lorawan.AES128Key
	if req.DeviceKeys.AppKey != "" {
		if err := appKey.UnmarshalText([]byte(req.DeviceKeys.AppKey)); err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
		}
	}

	// nwkKey
	var nwkKey lorawan.AES128Key
	if err := nwkKey.UnmarshalText([]byte(req.DeviceKeys.NwkKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// devEUI
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DeviceKeys.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.CreateDeviceKeys(ctx, storage.DB(), &storage.DeviceKeys{
		DevEUI: eui,
		NwkKey: nwkKey,
		AppKey: appKey,
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// GetKeys returns the device-keys for the given DevEUI.
func (a *DeviceAPI) GetKeys(ctx context.Context, req *pb.GetDeviceKeysRequest) (*pb.GetDeviceKeysResponse, error) {
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	dk, err := storage.GetDeviceKeys(ctx, storage.DB(), eui)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &pb.GetDeviceKeysResponse{
		DeviceKeys: &pb.DeviceKeys{
			DevEui: eui.String(),
			AppKey: dk.AppKey.String(),
			NwkKey: dk.NwkKey.String(),
		},
	}, nil
}

// UpdateKeys updates the device-keys.
func (a *DeviceAPI) UpdateKeys(ctx context.Context, req *pb.UpdateDeviceKeysRequest) (*empty.Empty, error) {
	if req.DeviceKeys == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "device_keys must not be nil")
	}

	var appKey lorawan.AES128Key
	// appKey is not used for LoRaWAN 1.0
	if req.DeviceKeys.AppKey != "" {
		if err := appKey.UnmarshalText([]byte(req.DeviceKeys.AppKey)); err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
		}
	}

	var nwkKey lorawan.AES128Key
	if err := nwkKey.UnmarshalText([]byte(req.DeviceKeys.NwkKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DeviceKeys.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	dk, err := storage.GetDeviceKeys(ctx, storage.DB(), eui)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	dk.NwkKey = nwkKey
	dk.AppKey = appKey

	err = storage.UpdateDeviceKeys(ctx, storage.DB(), &dk)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// DeleteKeys deletes the device-keys for the given DevEUI.
func (a *DeviceAPI) DeleteKeys(ctx context.Context, req *pb.DeleteDeviceKeysRequest) (*empty.Empty, error) {
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Delete),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	if err := storage.DeleteDeviceKeys(ctx, storage.DB(), eui); err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	return &empty.Empty{}, nil
}

// Deactivate de-activates the device.
func (a *DeviceAPI) Deactivate(ctx context.Context, req *pb.DeactivateDeviceRequest) (*empty.Empty, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(devEUI, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	d, err := storage.GetDevice(ctx, storage.DB(), devEUI, false, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	n, err := storage.GetNetworkServerForDevEUI(ctx, storage.DB(), d.DevEUI)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	_, _ = nsClient.DeactivateDevice(ctx, &ns.DeactivateDeviceRequest{
		DevEui: d.DevEUI[:],
	})

	return &empty.Empty{}, nil
}

// Activate activates the node (ABP only).
func (a *DeviceAPI) Activate(ctx context.Context, req *pb.ActivateDeviceRequest) (*empty.Empty, error) {
	if req.DeviceActivation == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "device_activation must not be nil")
	}

	var devAddr lorawan.DevAddr
	var devEUI lorawan.EUI64
	var appSKey lorawan.AES128Key
	var nwkSEncKey lorawan.AES128Key
	var sNwkSIntKey lorawan.AES128Key
	var fNwkSIntKey lorawan.AES128Key

	if err := devAddr.UnmarshalText([]byte(req.DeviceActivation.DevAddr)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devAddr: %s", err)
	}
	if err := devEUI.UnmarshalText([]byte(req.DeviceActivation.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}
	if err := appSKey.UnmarshalText([]byte(req.DeviceActivation.AppSKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "appSKey: %s", err)
	}
	if err := nwkSEncKey.UnmarshalText([]byte(req.DeviceActivation.NwkSEncKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "nwkSEncKey: %s", err)
	}
	if err := sNwkSIntKey.UnmarshalText([]byte(req.DeviceActivation.SNwkSIntKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "sNwkSIntKey: %s", err)
	}
	if err := fNwkSIntKey.UnmarshalText([]byte(req.DeviceActivation.FNwkSIntKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "fNwkSIntKey: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(devEUI, auth.Update)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	d, err := storage.GetDevice(ctx, storage.DB(), devEUI, false, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	n, err := storage.GetNetworkServerForDevEUI(ctx, storage.DB(), devEUI)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	_, _ = nsClient.DeactivateDevice(ctx, &ns.DeactivateDeviceRequest{
		DevEui: d.DevEUI[:],
	})

	actReq := ns.ActivateDeviceRequest{
		DeviceActivation: &ns.DeviceActivation{
			DevEui:      d.DevEUI[:],
			DevAddr:     devAddr[:],
			NwkSEncKey:  nwkSEncKey[:],
			SNwkSIntKey: sNwkSIntKey[:],
			FNwkSIntKey: fNwkSIntKey[:],
			FCntUp:      req.DeviceActivation.FCntUp,
			NFCntDown:   req.DeviceActivation.NFCntDown,
			AFCntDown:   req.DeviceActivation.AFCntDown,
		},
	}

	err = storage.Transaction(func(db sqlx.Ext) error {
		if err := storage.UpdateDeviceActivation(ctx, db, d.DevEUI, devAddr, appSKey); err != nil {
			return helpers.ErrToRPCError(err)
		}

		_, err := nsClient.ActivateDevice(ctx, &actReq)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"dev_addr": devAddr,
		"dev_eui":  d.DevEUI,
		"ctx_id":   ctx.Value(logging.ContextIDKey),
	}).Info("device activated")

	return &empty.Empty{}, nil
}

// GetActivation returns the device activation for the given DevEUI.
func (a *DeviceAPI) GetActivation(ctx context.Context, req *pb.GetDeviceActivationRequest) (*pb.GetDeviceActivationResponse, error) {
	var devAddr lorawan.DevAddr
	var devEUI lorawan.EUI64
	var sNwkSIntKey lorawan.AES128Key
	var fNwkSIntKey lorawan.AES128Key
	var nwkSEncKey lorawan.AES128Key

	if err := devEUI.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(devEUI, auth.Read)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	d, err := storage.GetDevice(ctx, storage.DB(), devEUI, false, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	n, err := storage.GetNetworkServerForDevEUI(ctx, storage.DB(), devEUI)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	devAct, err := nsClient.GetDeviceActivation(ctx, &ns.GetDeviceActivationRequest{
		DevEui: d.DevEUI[:],
	})
	if err != nil {
		return nil, err
	}

	copy(devAddr[:], devAct.DeviceActivation.DevAddr)
	copy(nwkSEncKey[:], devAct.DeviceActivation.NwkSEncKey)
	copy(sNwkSIntKey[:], devAct.DeviceActivation.SNwkSIntKey)
	copy(fNwkSIntKey[:], devAct.DeviceActivation.FNwkSIntKey)

	return &pb.GetDeviceActivationResponse{
		DeviceActivation: &pb.DeviceActivation{
			DevEui:      d.DevEUI.String(),
			DevAddr:     devAddr.String(),
			AppSKey:     d.AppSKey.String(),
			NwkSEncKey:  nwkSEncKey.String(),
			SNwkSIntKey: sNwkSIntKey.String(),
			FNwkSIntKey: fNwkSIntKey.String(),
			FCntUp:      devAct.DeviceActivation.FCntUp,
			NFCntDown:   devAct.DeviceActivation.NFCntDown,
			AFCntDown:   devAct.DeviceActivation.AFCntDown,
		},
	}, nil
}

// StreamFrameLogs streams the uplink and downlink frame-logs for the given DevEUI.
// Note: these are the raw LoRaWAN frames and this endpoint is intended for debugging.
func (a *DeviceAPI) StreamFrameLogs(req *pb.StreamDeviceFrameLogsRequest, srv pb.DeviceService_StreamFrameLogsServer) error {
	var devEUI lorawan.EUI64

	if err := devEUI.UnmarshalText([]byte(req.DevEui)); err != nil {
		return grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := a.validator.Validate(srv.Context(),
		auth.ValidateNodeAccess(devEUI, auth.Read)); err != nil {
		return grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	n, err := storage.GetNetworkServerForDevEUI(srv.Context(), storage.DB(), devEUI)
	if err != nil {
		return helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return helpers.ErrToRPCError(err)
	}

	streamClient, err := nsClient.StreamFrameLogsForDevice(srv.Context(), &ns.StreamFrameLogsForDeviceRequest{
		DevEui: devEUI[:],
	})
	if err != nil {
		return err
	}

	for {
		resp, err := streamClient.Recv()
		if err != nil {
			if grpc.Code(err) == codes.Canceled {
				return nil
			}

			return err
		}

		up, down, err := convertUplinkAndDownlinkFrames(resp.GetUplinkFrameSet(), resp.GetDownlinkFrame(), true)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		var frameResp pb.StreamDeviceFrameLogsResponse
		if up != nil {
			up.PublishedAt = resp.GetUplinkFrameSet().GetPublishedAt()
			frameResp.Frame = &pb.StreamDeviceFrameLogsResponse_UplinkFrame{
				UplinkFrame: up,
			}
		}

		if down != nil {
			down.PublishedAt = resp.GetDownlinkFrame().GetPublishedAt()
			frameResp.Frame = &pb.StreamDeviceFrameLogsResponse_DownlinkFrame{
				DownlinkFrame: down,
			}
		}

		err = srv.Send(&frameResp)
		if err != nil {
			return err
		}
	}
}

// StreamEventLogs stream the device events (uplink payloads, ACKs, joins, errors).
// Note: this endpoint is intended for debugging and should not be used for building
// integrations.
func (a *DeviceAPI) StreamEventLogs(req *pb.StreamDeviceEventLogsRequest, srv pb.DeviceService_StreamEventLogsServer) error {
	var devEUI lorawan.EUI64

	if err := devEUI.UnmarshalText([]byte(req.DevEui)); err != nil {
		return grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := a.validator.Validate(srv.Context(),
		auth.ValidateNodeAccess(devEUI, auth.Read)); err != nil {
		return grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	eventLogChan := make(chan eventlog.EventLog)
	go func() {
		err := eventlog.GetEventLogForDevice(srv.Context(), devEUI, eventLogChan)
		if err != nil {
			log.WithError(err).Error("get event-log for device error")
		}
		close(eventLogChan)
	}()

	for el := range eventLogChan {
		b, err := json.Marshal(el.Payload)
		if err != nil {
			return grpc.Errorf(codes.Internal, "marshal json error: %s", err)
		}

		resp := pb.StreamDeviceEventLogsResponse{
			PublishedAt: el.PublishedAt,
			Type:        el.Type,
			PayloadJson: string(b),
			StreamId:    el.StreamID,
		}

		err = srv.Send(&resp)
		if err != nil {
			log.WithError(err).Error("error sending event-log response")
		}
	}

	return nil
}

// GetRandomDevAddr returns a random DevAddr taking the NwkID prefix into account.
func (a *DeviceAPI) GetRandomDevAddr(ctx context.Context, req *pb.GetRandomDevAddrRequest) (*pb.GetRandomDevAddrResponse, error) {
	var devEUI lorawan.EUI64

	if err := devEUI.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	n, err := storage.GetNetworkServerForDevEUI(ctx, storage.DB(), devEUI)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp, err := nsClient.GetRandomDevAddr(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	var devAddr lorawan.DevAddr
	copy(devAddr[:], resp.DevAddr)

	return &pb.GetRandomDevAddrResponse{
		DevAddr: devAddr.String(),
	}, nil
}

// GetStats returns the device statistics.
func (a *DeviceAPI) GetStats(ctx context.Context, req *pb.GetDeviceStatsRequest) (*pb.GetDeviceStatsResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad dev_eui: %s", err)
	}

	err := a.validator.Validate(ctx, auth.ValidateNodeAccess(devEUI, auth.Read))
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	start, err := ptypes.Timestamp(req.StartTimestamp)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	end, err := ptypes.Timestamp(req.EndTimestamp)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	_, ok := ns.AggregationInterval_value[strings.ToUpper(req.Interval)]
	if !ok {
		return nil, grpc.Errorf(codes.InvalidArgument, "bad interval: %s", req.Interval)
	}

	metrics, err := storage.GetMetrics(ctx, storage.AggregationInterval(strings.ToUpper(req.Interval)), "device:"+devEUI.String(), start, end)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	result := make([]*pb.DeviceStats, len(metrics))
	for i, m := range metrics {
		result[i] = &pb.DeviceStats{
			RxPackets:             uint32(m.Metrics["rx_count"]),
			RxPacketsPerFrequency: make(map[uint32]uint32),
			RxPacketsPerDr:        make(map[uint32]uint32),
			Errors:                make(map[string]uint32),
		}

		result[i].Timestamp, err = ptypes.TimestampProto(m.Time)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		if (m.Metrics["rx_count"]) != 0 {
			result[i].GwRssi = float32(m.Metrics["gw_rssi_sum"] / m.Metrics["rx_count"])
			result[i].GwSnr = float32(m.Metrics["gw_snr_sum"] / m.Metrics["rx_count"])
		}

		for k, v := range m.Metrics {
			if strings.HasPrefix(k, "rx_freq_") {
				if freq, err := strconv.ParseUint(strings.TrimPrefix(k, "rx_freq_"), 10, 32); err == nil {
					result[i].RxPacketsPerFrequency[uint32(freq)] = uint32(v)
				}
			}

			if strings.HasPrefix(k, "rx_dr_") {
				if freq, err := strconv.ParseUint(strings.TrimPrefix(k, "rx_dr_"), 10, 32); err == nil {
					result[i].RxPacketsPerDr[uint32(freq)] = uint32(v)
				}
			}

			if strings.HasPrefix(k, "error_") {
				e := strings.TrimPrefix(k, "error_")
				result[i].Errors[e] = uint32(v)
			}
		}
	}

	return &pb.GetDeviceStatsResponse{
		Result: result,
	}, nil
}

// ClearDeviceNonces deletes the device older activation records for the given DevEUI.
// TODO: These are clear older DevNonce records from device activation records in Network Server
// TODO: These clears all DevNonce records but keeps latest 20 records for maintain device activation status
func (a *DeviceAPI) ClearDeviceNonces(ctx context.Context, req *pb.ClearDeviceNoncesRequest) (*empty.Empty, error) {
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Update)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	d, err := storage.GetDevice(ctx, storage.DB(), eui, false, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	n, err := storage.GetNetworkServerForDevEUI(ctx, storage.DB(), d.DevEUI)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	_, _ = nsClient.ClearDeviceNonces(ctx, &ns.ClearDeviceNoncesRequest{
		DevEui: d.DevEUI[:],
	})

	return &empty.Empty{}, nil
}

func (a *DeviceAPI) returnList(count int, devices []storage.DeviceListItem) (*pb.ListDeviceResponse, error) {
	resp := pb.ListDeviceResponse{
		TotalCount: int64(count),
	}
	for _, device := range devices {
		item := pb.DeviceListItem{
			DevEui:                          device.DevEUI.String(),
			Name:                            device.Name,
			Description:                     device.Description,
			ApplicationId:                   device.ApplicationID,
			DeviceProfileId:                 device.DeviceProfileID.String(),
			DeviceProfileName:               device.DeviceProfileName,
			DeviceStatusBattery:             256,
			DeviceStatusMargin:              256,
			DeviceStatusExternalPowerSource: device.DeviceStatusExternalPower,
		}

		if !device.DeviceStatusExternalPower && device.DeviceStatusBattery == nil {
			item.DeviceStatusBattery = 255
			item.DeviceStatusBatteryLevelUnavailable = true
		}

		if device.DeviceStatusExternalPower {
			item.DeviceStatusBattery = 0
		}

		if device.DeviceStatusBattery != nil {
			item.DeviceStatusBattery = uint32(254 / *device.DeviceStatusBattery * 100)
			item.DeviceStatusBatteryLevel = *device.DeviceStatusBattery
		}

		if device.DeviceStatusBattery != nil {
			item.DeviceStatusBattery = uint32(*device.DeviceStatusBattery)
		}
		if device.DeviceStatusMargin != nil {
			item.DeviceStatusMargin = int32(*device.DeviceStatusMargin)
		}
		if device.LastSeenAt != nil {
			var err error
			item.LastSeenAt, err = ptypes.TimestampProto(*device.LastSeenAt)
			if err != nil {
				return nil, helpers.ErrToRPCError(err)
			}
		}

		resp.Result = append(resp.Result, &item)
	}
	return &resp, nil
}

func convertUplinkAndDownlinkFrames(up *ns.UplinkFrameLog, down *ns.DownlinkFrameLog, decodeMACCommands bool) (*pb.UplinkFrameLog, *pb.DownlinkFrameLog, error) {
	var phy lorawan.PHYPayload

	if up != nil {
		if err := phy.UnmarshalBinary(up.PhyPayload); err != nil {
			return nil, nil, errors.Wrap(err, "unmarshal phypayload error")
		}
	}

	if down != nil {
		if err := phy.UnmarshalBinary(down.PhyPayload); err != nil {
			return nil, nil, errors.Wrap(err, "unmarshal phypayload error")
		}
	}

	if decodeMACCommands {
		switch v := phy.MACPayload.(type) {
		case *lorawan.MACPayload:
			if err := phy.DecodeFOptsToMACCommands(); err != nil {
				return nil, nil, errors.Wrap(err, "decode fopts to mac-commands error")
			}

			if v.FPort != nil && *v.FPort == 0 {
				if err := phy.DecodeFRMPayloadToMACCommands(); err != nil {
					return nil, nil, errors.Wrap(err, "decode frmpayload to mac-commands error")
				}
			}
		}
	}

	phyJSON, err := json.Marshal(phy)
	if err != nil {
		return nil, nil, errors.Wrap(err, "marshal phypayload error")
	}

	if up != nil {
		uplinkFrameLog := pb.UplinkFrameLog{
			TxInfo:         up.TxInfo,
			RxInfo:         up.RxInfo,
			PhyPayloadJson: string(phyJSON),
		}

		return &uplinkFrameLog, nil, nil
	}

	if down != nil {
		var gatewayID lorawan.EUI64
		copy(gatewayID[:], down.GatewayId)

		downlinkFrameLog := pb.DownlinkFrameLog{
			TxInfo:         down.TxInfo,
			PhyPayloadJson: string(phyJSON),
			GatewayId:      gatewayID.String(),
		}

		return nil, &downlinkFrameLog, nil
	}

	return nil, nil, nil
}
