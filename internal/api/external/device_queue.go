package external

import (
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/api/external/auth"
	"github.com/brocaar/chirpstack-application-server/internal/api/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
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

	if err := storage.Transaction(func(tx sqlx.Ext) error {
		// Lock the device to avoid concurrent enqueue actions for the same
		// device as this would result in re-use of the same frame-counter.
		dev, err := storage.GetDevice(ctx, tx, devEUI, true, true)
		if err != nil {
			return helpers.ErrToRPCError(err)
		}

		// if JSON object is set, try to encode it to bytes
		if req.DeviceQueueItem.JsonObject != "" && req.DeviceQueueItem.JsonObject != "null" {
			app, err := storage.GetApplication(ctx, storage.DB(), dev.ApplicationID)
			if err != nil {
				return helpers.ErrToRPCError(err)
			}

			dp, err := storage.GetDeviceProfile(ctx, storage.DB(), dev.DeviceProfileID, false, true)
			if err != nil {
				log.WithError(err).WithField("id", dev.DeviceProfileID).Error("get device-profile error")
				return grpc.Errorf(codes.Internal, "get device-profile error: %s", err)
			}

			// TODO: in the next major release, remove this and always use the
			// device-profile codec fields.
			payloadCodec := app.PayloadCodec
			payloadEncoderScript := app.PayloadEncoderScript

			if dp.PayloadCodec != "" {
				payloadCodec = dp.PayloadCodec
				payloadEncoderScript = dp.PayloadEncoderScript
			}

			req.DeviceQueueItem.Data, err = codec.JSONToBinary(payloadCodec, uint8(req.DeviceQueueItem.FPort), dev.Variables, payloadEncoderScript, []byte(req.DeviceQueueItem.JsonObject))
			if err != nil {
				return helpers.ErrToRPCError(err)
			}
		}

		fCnt, err = storage.EnqueueDownlinkPayload(ctx, tx, devEUI, req.DeviceQueueItem.Confirmed, uint8(req.DeviceQueueItem.FPort), req.DeviceQueueItem.Data)
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

	n, err := storage.GetNetworkServerForDevEUI(ctx, storage.DB(), devEUI)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
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

	device, err := storage.GetDevice(ctx, storage.DB(), devEUI, false, true)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	n, err := storage.GetNetworkServerForDevEUI(ctx, storage.DB(), devEUI)
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return nil, helpers.ErrToRPCError(err)
	}

	queueItemsResp, err := nsClient.GetDeviceQueueItemsForDevEUI(ctx, &ns.GetDeviceQueueItemsForDevEUIRequest{
		DevEui:    devEUI[:],
		CountOnly: req.CountOnly,
	})
	if err != nil {
		return nil, err
	}

	resp := pb.ListDeviceQueueItemsResponse{
		TotalCount: queueItemsResp.TotalCount,
	}
	for _, qi := range queueItemsResp.Items {
		b, err := lorawan.EncryptFRMPayload(device.AppSKey, false, device.DevAddr, qi.FCnt, qi.FrmPayload)
		if err != nil {
			return nil, helpers.ErrToRPCError(err)
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
