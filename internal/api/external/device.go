package external

import (
	"encoding/hex"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/external/auth"
	"github.com/brocaar/lora-app-server/internal/api/helpers"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/eventlog"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/common"
	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/loraserver/api/ns"
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

	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.Device.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	dpID, err := uuid.FromString(req.Device.DeviceProfileId)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodesAccess(req.Device.ApplicationId, auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// if Name is "", set it to the DevEUI
	if req.Device.Name == "" {
		req.Device.Name = req.Device.DevEui
	}

	d := storage.Device{
		DevEUI:            devEUI,
		ApplicationID:     req.Device.ApplicationId,
		DeviceProfileID:   dpID,
		Name:              req.Device.Name,
		Description:       req.Device.Description,
		SkipFCntCheck:     req.Device.SkipFCntCheck,
		ReferenceAltitude: req.Device.ReferenceAltitude,
	}

	// as this also performs a remote call to create the node on the
	// network-server, wrap it in a transaction
	err = storage.Transaction(func(tx sqlx.Ext) error {
		return storage.CreateDevice(tx, &d)
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

	d, err := storage.GetDevice(storage.DB(), eui, false, false)
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
			Source:    common.LocationSource_GEO_RESOLVER,
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
		Limit:         int(req.Limit),
		Offset:        int(req.Offset),
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
		isAdmin, err := a.validator.GetIsAdmin(ctx)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
		}

		if !isAdmin {
			return nil, grpc.Errorf(codes.Unauthenticated, "client must be global admin for unfiltered request")
		}
	}

	count, err := storage.GetDeviceCount(storage.DB(), filters)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	devices, err := storage.GetDevices(storage.DB(), filters)
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

	err = storage.Transaction(func(tx sqlx.Ext) error {
		d, err := storage.GetDevice(tx, devEUI, true, false)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		// If the device is moved to a different application, validate that
		// the new application is assigned to the same service-profile.
		// This to guarantee that the new application is still on the same
		// network-server and is not assigned to a different organization.
		if req.Device.ApplicationId != d.ApplicationID {
			appOld, err := storage.GetApplication(tx, d.ApplicationID)
			if err != nil {
				return helpers.ErrToRPCError(err)
			}

			appNew, err := storage.GetApplication(tx, req.Device.ApplicationId)
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

		if err := storage.UpdateDevice(tx, &d, false); err != nil {
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
		return storage.DeleteDevice(tx, eui)
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

	err := storage.CreateDeviceKeys(storage.DB(), &storage.DeviceKeys{
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

	dk, err := storage.GetDeviceKeys(storage.DB(), eui)
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

	dk, err := storage.GetDeviceKeys(storage.DB(), eui)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}
	dk.NwkKey = nwkKey
	dk.AppKey = appKey

	err = storage.UpdateDeviceKeys(storage.DB(), &dk)
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

	if err := storage.DeleteDeviceKeys(storage.DB(), eui); err != nil {
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

	d, err := storage.GetDevice(storage.DB(), devEUI, false, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	n, err := storage.GetNetworkServerForDevEUI(storage.DB(), d.DevEUI)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	_, _ = nsClient.DeactivateDevice(context.Background(), &ns.DeactivateDeviceRequest{
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

	d, err := storage.GetDevice(storage.DB(), devEUI, false, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	dp, err := storage.GetDeviceProfile(storage.DB(), d.DeviceProfileID, false, false)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	if dp.DeviceProfile.SupportsJoin {
		return nil, grpc.Errorf(codes.FailedPrecondition, "node must be an ABP node")
	}

	n, err := storage.GetNetworkServerForDevEUI(storage.DB(), devEUI)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	_, _ = nsClient.DeactivateDevice(context.Background(), &ns.DeactivateDeviceRequest{
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

	_, err = nsClient.ActivateDevice(context.Background(), &actReq)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	err = storage.CreateDeviceActivation(storage.DB(), &storage.DeviceActivation{
		DevEUI:  d.DevEUI,
		DevAddr: devAddr,
		AppSKey: appSKey,
	})
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	log.WithFields(log.Fields{
		"dev_addr": devAddr,
		"dev_eui":  d.DevEUI,
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

	d, err := storage.GetDevice(storage.DB(), devEUI, false, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	da, err := storage.GetLastDeviceActivationForDevEUI(storage.DB(), devEUI)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	n, err := storage.GetNetworkServerForDevEUI(storage.DB(), devEUI)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	devAct, err := nsClient.GetDeviceActivation(context.Background(), &ns.GetDeviceActivationRequest{
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
			DevEui:      da.DevEUI.String(),
			DevAddr:     devAddr.String(),
			AppSKey:     da.AppSKey.String(),
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

	n, err := storage.GetNetworkServerForDevEUI(storage.DB(), devEUI)
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
			return err
		}

		up, down, err := convertUplinkAndDownlinkFrames(resp.GetUplinkFrameSet(), resp.GetDownlinkFrame(), true)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		var frameResp pb.StreamDeviceFrameLogsResponse
		if up != nil {
			frameResp.Frame = &pb.StreamDeviceFrameLogsResponse_UplinkFrame{
				UplinkFrame: up,
			}
		}

		if down != nil {
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
			Type:        el.Type,
			PayloadJson: string(b),
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

	n, err := storage.GetNetworkServerForDevEUI(storage.DB(), devEUI)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	resp, err := nsClient.GetRandomDevAddr(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}

	var devAddr lorawan.DevAddr
	copy(devAddr[:], resp.DevAddr)

	return &pb.GetRandomDevAddrResponse{
		DevAddr: devAddr.String(),
	}, nil
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

func convertUplinkAndDownlinkFrames(up *gw.UplinkFrameSet, down *gw.DownlinkFrame, decodeMACCommands bool) (*pb.UplinkFrameLog, *pb.DownlinkFrameLog, error) {
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
			PhyPayloadJson: string(phyJSON),
		}

		for _, rxInfo := range up.RxInfo {
			var mac lorawan.EUI64
			copy(mac[:], rxInfo.GatewayId)

			upRXInfo := pb.UplinkRXInfo{
				GatewayId:         mac.String(),
				Time:              rxInfo.Time,
				TimeSinceGpsEpoch: rxInfo.TimeSinceGpsEpoch,
				Timestamp:         rxInfo.Timestamp,
				Rssi:              rxInfo.Rssi,
				LoraSnr:           rxInfo.LoraSnr,
				Channel:           rxInfo.Channel,
				RfChain:           rxInfo.RfChain,
				Board:             rxInfo.Board,
				Antenna:           rxInfo.Antenna,
				Location:          rxInfo.Location,
				FineTimestampType: rxInfo.FineTimestampType,
				Context:           rxInfo.Context,
			}

			switch rxInfo.FineTimestampType {
			case gw.FineTimestampType_ENCRYPTED:
				fineTS := rxInfo.GetEncryptedFineTimestamp()
				if fineTS != nil {
					upRXInfo.FineTimestamp = &pb.UplinkRXInfo_EncryptedFineTimestamp{
						EncryptedFineTimestamp: &pb.EncryptedFineTimestamp{
							AesKeyIndex: fineTS.AesKeyIndex,
							EncryptedNs: fineTS.EncryptedNs,
							FpgaId:      hex.EncodeToString(fineTS.FpgaId),
						},
					}
				}
			case gw.FineTimestampType_PLAIN:
				fineTS := rxInfo.GetPlainFineTimestamp()
				if fineTS != nil {
					upRXInfo.FineTimestamp = &pb.UplinkRXInfo_PlainFineTimestamp{
						PlainFineTimestamp: fineTS,
					}
				}
			}

			uplinkFrameLog.RxInfo = append(uplinkFrameLog.RxInfo, &upRXInfo)
		}

		return &uplinkFrameLog, nil, nil
	}

	if down != nil {
		downlinkFrameLog := pb.DownlinkFrameLog{
			PhyPayloadJson: string(phyJSON),
		}

		if down.TxInfo != nil {
			var mac lorawan.EUI64
			copy(mac[:], down.TxInfo.GatewayId[:])

			downlinkFrameLog.TxInfo = &pb.DownlinkTXInfo{
				GatewayId:  mac.String(),
				Frequency:  down.TxInfo.Frequency,
				Power:      down.TxInfo.Power,
				Modulation: down.TxInfo.Modulation,
				Board:      down.TxInfo.Board,
				Antenna:    down.TxInfo.Antenna,
				Timing:     down.TxInfo.Timing,
				Context:    down.TxInfo.Context,
			}

			if lora := down.TxInfo.GetLoraModulationInfo(); lora != nil {
				downlinkFrameLog.TxInfo.ModulationInfo = &pb.DownlinkTXInfo_LoraModulationInfo{
					LoraModulationInfo: lora,
				}
			}

			if fsk := down.TxInfo.GetFskModulationInfo(); fsk != nil {
				downlinkFrameLog.TxInfo.ModulationInfo = &pb.DownlinkTXInfo_FskModulationInfo{
					FskModulationInfo: fsk,
				}
			}

			if ti := down.TxInfo.GetImmediatelyTimingInfo(); ti != nil {
				downlinkFrameLog.TxInfo.TimingInfo = &pb.DownlinkTXInfo_ImmediatelyTimingInfo{
					ImmediatelyTimingInfo: ti,
				}
			}

			if ti := down.TxInfo.GetDelayTimingInfo(); ti != nil {
				downlinkFrameLog.TxInfo.TimingInfo = &pb.DownlinkTXInfo_DelayTimingInfo{
					DelayTimingInfo: ti,
				}
			}

			if ti := down.TxInfo.GetGpsEpochTimingInfo(); ti != nil {
				downlinkFrameLog.TxInfo.TimingInfo = &pb.DownlinkTXInfo_GpsEpochTimingInfo{
					GpsEpochTimingInfo: ti,
				}
			}
		}

		return nil, &downlinkFrameLog, nil
	}

	return nil, nil, nil
}
