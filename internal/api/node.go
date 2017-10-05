package api

import (
	"encoding/hex"
	"encoding/json"

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

// NodeAPI exports the Node related functions.
type NodeAPI struct {
	validator auth.Validator
}

// NewNodeAPI creates a new NodeAPI.
func NewNodeAPI(validator auth.Validator) *NodeAPI {
	return &NodeAPI{
		validator: validator,
	}
}

// Create creates the given device.
func (a *NodeAPI) Create(ctx context.Context, req *pb.CreateDeviceRequest) (*pb.CreateDeviceResponse, error) {
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

	if err := storage.CreateDevice(common.DB, &d); err != nil {
		return nil, errToRPCError(err)
	}

	log.WithField("dev_eui", devEUI).Info("device created")

	return &pb.CreateDeviceResponse{}, nil
}

// Get returns the device matching the given DevEUI.
func (a *NodeAPI) Get(ctx context.Context, req *pb.GetDeviceRequest) (*pb.GetDeviceResponse, error) {
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
func (a *NodeAPI) ListByApplicationID(ctx context.Context, req *pb.ListDeviceByApplicationIDRequest) (*pb.ListDeviceResponse, error) {
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
func (a *NodeAPI) Update(ctx context.Context, req *pb.UpdateDeviceRequest) (*pb.UpdateDeviceResponse, error) {
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

	if err := storage.UpdateDevice(common.DB, &d); err != nil {
		return nil, errToRPCError(err)
	}

	log.WithField("dev_eui", devEUI).Info("device updated")

	return &pb.UpdateDeviceResponse{}, nil
}

// Delete deletes the node matching the given name.
func (a *NodeAPI) Delete(ctx context.Context, req *pb.DeleteDeviceRequest) (*pb.DeleteDeviceResponse, error) {
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

	if err := storage.DeleteDevice(common.DB, d.DevEUI); err != nil {
		return nil, errToRPCError(err)
	}

	log.WithField("dev_eui", eui).Info("device deleted")

	return &pb.DeleteDeviceResponse{}, nil
}

// CreateCredentials creates the given device-credentials.
func (a *NodeAPI) CreateCredentials(ctx context.Context, req *pb.CreateDeviceCredentialsRequest) (*pb.CreateDeviceCredentialsResponse, error) {
	panic("not implemented")
}

// GetCredentials returns the device-credentials for the given DevEUI.
func (a *NodeAPI) GetCredentials(ctx context.Context, req *pb.GetDeviceCredentialsRequest) (*pb.GetDeviceCredentialsResponse, error) {
	panic("not implemented")
}

// UpdateCredentials updates the device-credentials.
func (a *NodeAPI) UpdateCredentials(ctx context.Context, req *pb.UpdateDeviceCredentialsRequest) (*pb.UpdateDeviceCredentialsResponse, error) {
	panic("not implemented")
}

// DeleteCredentials deletes the device-credentials for the given DevEUI.
func (a *NodeAPI) DeleteCredentials(ctx context.Context, req *pb.DeleteDeviceCredentialsRequest) (*pb.DeleteDeviceCredentialsResponse, error) {
	panic("not implemeted")
}

// Activate activates the node (ABP only).
func (a *NodeAPI) Activate(ctx context.Context, req *pb.ActivateDeviceRequest) (*pb.ActivateDeviceResponse, error) {
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
	_, _ = common.NetworkServer.DeactivateDevice(context.Background(), &ns.DeactivateDeviceRequest{
		DevEUI: d.DevEUI[:],
	})

	actReq := ns.ActivateDeviceRequest{
		DevEUI:   d.DevEUI[:],
		DevAddr:  devAddr[:],
		NwkSKey:  nwkSKey[:],
		FCntUp:   req.FCntUp,
		FCntDown: req.FCntDown,
		// SkipFCntCheck: d.RelaxFCnt,
	}

	_, err = common.NetworkServer.ActivateDevice(context.Background(), &actReq)
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

	log.WithFields(log.Fields{
		"dev_addr": devAddr,
		"dev_eui":  d.DevEUI,
	}).Info("device activated")

	return &pb.ActivateDeviceResponse{}, nil
}

// GetActivation returns the device activation for the given DevEUI.
func (a *NodeAPI) GetActivation(ctx context.Context, req *pb.GetDeviceActivationRequest) (*pb.GetDeviceActivationResponse, error) {
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

	devAct, err := common.NetworkServer.GetDeviceActivation(context.Background(), &ns.GetDeviceActivationRequest{
		DevEUI: d.DevEUI[:],
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	copy(devAddr[:], devAct.DevAddr)
	copy(nwkSKey[:], devAct.NwkSKey)

	return &pb.GetDeviceActivationResponse{
		DevAddr:  devAddr.String(),
		AppSKey:  da.AppSKey.String(),
		NwkSKey:  nwkSKey.String(),
		FCntUp:   devAct.FCntUp,
		FCntDown: devAct.FCntDown,
	}, nil
}

// GetFrameLogs returns the uplink / downlink frame log for the given DevEUI.
func (a *NodeAPI) GetFrameLogs(ctx context.Context, req *pb.GetFrameLogsRequest) (*pb.GetFrameLogsResponse, error) {
	var devEUI lorawan.EUI64

	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(devEUI, auth.Read)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	resp, err := common.NetworkServer.GetFrameLogsForDevEUI(ctx, &ns.GetFrameLogsForDevEUIRequest{
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
func (a *NodeAPI) GetRandomDevAddr(ctx context.Context, req *pb.GetRandomDevAddrRequest) (*pb.GetRandomDevAddrResponse, error) {
	resp, err := common.NetworkServer.GetRandomDevAddr(context.Background(), &ns.GetRandomDevAddrRequest{})
	if err != nil {
		return nil, err
	}

	var devAddr lorawan.DevAddr
	copy(devAddr[:], resp.DevAddr)

	return &pb.GetRandomDevAddrResponse{
		DevAddr: devAddr.String(),
	}, nil
}

func (a *NodeAPI) returnList(count int, devices []storage.Device) (*pb.ListDeviceResponse, error) {
	resp := pb.ListDeviceResponse{
		TotalCount: int64(count),
	}
	for _, device := range devices {
		item := pb.GetDeviceResponse{
			DevEUI:          device.DevEUI.String(),
			Name:            device.Name,
			Description:     device.Description,
			ApplicationID:   device.ApplicationID,
			DeviceProfileID: device.DeviceProfileID,
		}

		resp.Result = append(resp.Result, &item)
	}
	return &resp, nil
}
