package api

import (
	"encoding/json"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/codec"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/downlink"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
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

// Enqueue adds the given item to the device-queue.
func (d *DeviceQueueAPI) Enqueue(ctx context.Context, req *pb.EnqueueDeviceQueueItemRequest) (*pb.EnqueueDeviceQueueItemResponse, error) {
	var fCnt uint32

	if req.DeviceQueueItem == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "queue_item must not be nil")
	}

	if req.DeviceQueueItem.FPort == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "f_port must be > 0")
	}

	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DeviceQueueItem.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateDeviceQueueAccess(devEUI, auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	if err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		// if JSON object is set, try to encode it to bytes
		if req.DeviceQueueItem.JsonObject != "" {
			dev, err := storage.GetDevice(config.C.PostgreSQL.DB, devEUI, true, true)
			if err != nil {
				return errToRPCError(err)
			}

			app, err := storage.GetApplication(config.C.PostgreSQL.DB, dev.ApplicationID)
			if err != nil {
				return errToRPCError(err)
			}

			// get codec payload configured for the application
			codecPL := codec.NewPayload(app.PayloadCodec, uint8(req.DeviceQueueItem.FPort), app.PayloadEncoderScript, app.PayloadDecoderScript)
			if codecPL == nil {
				return grpc.Errorf(codes.FailedPrecondition, "no or invalid codec configured for application")
			}

			err = json.Unmarshal([]byte(req.DeviceQueueItem.JsonObject), &codecPL)
			if err != nil {
				return errToRPCError(err)
			}

			req.DeviceQueueItem.Data, err = codecPL.EncodeToBytes()
			if err != nil {
				return errToRPCError(err)
			}
		}

		var err error
		fCnt, err = downlink.EnqueueDownlinkPayload(tx, devEUI, req.DeviceQueueItem.Confirmed, uint8(req.DeviceQueueItem.FPort), req.DeviceQueueItem.Data)
		if err != nil {
			return grpc.Errorf(codes.Internal, "enqueue downlink payload error: %s", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &pb.EnqueueDeviceQueueItemResponse{
		FCnt: fCnt,
	}, nil
}

// Flush flushes the downlink device-queue.
func (d *DeviceQueueAPI) Flush(ctx context.Context, req *pb.FlushDeviceQueueRequest) (*empty.Empty, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateDeviceQueueAccess(devEUI, auth.Delete)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	n, err := storage.GetNetworkServerForDevEUI(config.C.PostgreSQL.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, errToRPCError(err)
	}

	_, err = nsClient.FlushDeviceQueueForDevEUI(ctx, &ns.FlushDeviceQueueForDevEUIRequest{
		DevEui: devEUI[:],
	})
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// List lists the items in the device-queue.
func (d *DeviceQueueAPI) List(ctx context.Context, req *pb.ListDeviceQueueItemsRequest) (*pb.ListDeviceQueueItemsResponse, error) {
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEui)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateDeviceQueueAccess(devEUI, auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	da, err := storage.GetLastDeviceActivationForDevEUI(config.C.PostgreSQL.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	n, err := storage.GetNetworkServerForDevEUI(config.C.PostgreSQL.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, errToRPCError(err)
	}

	queueItemsResp, err := nsClient.GetDeviceQueueItemsForDevEUI(ctx, &ns.GetDeviceQueueItemsForDevEUIRequest{
		DevEui: devEUI[:],
	})

	var resp pb.ListDeviceQueueItemsResponse
	for _, qi := range queueItemsResp.Items {
		b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, qi.FCnt, qi.FrmPayload)
		if err != nil {
			return nil, errToRPCError(err)
		}

		resp.DeviceQueueItems = append(resp.DeviceQueueItems, &pb.DeviceQueueItem{
			DevEui:    devEUI.String(),
			Confirmed: qi.Confirmed,
			FPort:     qi.FPort,
			FCnt:      qi.FCnt,
			Data:      b,
		})
	}

	return &resp, nil
}
