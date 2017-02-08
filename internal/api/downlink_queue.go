package api

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
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
	if err := d.validator.Validate(ctx,
		auth.ValidateAPIMethod("DownlinkQueue.Enqueue"),
		auth.ValidateApplicationName(req.ApplicationName),
		auth.ValidateNodeName(req.NodeName),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	node, err := storage.GetNodeByName(d.ctx.DB, req.ApplicationName, req.NodeName)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	qi := storage.DownlinkQueueItem{
		DevEUI:    node.DevEUI,
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
	if err := d.validator.Validate(ctx,
		auth.ValidateAPIMethod("DownlinkQueue.Delete"),
		auth.ValidateApplicationName(req.ApplicationName),
		auth.ValidateNodeName(req.NodeName),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	qi, err := storage.GetDownlinkQueueItem(d.ctx.DB, req.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	node, err := storage.GetNodeByName(d.ctx.DB, req.ApplicationName, req.NodeName)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
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
	if err := d.validator.Validate(ctx,
		auth.ValidateAPIMethod("DownlinkQueue.List"),
		auth.ValidateApplicationName(req.ApplicationName),
		auth.ValidateNodeName(req.NodeName),
	); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	node, err := storage.GetNodeByName(d.ctx.DB, req.ApplicationName, req.NodeName)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
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
			NodeName:  node.Name,
			Confirmed: item.Confirmed,
			Pending:   item.Pending,
			FPort:     uint32(item.FPort),
			Data:      item.Data,
		}
		resp.Items = append(resp.Items, &qi)
	}

	return &resp, nil
}
