package api

import (
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/gusseleet/lora-app-server/api"
	"github.com/gusseleet/lora-app-server/internal/api/auth"
	"github.com/gusseleet/lora-app-server/internal/codec"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/downlink"
	"github.com/gusseleet/lora-app-server/internal/storage"
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
	var devEUI lorawan.EUI64
	if err := devEUI.UnmarshalText([]byte(req.DevEUI)); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "devEUI: %s", err)
	}

	if err := d.validator.Validate(ctx,
		auth.ValidateDeviceQueueAccess(devEUI, auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	// if JSON object is set, try to encode it to bytes
	if req.JsonObject != "" {
		dev, err := storage.GetDevice(config.C.PostgreSQL.DB, devEUI)
		if err != nil {
			return nil, errToRPCError(err)
		}

		app, err := storage.GetApplication(config.C.PostgreSQL.DB, dev.ApplicationID)
		if err != nil {
			return nil, errToRPCError(err)
		}

		// get codec payload configured for the application
		codecPL := codec.NewPayload(app.PayloadCodec, uint8(req.FPort), app.PayloadEncoderScript, app.PayloadDecoderScript)
		if codecPL == nil {
			return nil, grpc.Errorf(codes.FailedPrecondition, "no or invalid codec configured for application")
		}

		err = json.Unmarshal([]byte(req.JsonObject), &codecPL)
		if err != nil {
			return nil, errToRPCError(err)
		}

		req.Data, err = codecPL.MarshalBinary()
		if err != nil {
			return nil, errToRPCError(err)
		}
	}

	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
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

	n, err := storage.GetNetworkServerForDevEUI(config.C.PostgreSQL.DB, devEUI)
	if err != nil {
		return nil, errToRPCError(err)
	}

	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, errToRPCError(err)
	}

	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
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
