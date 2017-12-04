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
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
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

// Enqueue adds the given item to the device-queue.
func (d *DeviceQueueAPI) Enqueue(ctx context.Context, req *pb.EnqueueDeviceQueueItemRequest) (*pb.EnqueueDeviceQueueItemResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateDeviceQueueAccess(devEUI, auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.Transaction(common.DB, func(tx *sqlx.Tx) error {
		if err := downlink.EnqueueDownlinkPayload(tx, devEUI, req.Reference, req.Confirmed, uint8(req.FPort), req.Data); err != nil {
			return errors.Wrap(err, "enqueue downlink payload error")
		}
		return nil
	})
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.EnqueueDeviceQueueItemResponse{}, nil
}

// Flush flushes the downlink device-queue.
func (d *DeviceQueueAPI) Flush(ctx context.Context, req *pb.FlushDeviceQueueRequest) (*pb.FlushDeviceQueueResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateDeviceQueueAccess(devEUI, auth.Delete)); err != nil {
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

	err = storage.Transaction(common.DB, func(tx *sqlx.Tx) error {
		if err := storage.FlushDeviceQueueMappingForDevEUI(tx, devEUI); err != nil {
			return errToRPCError(err)
		}

		_, err := nsClient.FlushDeviceQueueForDevEUI(ctx, &ns.FlushDeviceQueueForDevEUIRequest{
			DevEUI: devEUI[:],
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	return &pb.FlushDeviceQueueResponse{}, nil
}

// List lists the items in the device-queue.
func (d *DeviceQueueAPI) List(ctx context.Context, req *pb.ListDeviceQueueItemsRequest) (*pb.ListDeviceQueueItemsResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateDeviceQueueAccess(devEUI, auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
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

	queueItemsResp, err := nsClient.GetDeviceQueueItemsForDevEUI(ctx, &ns.GetDeviceQueueItemsForDevEUIRequest{
		DevEUI: devEUI[:],
	})

	var resp pb.ListDeviceQueueItemsResponse
	for _, qi := range queueItemsResp.Items {
		b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, qi.FCnt, qi.FrmPayload)
		if err != nil {
			return nil, errToRPCError(err)
		}

		resp.Items = append(resp.Items, &pb.DeviceQueueItem{
			DevEUI:    devEUI.String(),
			Confirmed: qi.Confirmed,
			FPort:     qi.FPort,
			Data:      b,
			FCnt:      qi.FCnt,
		})
	}

	return &resp, nil
}
