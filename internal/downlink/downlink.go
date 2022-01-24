package downlink

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

// HandleDataDownPayloads handles received downlink payloads to be emitted to the
// devices.
func HandleDataDownPayloads(downChan chan models.DataDownPayload) {
	for pl := range downChan {
		go func(pl models.DataDownPayload) {
			ctxID, err := uuid.NewV4()
			if err != nil {
				log.WithError(err).Error("new uuid error")
				return
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, logging.ContextIDKey, ctxID)

			if err := handleDataDownPayload(ctx, pl); err != nil {
				log.WithFields(log.Fields{
					"dev_eui":        pl.DevEUI,
					"application_id": pl.ApplicationID,
				}).Errorf("handle data-down payload error: %s", err)
			}
		}(pl)
	}
}

func handleDataDownPayload(ctx context.Context, pl models.DataDownPayload) error {
	return storage.Transaction(func(tx sqlx.Ext) error {
		// lock the device so that a concurrent Enqueue action will block
		// until this transaction has been completed
		d, err := storage.GetDevice(ctx, tx, pl.DevEUI, true, true)
		if err != nil {
			return fmt.Errorf("get device error: %s", err)
		}

		// Validate that the ApplicationID matches the actual DevEUI.
		// This is needed as authorisation might be performed on MQTT topic level
		// where it is unknown if the given ApplicationID matches the given
		// DevEUI.
		if d.ApplicationID != pl.ApplicationID {
			return errors.New("enqueue downlink payload: device does not exist for given application")
		}

		// if Object is set, try to encode it to bytes using the application codec
		//if pl.Object != nil && string(pl.Object) != "null" {
		if pl.Object != nil && string(pl.Object) != "null" {
			app, err := storage.GetApplication(ctx, tx, d.ApplicationID)
			if err != nil {
				return errors.Wrap(err, "get application error")
			}

			dp, err := storage.GetDeviceProfile(ctx, storage.DB(), d.DeviceProfileID, false, true)
			if err != nil {
				return errors.Wrap(err, "get device-profile error")
			}

			// TODO: in the next major release, remove this and always use the
			// device-profile codec fields.
			payloadCodec := app.PayloadCodec
			payloadEncoderScript := app.PayloadEncoderScript

			if dp.PayloadCodec != "" {
				payloadCodec = dp.PayloadCodec
				payloadEncoderScript = dp.PayloadEncoderScript
			}

			pl.Data, err = codec.JSONToBinary(payloadCodec, pl.FPort, d.Variables, payloadEncoderScript, []byte(pl.Object))
			if err != nil {
				logCodecError(ctx, app, d, err)
				return errors.Wrap(err, "encode object error")
			}
		}

		if _, err := storage.EnqueueDownlinkPayload(ctx, tx, pl.DevEUI, pl.Confirmed, pl.FPort, pl.Data); err != nil {
			return errors.Wrap(err, "enqueue downlink device-queue item error")
		}

		return nil
	})
}

func logCodecError(ctx context.Context, a storage.Application, d storage.Device, err error) {
	errEvent := pb.ErrorEvent{
		ApplicationId:   uint64(a.ID),
		ApplicationName: a.Name,
		DeviceName:      d.Name,
		DevEui:          d.DevEUI[:],
		Type:            pb.ErrorType_DOWNLINK_CODEC,
		Error:           err.Error(),
		Tags:            make(map[string]string),
		PublishedAt:     ptypes.TimestampNow(),
	}

	for k, v := range d.Tags.Map {
		if v.Valid {
			errEvent.Tags[k] = v.String
		}
	}

	vars := make(map[string]string)
	for k, v := range d.Variables.Map {
		if v.Valid {
			vars[k] = v.String
		}
	}

	if err := integration.ForApplicationID(a.ID).HandleErrorEvent(ctx, vars, errEvent); err != nil {
		log.WithError(err).WithField("ctx_id", ctx.Value(logging.ContextIDKey)).Error("send error event to integration error")
	}
}
