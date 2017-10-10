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

// Create creates the given Node.
func (a *NodeAPI) Create(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeResponse, error) {
	var appEUI, devEUI lorawan.EUI64
	var appKey lorawan.AES128Key

	if err := appEUI.UnmarshalText([]byte(req.AppEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}
	if err := appKey.UnmarshalText([]byte(req.AppKey)); err != nil {
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

	node := storage.Node{
		ApplicationID:          req.ApplicationID,
		UseApplicationSettings: req.UseApplicationSettings,
		Name:        req.Name,
		Description: req.Description,
		DevEUI:      devEUI,
		AppEUI:      appEUI,
		AppKey:      appKey,
		IsABP:       req.IsABP,
		IsClassC:    req.IsClassC,
		RelaxFCnt:   req.RelaxFCnt,

		RXDelay:     uint8(req.RxDelay),
		RX1DROffset: uint8(req.Rx1DROffset),
		RXWindow:    storage.RXWindow(req.RxWindow),
		RX2DR:       uint8(req.Rx2DR),

		ADRInterval:        req.AdrInterval,
		InstallationMargin: req.InstallationMargin,
	}

	if err := storage.CreateNode(common.DB, node); err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateNodeResponse{}, nil
}

// Get returns the Node for the given name.
func (a *NodeAPI) Get(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Read)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	node, err := storage.GetNode(common.DB, eui)
	if err != nil {
		return nil, errToRPCError(err)
	}

	resp := pb.GetNodeResponse{
		Name:                   node.Name,
		Description:            node.Description,
		DevEUI:                 node.DevEUI.String(),
		AppEUI:                 node.AppEUI.String(),
		AppKey:                 node.AppKey.String(),
		IsABP:                  node.IsABP,
		IsClassC:               node.IsClassC,
		RxDelay:                uint32(node.RXDelay),
		Rx1DROffset:            uint32(node.RX1DROffset),
		RxWindow:               pb.RXWindow(node.RXWindow),
		Rx2DR:                  uint32(node.RX2DR),
		RelaxFCnt:              node.RelaxFCnt,
		AdrInterval:            node.ADRInterval,
		InstallationMargin:     node.InstallationMargin,
		ApplicationID:          node.ApplicationID,
		UseApplicationSettings: node.UseApplicationSettings,
	}

	return &resp, nil
}

// ListByApplicationID returns a list of nodes (given an application id, limit and offset).
func (a *NodeAPI) ListByApplicationID(ctx context.Context, req *pb.ListNodeByApplicationIDRequest) (*pb.ListNodeResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateNodesAccess(req.ApplicationID, auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	nodes, err := storage.GetNodesForApplicationID(common.DB, req.ApplicationID, int(req.Limit), int(req.Offset), req.Search)
	if err != nil {
		return nil, errToRPCError(err)
	}
	count, err := storage.GetNodesCountForApplicationID(common.DB, req.ApplicationID, req.Search)

	if err != nil {
		return nil, errToRPCError(err)
	}
	return a.returnList(count, nodes)
}

// Update updates the node matching the given name.
func (a *NodeAPI) Update(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.UpdateNodeResponse, error) {
	var appEUI, devEUI lorawan.EUI64
	var appKey lorawan.AES128Key

	if err := appEUI.UnmarshalText([]byte(req.AppEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}
	if err := appKey.UnmarshalText([]byte(req.AppKey)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(devEUI, auth.Update)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	node, err := storage.GetNode(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	node.Name = req.Name
	node.Description = req.Description
	node.AppEUI = appEUI
	node.AppKey = appKey
	node.IsABP = req.IsABP
	node.IsClassC = req.IsClassC
	node.RXDelay = uint8(req.RxDelay)
	node.RX1DROffset = uint8(req.Rx1DROffset)
	node.RXWindow = storage.RXWindow(req.RxWindow)
	node.RX2DR = uint8(req.Rx2DR)
	node.RelaxFCnt = req.RelaxFCnt
	node.ADRInterval = req.AdrInterval
	node.InstallationMargin = req.InstallationMargin
	node.ApplicationID = req.ApplicationID
	node.UseApplicationSettings = req.UseApplicationSettings

	if err := storage.UpdateNode(common.DB, node); err != nil {
		return nil, errToRPCError(err)
	}

	log.WithFields(log.Fields{
		"dev_eui":        node.DevEUI,
		"application_id": node.ApplicationID,
	}).Info("node updated")

	return &pb.UpdateNodeResponse{}, nil
}

// Delete deletes the node matching the given name.
func (a *NodeAPI) Delete(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.DeleteNodeResponse, error) {
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := a.validator.Validate(ctx,
		auth.ValidateNodeAccess(eui, auth.Delete)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	node, err := storage.GetNode(common.DB, eui)
	if err != nil {
		return nil, errToRPCError(err)
	}

	if err := storage.DeleteNode(common.DB, node.DevEUI); err != nil {
		return nil, errToRPCError(err)
	}

	// try to delete the node-session
	_, _ = common.NetworkServer.DeleteNodeSession(context.Background(), &ns.DeleteNodeSessionRequest{
		DevEUI: node.DevEUI[:],
	})

	log.WithFields(log.Fields{
		"dev_eui":        node.DevEUI,
		"application_id": node.ApplicationID,
	}).Info("node deleted")

	return &pb.DeleteNodeResponse{}, nil
}

// Activate activates the node (ABP only).
func (a *NodeAPI) Activate(ctx context.Context, req *pb.ActivateNodeRequest) (*pb.ActivateNodeResponse, error) {
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

	node, err := storage.GetNode(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	if !node.IsABP {
		return nil, grpc.Errorf(codes.FailedPrecondition, "node must be an ABP node")
	}

	// try to remove an existing node-session.
	// TODO: refactor once https://github.com/brocaar/loraserver/pull/124 is in place?
	// so that we can call something like SaveNodeSession which will either
	// create or update an existing node-session
	_, _ = common.NetworkServer.DeleteNodeSession(context.Background(), &ns.DeleteNodeSessionRequest{
		DevEUI: node.DevEUI[:],
	})

	createNSReq := ns.CreateNodeSessionRequest{
		DevAddr:            devAddr[:],
		AppEUI:             node.AppEUI[:],
		DevEUI:             node.DevEUI[:],
		NwkSKey:            nwkSKey[:],
		FCntUp:             req.FCntUp,
		FCntDown:           req.FCntDown,
		RxDelay:            uint32(node.RXDelay),
		Rx1DROffset:        uint32(node.RX1DROffset),
		RxWindow:           ns.RXWindow(node.RXWindow),
		Rx2DR:              uint32(node.RX2DR),
		RelaxFCnt:          node.RelaxFCnt,
		AdrInterval:        node.ADRInterval,
		InstallationMargin: node.InstallationMargin,
	}

	_, err = common.NetworkServer.CreateNodeSession(context.Background(), &createNSReq)
	if err != nil {
		return nil, errToRPCError(err)
	}

	node.AppSKey = appSKey
	node.DevAddr = devAddr
	node.NwkSKey = nwkSKey

	if err = storage.UpdateNode(common.DB, node); err != nil {
		return nil, errToRPCError(err)
	}

	log.WithFields(log.Fields{
		"dev_addr":       devAddr,
		"dev_eui":        node.DevEUI,
		"application_id": node.ApplicationID,
	}).Info("node activated")

	return &pb.ActivateNodeResponse{}, nil
}

func (a *NodeAPI) GetActivation(ctx context.Context, req *pb.GetNodeActivationRequest) (*pb.GetNodeActivationResponse, error) {
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

	node, err := storage.GetNode(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	ns, err := common.NetworkServer.GetNodeSession(context.Background(), &ns.GetNodeSessionRequest{
		DevEUI: node.DevEUI[:],
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	copy(devAddr[:], ns.DevAddr)
	copy(nwkSKey[:], ns.NwkSKey)

	return &pb.GetNodeActivationResponse{
		DevAddr:  devAddr.String(),
		AppSKey:  node.AppSKey.String(),
		NwkSKey:  nwkSKey.String(),
		FCntUp:   ns.FCntUp,
		FCntDown: ns.FCntDown,
	}, nil
}

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

func (a *NodeAPI) returnList(count int, nodes []storage.Node) (*pb.ListNodeResponse, error) {
	resp := pb.ListNodeResponse{
		TotalCount: int64(count),
	}
	for _, node := range nodes {
		item := pb.GetNodeResponse{
			Name:                   node.Name,
			Description:            node.Description,
			DevEUI:                 node.DevEUI.String(),
			AppEUI:                 node.AppEUI.String(),
			AppKey:                 node.AppKey.String(),
			IsABP:                  node.IsABP,
			IsClassC:               node.IsClassC,
			RxDelay:                uint32(node.RXDelay),
			Rx1DROffset:            uint32(node.RX1DROffset),
			RxWindow:               pb.RXWindow(node.RXWindow),
			Rx2DR:                  uint32(node.RX2DR),
			RelaxFCnt:              node.RelaxFCnt,
			AdrInterval:            node.ADRInterval,
			InstallationMargin:     node.InstallationMargin,
			ApplicationID:          node.ApplicationID,
			UseApplicationSettings: node.UseApplicationSettings,
		}

		resp.Result = append(resp.Result, &item)
	}
	return &resp, nil
}
