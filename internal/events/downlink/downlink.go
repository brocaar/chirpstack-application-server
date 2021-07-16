package downlink

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-api/go/v3/as"
	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"

	psql "github.com/brocaar/chirpstack-application-server/internal/integration/postgresql"
)

type downlinkContext struct {
	downlinkDataReq as.HandleDownlinkDataRequest

	ctx           context.Context
	device        storage.Device
	application   storage.Application
	deviceProfile storage.DeviceProfile

	data       []byte
	objectJSON string
}

var tasks = []func(*downlinkContext) error{
	getDevice,
	getApplication,
	getDeviceProfile,
	//decryptPayload,
	handleCodec,
	handlePostgresqlIntegration,
}

// Handle handles the downlink event.
func Handle(ctx context.Context, req as.HandleDownlinkDataRequest) error {
	uc := downlinkContext{
		ctx:             ctx,
		downlinkDataReq: req,
	}

	for _, f := range tasks {
		if err := f(&uc); err != nil {
			if err == ErrAbort {
				return nil
			}
			return err
		}
	}

	return nil
}

func getDevice(ctx *downlinkContext) error {
	var err error
	var devEUI lorawan.EUI64
	copy(devEUI[:], ctx.downlinkDataReq.DevEui)

	ctx.device, err = storage.GetDevice(ctx.ctx, storage.DB(), devEUI, false, true)
	if err != nil {
		return errors.Wrap(err, "get device error")
	}
	return nil
}

func getDeviceProfile(ctx *downlinkContext) error {
	var err error
	ctx.deviceProfile, err = storage.GetDeviceProfile(ctx.ctx, storage.DB(), ctx.device.DeviceProfileID, false, true)
	if err != nil {
		return errors.Wrap(err, "get device-profile error")
	}
	return nil
}

func getApplication(ctx *downlinkContext) error {
	var err error
	ctx.application, err = storage.GetApplication(ctx.ctx, storage.DB(), ctx.device.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get application error")
	}
	return nil
}

func handleCodec(ctx *downlinkContext) error {
	codecType := ctx.application.PayloadCodec
	decoderScript := ctx.application.PayloadDecoderScript

	if ctx.deviceProfile.PayloadCodec != "" {
		codecType = ctx.deviceProfile.PayloadCodec
		decoderScript = ctx.deviceProfile.PayloadDecoderScript
	}

	if codecType == codec.None {
		return nil
	}

	start := time.Now()
	b, err := codec.BinaryToJSON(codecType, uint8(ctx.downlinkDataReq.FPort), ctx.device.Variables, decoderScript, ctx.data)
	if err != nil {
		log.WithFields(log.Fields{
			"codec":          codecType,
			"application_id": ctx.device.ApplicationID,
			"f_port":         ctx.downlinkDataReq.FPort,
			"f_cnt":          ctx.downlinkDataReq.FCnt,
			"dev_eui":        ctx.device.DevEUI,
		}).WithError(err).Error("decode payload error")

		errEvent := pb.ErrorEvent{
			ApplicationId:   uint64(ctx.device.ApplicationID),
			ApplicationName: ctx.application.Name,
			DeviceName:      ctx.device.Name,
			DevEui:          ctx.device.DevEUI[:],
			Type:            pb.ErrorType_DOWNLINK_CODEC,
			Error:           err.Error(),
			FCnt:            ctx.downlinkDataReq.FCnt,
			Tags:            make(map[string]string),
		}

		for k, v := range ctx.device.Tags.Map {
			if v.Valid {
				errEvent.Tags[k] = v.String
			}
		}

		vars := make(map[string]string)
		for k, v := range ctx.device.Variables.Map {
			if v.Valid {
				vars[k] = v.String
			}
		}

		if err := integration.ForApplicationID(ctx.device.ApplicationID).HandleErrorEvent(ctx.ctx, vars, errEvent); err != nil {
			log.WithError(err).Error("send error event to integration error")
		}
	}

	log.WithFields(log.Fields{
		"application_id": ctx.application.ID,
		"codec":          codecType,
		"duration":       time.Since(start),
	}).Debug("payload codec completed Decode execution")

	ctx.objectJSON = string(b)

	return nil
}

func handlePostgresqlIntegration(ctx *downlinkContext) error {
	ctx.data = ctx.downlinkDataReq.Data
	pl := pb.DownlinkEvent{
		ApplicationId:     uint64(ctx.device.ApplicationID),
		ApplicationName:   ctx.application.Name,
		DeviceName:        ctx.device.Name,
		DevEui:            ctx.device.DevEUI[:],
		TxInfo:            ctx.downlinkDataReq.TxInfo,
		Dr:                ctx.downlinkDataReq.Dr,
		Adr:               ctx.downlinkDataReq.Adr,
		FCnt:              ctx.downlinkDataReq.FCnt,
		FPort:             ctx.downlinkDataReq.FPort,
		Data:              ctx.data,
		ObjectJson:        ctx.objectJSON,
		Tags:              make(map[string]string),
		ConfirmedDownlink: ctx.downlinkDataReq.ConfirmedDownlink,
		DevAddr:           ctx.device.DevAddr[:],
		PublishedAt:       ctx.downlinkDataReq.SentAt,
	}

	// set tags
	for k, v := range ctx.device.Tags.Map {
		if v.Valid {
			pl.Tags[k] = v.String
		}
	}
	vars := make(map[string]string)
	for k, v := range ctx.device.Variables.Map {
		if v.Valid {
			vars[k] = v.String
		}
	}

	bgCtx := context.Background()
	bgCtx = context.WithValue(bgCtx, logging.ContextIDKey, ctx.ctx.Value(logging.ContextIDKey))

	err := psql.StaticIntegration.HandleDownlinkEvent(bgCtx, vars, pl)
	if err != nil {
		log.WithError(err).Error("send downlink event error")
	}

	return nil
}
