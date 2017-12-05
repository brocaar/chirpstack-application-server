package api

import (
	"encoding/hex"
	"encoding/json"

	"github.com/jmoiron/sqlx"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
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
func (a *DeviceAPI) Create(ctx context.Context, req *pb.CreateDeviceRequest) (*pb.CreateDeviceResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodesAccess(req.ApplicationID, auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// if Name is "", set it to the DevEUI
	if req.Name == "" {
		req.Name = req.DevEUI
	}

	d := storage.Device{
		DevEUI:          devEUI,
		ApplicationID:   req.ApplicationID,
		DeviceProfileID: req.DeviceProfileID,
		Name:            req.Name,
		Description:     req.Description,
	}

	// as this also performs a remote call to create the node on the
	// network-server, wrap it in a transaction
	err := storage.Transaction(common.DB, func(tx *sqlx.Tx) error {
		return storage.CreateDevice(tx, &d)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateDeviceResponse{}, nil
}

// Get returns the device matching the given DevEUI.
func (a *DeviceAPI) Get(ctx context.Context, req *pb.GetDeviceRequest) (*pb.GetDeviceResponse, error) {
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Read)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	d, err := storage.GetDevice(common.DB, eui)
	if err != nil {
		return nil, errToRPCError(err)
	}

	resp := pb.GetDeviceResponse{
		DevEUI:          d.DevEUI.String(),
		Name:            d.Name,
		ApplicationID:   d.ApplicationID,
		Description:     d.Description,
		DeviceProfileID: d.DeviceProfileID,
	}

	return &resp, nil
}

// ListByApplicationID lists the devices by the given application ID, sorted by the name of the device.
func (a *DeviceAPI) ListByApplicationID(ctx context.Context, req *pb.ListDeviceByApplicationIDRequest) (*pb.ListDeviceResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateNodesAccess(req.ApplicationID, auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	devices, err := storage.GetDevicesForApplicationID(common.DB, req.ApplicationID, int(req.Limit), int(req.Offset), req.Search)
	if err != nil {
		return nil, errToRPCError(err)
	}
	count, err := storage.GetDeviceCountForApplicationID(common.DB, req.ApplicationID, req.Search)
	if err != nil {
		return nil, errToRPCError(err)
	}
	return a.returnList(count, devices)
}

// Update updates the device matching the given DevEUI.
func (a *DeviceAPI) Update(ctx context.Context, req *pb.UpdateDeviceRequest) (*pb.UpdateDeviceResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(devEUI, auth.Update)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	d, err := storage.GetDevice(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	d.DeviceProfileID = req.DeviceProfileID
	d.Name = req.Name
	d.Description = req.Description

	// as this also performs a remote call to update the node on the
	// network-server, wrap it in a transaction
	err = storage.Transaction(common.DB, func(tx *sqlx.Tx) error {
		return storage.UpdateDevice(tx, &d)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.UpdateDeviceResponse{}, nil
}

// Delete deletes the node matching the given name.
func (a *DeviceAPI) Delete(ctx context.Context, req *pb.DeleteDeviceRequest) (*pb.DeleteDeviceResponse, error) {
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Delete)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	d, err := storage.GetDevice(common.DB, eui)
	if err != nil {
		return nil, errToRPCError(err)
	}

	// as this also performs a remote call to delete the node from the
	// network-server, wrap it in a transaction
	err = storage.Transaction(common.DB, func(tx *sqlx.Tx) error {
		return storage.DeleteDevice(tx, d.DevEUI)
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.DeleteDeviceResponse{}, nil
}

// CreateKeys creates the given device-keys.
func (a *DeviceAPI) CreateKeys(ctx context.Context, req *pb.CreateDeviceKeysRequest) (*pb.CreateDeviceKeysResponse, error) {
	if req.DeviceKeys == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devicesKeys expected")
	}

	var key lorawan.AES128Key
	if err := key.UnmarshalText([]byte(req.DeviceKeys.AppKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.CreateDeviceKeys(common.DB, &storage.DeviceKeys{
		DevEUI: eui,
		AppKey: key,
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateDeviceKeysResponse{}, nil
}

// GetKeys returns the device-keys for the given DevEUI.
func (a *DeviceAPI) GetKeys(ctx context.Context, req *pb.GetDeviceKeysRequest) (*pb.GetDeviceKeysResponse, error) {
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	dk, err := storage.GetDeviceKeys(common.DB, eui)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.GetDeviceKeysResponse{
		DeviceKeys: &pb.DeviceKeys{
			AppKey: dk.AppKey.String(),
		},
	}, nil
}

// UpdateKeys updates the device-keys.
func (a *DeviceAPI) UpdateKeys(ctx context.Context, req *pb.UpdateDeviceKeysRequest) (*pb.UpdateDeviceKeysResponse, error) {
	if req.DeviceKeys == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devicesKeys expected")
	}

	var key lorawan.AES128Key
	if err := key.UnmarshalText([]byte(req.DeviceKeys.AppKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Update),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	dk, err := storage.GetDeviceKeys(common.DB, eui)
	if err != nil {
		return nil, errToRPCError(err)
	}
	dk.AppKey = key

	err = storage.UpdateDeviceKeys(common.DB, &dk)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.UpdateDeviceKeysResponse{}, nil
}

// DeleteKeys deletes the device-keys for the given DevEUI.
func (a *DeviceAPI) DeleteKeys(ctx context.Context, req *pb.DeleteDeviceKeysRequest) (*pb.DeleteDeviceKeysResponse, error) {
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Delete),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	if err := storage.DeleteDeviceKeys(common.DB, eui); err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.DeleteDeviceKeysResponse{}, nil
}

// Activate activates the node (ABP only).
func (a *DeviceAPI) Activate(ctx context.Context, req *pb.ActivateDeviceRequest) (*pb.ActivateDeviceResponse, error) {
	var devAddr lorawan.DevAddr
	var devEUI lorawan.EUI64
	var appSKey, nwkSKey lorawan.AES128Key

	if err := devAddr.UnmarshalText([]byte(req.DevAddr)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devAddr: %s", err)
	}
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}
	if err := appSKey.UnmarshalText([]byte(req.AppSKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "appSKey: %s", err)
	}
	if err := nwkSKey.UnmarshalText([]byte(req.NwkSKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "nwkSKey: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(devEUI, auth.Update)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	d, err := storage.GetDevice(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	dp, err := storage.GetDeviceProfile(common.DB, d.DeviceProfileID)
	if err != nil {
		return nil, errToRPCError(err)
	}

	if dp.DeviceProfile.SupportsJoin {
		return nil, grpc.Errorf(codes.FailedPrecondition, "node must be an ABP node")
	}

	// try to remove an existing node-session.
	// TODO: refactor once https://github.com/brocaar/loraserver/pull/124 is in place?
	// so that we can call something like SaveNodeSession which will either
	// create or update an existing node-session
	n, err := storage.GetNetworkServerForDevEUI(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	nsClient, err := common.NetworkServerPool.Get(n.Server)
	if err != nil {
		return nil, errToRPCError(err)
	}

	_, _ = nsClient.DeactivateDevice(context.Background(), &ns.DeactivateDeviceRequest{
		DevEUI: d.DevEUI[:],
	})

	actReq := ns.ActivateDeviceRequest{
		DevEUI:        d.DevEUI[:],
		DevAddr:       devAddr[:],
		NwkSKey:       nwkSKey[:],
		FCntUp:        req.FCntUp,
		FCntDown:      req.FCntDown,
		SkipFCntCheck: req.SkipFCntCheck,
	}

	_, err = nsClient.ActivateDevice(context.Background(), &actReq)
	if err != nil {
		return nil, errToRPCError(err)
	}

	err = storage.CreateDeviceActivation(common.DB, &storage.DeviceActivation{
		DevEUI:  d.DevEUI,
		DevAddr: devAddr,
		AppSKey: appSKey,
		NwkSKey: nwkSKey,
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	if err = storage.FlushDeviceQueueMappingForDevEUI(common.DB, d.DevEUI); err != nil {
		return nil, errToRPCError(err)
	}

	log.WithFields(log.Fields{
		"dev_addr": devAddr,
		"dev_eui":  d.DevEUI,
	}).Info("device activated")

	return &pb.ActivateDeviceResponse{}, nil
}

// GetActivation returns the device activation for the given DevEUI.
func (a *DeviceAPI) GetActivation(ctx context.Context, req *pb.GetDeviceActivationRequest) (*pb.GetDeviceActivationResponse, error) {
	var devAddr lorawan.DevAddr
	var devEUI lorawan.EUI64
	var nwkSKey lorawan.AES128Key

	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(devEUI, auth.Read)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	d, err := storage.GetDevice(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	da, err := storage.GetLastDeviceActivationForDevEUI(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	n, err := storage.GetNetworkServerForDevEUI(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	nsClient, err := common.NetworkServerPool.Get(n.Server)
	if err != nil {
		return nil, errToRPCError(err)
	}

	devAct, err := nsClient.GetDeviceActivation(context.Background(), &ns.GetDeviceActivationRequest{
		DevEUI: d.DevEUI[:],
	})
	if err != nil {
		return nil, err
	}

	copy(devAddr[:], devAct.DevAddr)
	copy(nwkSKey[:], devAct.NwkSKey)

	return &pb.GetDeviceActivationResponse{
		DevAddr:       devAddr.String(),
		AppSKey:       da.AppSKey.String(),
		NwkSKey:       nwkSKey.String(),
		FCntUp:        devAct.FCntUp,
		FCntDown:      devAct.FCntDown,
		SkipFCntCheck: devAct.SkipFCntCheck,
	}, nil
}

// GetFrameLogs returns the uplink / downlink frame log for the given DevEUI.
func (a *DeviceAPI) GetFrameLogs(ctx context.Context, req *pb.GetFrameLogsRequest) (*pb.GetFrameLogsResponse, error) {
	var devEUI lorawan.EUI64

	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(devEUI, auth.Read)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	n, err := storage.GetNetworkServerForDevEUI(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	nsClient, err := common.NetworkServerPool.Get(n.Server)
	if err != nil {
		return nil, errToRPCError(err)
	}

	resp, err := nsClient.GetFrameLogsForDevEUI(ctx, &ns.GetFrameLogsForDevEUIRequest{
		DevEUI: devEUI[:],
		Limit:  int32(req.Limit),
		Offset: int32(req.Offset),
	})
	if err != nil {
		return nil, err
	}

	out := pb.GetFrameLogsResponse{
		TotalCount: resp.TotalCount,
	}

	for i := range resp.Result {
		log := pb.FrameLog{
			CreatedAt: resp.Result[i].CreatedAt,
		}

		if txInfo := resp.Result[i].TxInfo; txInfo != nil {
			log.TxInfo = &pb.TXInfo{
				CodeRate:    txInfo.CodeRate,
				Frequency:   txInfo.Frequency,
				Immediately: txInfo.Immediately,
				Mac:         hex.EncodeToString(txInfo.Mac),
				Power:       txInfo.Power,
				Timestamp:   txInfo.Timestamp,
				DataRate: &pb.DataRate{
					Modulation:   txInfo.DataRate.Modulation,
					BandWidth:    txInfo.DataRate.BandWidth,
					SpreadFactor: txInfo.DataRate.SpreadFactor,
					Bitrate:      txInfo.DataRate.Bitrate,
				},
			}
		}

		for _, rxInfo := range resp.Result[i].RxInfoSet {
			log.RxInfoSet = append(log.RxInfoSet, &pb.RXInfo{
				Channel:   rxInfo.Channel,
				CodeRate:  rxInfo.CodeRate,
				Frequency: rxInfo.Frequency,
				LoRaSNR:   rxInfo.LoRaSNR,
				Rssi:      rxInfo.Rssi,
				Time:      rxInfo.Time,
				Timestamp: rxInfo.Timestamp,
				Mac:       hex.EncodeToString(rxInfo.Mac),
				DataRate: &pb.DataRate{
					Modulation:   rxInfo.DataRate.Modulation,
					BandWidth:    rxInfo.DataRate.BandWidth,
					SpreadFactor: rxInfo.DataRate.SpreadFactor,
					Bitrate:      rxInfo.DataRate.Bitrate,
				},
			})
		}

		var phy lorawan.PHYPayload
		if err = phy.UnmarshalBinary(resp.Result[i].PhyPayload); err != nil {
			return nil, errToRPCError(err)
		}

		phyB, err := json.Marshal(phy)
		if err != nil {
			return nil, errToRPCError(err)
		}
		log.PhyPayloadJSON = string(phyB)

		out.Result = append(out.Result, &log)
	}

	return &out, nil
}

// GetRandomDevAddr returns a random DevAddr taking the NwkID prefix into account.
func (a *DeviceAPI) GetRandomDevAddr(ctx context.Context, req *pb.GetRandomDevAddrRequest) (*pb.GetRandomDevAddrResponse, error) {
	var devEUI lorawan.EUI64

	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	n, err := storage.GetNetworkServerForDevEUI(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	nsClient, err := common.NetworkServerPool.Get(n.Server)
	if err != nil {
		return nil, errToRPCError(err)
	}

	resp, err := nsClient.GetRandomDevAddr(context.Background(), &ns.GetRandomDevAddrRequest{})
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
			DevEUI:            device.DevEUI.String(),
			Name:              device.Name,
			Description:       device.Description,
			ApplicationID:     device.ApplicationID,
			DeviceProfileID:   device.DeviceProfileID,
			DeviceProfileName: device.DeviceProfileName,
		}

		resp.Result = append(resp.Result, &item)
	}
	return &resp, nil
}
