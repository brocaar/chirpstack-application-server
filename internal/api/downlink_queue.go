package api

import (
	"encoding/hex"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lorawan"
)

// DownlinkQueueAPI exposes the downlink queue methods.
type DownlinkQueueAPI struct {
	ctx       common.Context
	validator auth.Validator
}

// NewDownlinkQueueAPI creates a new DownlinkQueueAPI.
func NewDownlinkQueueAPI(ctx common.Context, validator auth.Validator) *DownlinkQueueAPI {
	return &DownlinkQueueAPI{
		ctx:       ctx,
		validator: validator,
	}
}

func (d *DownlinkQueueAPI) Enqueue(ctx context.Context, req *pb.EnqueueDownlinkQueueItemRequest) (*pb.EnqueueDownlinkQueueItemResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateAPIMethod("DownlinkQueue.Enqueue"),
		auth.ValidateApplicationName(req.ApplicationName),
		auth.ValidateNode(devEUI),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	app, err := storage.GetApplicationByName(d.ctx.DB, req.ApplicationName)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	node, err := storage.GetNode(d.ctx.DB, devEUI)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	// test that the node belongs to the given application
	if node.ApplicationID != app.ID {
		return nil, grpc.Errorf(codes.NotFound, "node does not exists for the given application")
	}

	qi := storage.DownlinkQueueItem{
		DevEUI:    devEUI,
		Reference: req.Reference,
		Confirmed: req.Confirmed,
		FPort:     uint8(req.FPort),
		Data:      req.Data,
	}

	if err := storage.CreateDownlinkQueueItem(d.ctx.DB, &qi); err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	return &pb.EnqueueDownlinkQueueItemResponse{}, nil
}

func (d *DownlinkQueueAPI) Delete(ctx context.Context, req *pb.DeleteDownlinkQeueueItemRequest) (*pb.DeleteDownlinkQueueItemResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateAPIMethod("DownlinkQueue.Delete"),
		auth.ValidateApplicationName(req.ApplicationName),
		auth.ValidateNode(devEUI),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	qi, err := storage.GetDownlinkQueueItem(d.ctx.DB, req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	node, err := storage.GetNode(d.ctx.DB, qi.DevEUI)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	app, err := storage.GetApplication(d.ctx.DB, node.ApplicationID)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	// test that the node belongs to the given application
	if node.ApplicationID != app.ID {
		return nil, grpc.Errorf(codes.NotFound, "node does not exists for the given application")
	}
	// test that the queue-item belongs to the given node
	if qi.DevEUI != node.DevEUI {
		return nil, grpc.Errorf(codes.NotFound, "queue-item does not exist for the given node")
	}

	if err := storage.DeleteDownlinkQueueItem(d.ctx.DB, req.Id); err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	return &pb.DeleteDownlinkQueueItemResponse{}, nil
}

func (d *DownlinkQueueAPI) List(ctx context.Context, req *pb.ListDownlinkQueueItemsRequest) (*pb.ListDownlinkQueueItemsResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateAPIMethod("DownlinkQueue.List"),
		auth.ValidateApplicationName(req.ApplicationName),
		auth.ValidateNode(devEUI),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	node, err := storage.GetNode(d.ctx.DB, devEUI)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	app, err := storage.GetApplication(d.ctx.DB, node.ApplicationID)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	// test that the node belongs to the given application
	if node.ApplicationID != app.ID {
		return nil, grpc.Errorf(codes.NotFound, "node does not exists for the given application")
	}

	items, err := storage.GetDownlinkQueueItems(d.ctx.DB, node.DevEUI)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	var resp pb.ListDownlinkQueueItemsResponse
	for _, item := range items {
		qi := pb.DownlinkQueueItem{
			Id:        item.ID,
			Reference: item.Reference,
			DevEUI:    hex.EncodeToString(item.DevEUI[:]),
			Confirmed: item.Confirmed,
			Pending:   item.Pending,
			FPort:     uint32(item.FPort),
			Data:      item.Data,
		}
		resp.Items = append(resp.Items, &qi)
	}

	return &resp, nil
}
