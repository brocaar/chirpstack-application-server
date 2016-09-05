package api

import (
	"context"

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

func (d *DownlinkQueueAPI) Enqueue(ctx context.Context, req *pb.EnqueueDownlinkQueueItemRequest) (*pb.EnqueueDownlinkQueueItemResponse, error) {
	var devEUI lorawan.EUI64

	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	node, err := storage.GetNode(d.ctx.DB, devEUI)
	if err != nil {
		return grpc.Errorf(codes.Unknown, err.Error())
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateAPIMethod("DownlinkQueue.Enqueue"),
		auth.ValidateApplication(node.AppEUI),
		auth.ValidateNode(node.DevEUI),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	qi := storage.DownlinkQueueItem{
		Confirmed: req.Confirmed,
		FPort:     uint8(req.FPort),
		Data:      req.Data,
	}
	copy(di.DevEUI[:], req.DevEUI)
	if err := storage.CreateDownlinkQueueItem(d.ctx.DB, &qi); err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	return &pb.EnqueueDownlinkQueueItemResponse{}, nil
}

func (d *DownlinkQueueAPI) Delete(ctx context.Context, req *pb.DeleteDownlinkQeueueItemRequest) (*pb.DeleteDownlinkQueueItemResponse, error) {
	qi, err := storage.GetDownlinkQueueItem(d.ctx.DB, req.ID)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	node, err := storage.GetNode(d.ctx.DB, qi.DevEUI)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateAPIMethod("DownlinkQueue.Delete"),
		auth.ValidateApplication(node.AppEUI),
		auth.ValidateNode(node.DevEUI),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	if err := storage.DeleteDownlinkQueueItem(d.ctx.DB, req.ID); err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	return &pb.DeleteDownlinkQueueItemResponse{}, nil
}

func (d *DownlinkQueueAPI) List(ctx context.Context, req *pb.ListDownlinkQueueItemsRequest) (*pb.ListDownlinkQueueItemsResponse, error) {
	var devEUI lorawan.EUI64

	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	node, err := storage.GetNode(d.ctx.DB, devEUI)
	if err != nil {
		return grpc.Errorf(codes.Unknown, err.Error())
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateAPIMethod("DownlinkQueue.List"),
		auth.ValidateApplication(node.AppEUI),
		auth.ValidateNode(node.DevEUI),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	items, err := storage.GetDownlinkQueueItems(d.ctx.DB, node.DevEUI)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	var resp pb.ListDownlinkQueueItemsResponse
	for _, item := range items {
		qi := pb.DownlinkQueueItem{
			Id:        item.ID,
			Confirmed: item.Confirmed,
			Pending:   item.Pending,
			FPort:     uint32(item.FPort),
			Data:      item.Data,
		}
		copy(qi.DevEUI[:], item.DevEUI)
		resp.Items = append(resp.Items, &qi)
	}

	return &resp, nil
}
