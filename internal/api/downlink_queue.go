package api

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/downlink"
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
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateNodeQueueAccess(devEUI, auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	node, err := storage.GetNode(d.ctx.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	qi := storage.DownlinkQueueItem{
		DevEUI:    node.DevEUI,
		Reference: req.Reference,
		Confirmed: req.Confirmed,
		FPort:     uint8(req.FPort),
		Data:      req.Data,
	}

	if err := downlink.HandleDownlinkQueueItem(d.ctx, node, &qi); err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.EnqueueDownlinkQueueItemResponse{}, nil
}

func (d *DownlinkQueueAPI) Delete(ctx context.Context, req *pb.DeleteDownlinkQeueueItemRequest) (*pb.DeleteDownlinkQueueItemResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateNodeQueueAccess(devEUI, auth.Delete)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	node, err := storage.GetNode(d.ctx.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	qi, err := storage.GetDownlinkQueueItem(d.ctx.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}
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
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateNodeQueueAccess(devEUI, auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	node, err := storage.GetNode(d.ctx.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	items, err := storage.GetDownlinkQueueItems(d.ctx.DB, node.DevEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var resp pb.ListDownlinkQueueItemsResponse
	for _, item := range items {
		qi := pb.DownlinkQueueItem{
			Id:        item.ID,
			Reference: item.Reference,
			DevEUI:    node.DevEUI.String(),
			Confirmed: item.Confirmed,
			Pending:   item.Pending,
			FPort:     uint32(item.FPort),
			Data:      item.Data,
		}
		resp.Items = append(resp.Items, &qi)
	}

	return &resp, nil
}
