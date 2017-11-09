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

// DeviceQueueAPI exposes the downlink queue methods.
type DeviceQueueAPI struct {
	validator auth.Validator
}

// NewDeviceQueueAPI creates a new DeviceQueueAPI.
func NewDeviceQueueAPI(validator auth.Validator) *DeviceQueueAPI {
	return &DeviceQueueAPI{
		validator: validator,
	}
}

// Enqueue adds the given item to the queue. When the node operates in
// Class-C mode, the data will be pushed directly to the network-server.
func (d *DeviceQueueAPI) Enqueue(ctx context.Context, req *pb.EnqueueDeviceQueueItemRequest) (*pb.EnqueueDeviceQueueItemResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateDeviceQueueAccess(devEUI, auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	device, err := storage.GetDevice(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	qi := storage.DeviceQueueItem{
		DevEUI:    device.DevEUI,
		Reference: req.Reference,
		Confirmed: req.Confirmed,
		FPort:     uint8(req.FPort),
		Data:      req.Data,
	}

	if err := downlink.HandleDownlinkQueueItem(device, &qi); err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.EnqueueDeviceQueueItemResponse{}, nil
}

// Delete deletes an item from the queue.
func (d *DeviceQueueAPI) Delete(ctx context.Context, req *pb.DeleteDeviceQueueItemRequest) (*pb.DeleteDeviceQueueItemResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateDeviceQueueAccess(devEUI, auth.Delete)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	device, err := storage.GetDevice(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	qi, err := storage.GetDeviceQueueItem(common.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}
	if qi.DevEUI != device.DevEUI {
		return nil, grpc.Errorf(codes.NotFound, "queue-item does not exist for the given node")
	}

	if err := storage.DeleteDeviceQueueItem(common.DB, req.Id); err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	return &pb.DeleteDeviceQueueItemResponse{}, nil
}

// List lists the items in the queue for the given node.
func (d *DeviceQueueAPI) List(ctx context.Context, req *pb.ListDeviceQueueItemsRequest) (*pb.ListDeviceQueueItemsResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateDeviceQueueAccess(devEUI, auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	device, err := storage.GetDevice(common.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	items, err := storage.GetDeviceQueueItems(common.DB, device.DevEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var resp pb.ListDeviceQueueItemsResponse
	for _, item := range items {
		qi := pb.DeviceQueueItem{
			Id:        item.ID,
			Reference: item.Reference,
			DevEUI:    device.DevEUI.String(),
			Confirmed: item.Confirmed,
			Pending:   item.Pending,
			FPort:     uint32(item.FPort),
			Data:      item.Data,
		}
		resp.Items = append(resp.Items, &qi)
	}

	return &resp, nil
}
