package api

import (
	log "github.com/Sirupsen/logrus"
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

// NodeSessionAPI exports the node-session related functions.
type NodeSessionAPI struct {
	ctx       common.Context
	validator auth.Validator
}

// NewNodeSessionAPI create a new NodeSessionAPI.
func NewNodeSessionAPI(ctx common.Context, validator auth.Validator) *NodeSessionAPI {
	return &NodeSessionAPI{
		ctx:       ctx,
		validator: validator,
	}
}

// Create creates the given node-session.
func (n *NodeSessionAPI) Create(ctx context.Context, req *pb.CreateNodeSessionRequest) (*pb.CreateNodeSessionResponse, error) {
	var devAddr lorawan.DevAddr
	var appEUI, devEUI lorawan.EUI64
	var appSKey, nwkSKey lorawan.AES128Key

	if err := devAddr.UnmarshalText([]byte(req.DevAddr)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devAddr: %s", err)
	}
	if err := appEUI.UnmarshalText([]byte(req.AppEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "appEUI: %s", err)
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

	if err := n.validator.Validate(ctx,
		auth.ValidateAPIMethod("NodeSession.Create"),
		auth.ValidateApplication(appEUI),
		auth.ValidateNode(devEUI),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	node, err := storage.GetNode(n.ctx.DB, devEUI)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "get node error: %s", err)
	}

	if node.AppEUI.String() != appEUI.String() {
		return nil, grpc.Errorf(codes.InvalidArgument, "node belongs to a different AppEUI")
	}

	_, err = n.ctx.NetworkServer.CreateNodeSession(context.Background(), &ns.CreateNodeSessionRequest{
		DevAddr:     devAddr[:],
		AppEUI:      appEUI[:],
		DevEUI:      devEUI[:],
		NwkSKey:     nwkSKey[:],
		FCntUp:      req.FCntUp,
		FCntDown:    req.FCntDown,
		RxDelay:     req.RxDelay,
		Rx1DROffset: req.Rx1DROffset,
		CFList:      req.CFList,
		RxWindow:    ns.RXWindow(req.RxWindow),
		Rx2DR:       req.Rx2DR,
	})
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "create node-session error: %s", err)
	}

	node.AppSKey = appSKey
	if err := storage.UpdateNode(n.ctx.DB, node); err != nil {
		return nil, grpc.Errorf(codes.Internal, "update node error: %s", err)
	}

	log.WithFields(log.Fields{
		"dev_addr": devAddr,
		"app_eui":  appEUI,
		"dev_eui":  devEUI,
	}).Info("node-session created")

	return &pb.CreateNodeSessionResponse{}, nil
}

// Get returns the node-session matching the given DevEUI.
func (n *NodeSessionAPI) Get(ctx context.Context, req *pb.GetNodeSessionRequest) (*pb.GetNodeSessionResponse, error) {
	return nil, nil
}

// Update updates the given node-session.
func (n *NodeSessionAPI) Update(ctx context.Context, req *pb.UpdateNodeSessionRequest) (*pb.UpdateNodeSessionResponse, error) {
	return nil, nil
}

// Delete deletes the node-session matching the given DevEUI.
func (n *NodeSessionAPI) Delete(ctx context.Context, req *pb.DeleteNodeSessionRequest) (*pb.DeleteNodeSessionResponse, error) {
	return nil, nil
}

// GetRandomDevAddr returns a random DevAddr given the NwkID prefix into account.
func (n *NodeSessionAPI) GetRandomDevAddr(ctx context.Context, req *pb.GetRandomDevAddrRequest) (*pb.GetRandomDevAddrResponse, error) {
	return nil, nil
}
