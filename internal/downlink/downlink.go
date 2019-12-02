package downlink

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	pb "github.com/brocaar/chirpstack-api/go/as/integration"
	"github.com/brocaar/chirpstack-api/go/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/eventlog"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

// HandleDataDownPayloads handles received downlink payloads to be emitted to the
// devices.
func HandleDataDownPayloads() {
	for pl := range integration.Integration().DataDownChan() {
		go func(pl integration.DataDownPayload) {
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

func handleDataDownPayload(ctx context.Context, pl integration.DataDownPayload) error {
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
			payloadDecoderScript := app.PayloadDecoderScript

			if dp.PayloadCodec != "" {
				payloadCodec = dp.PayloadCodec
				payloadEncoderScript = dp.PayloadEncoderScript
				payloadDecoderScript = dp.PayloadDecoderScript
			}

			// get the codec payload configured for the application
			codecPL := codec.NewPayload(payloadCodec, pl.FPort, payloadEncoderScript, payloadDecoderScript)
			if codecPL == nil {
				logCodecError(ctx, app, d, errors.New("no or invalid codec configured for application"))
				return errors.New("no or invalid codec configured for application")
			}

			err = json.Unmarshal(pl.Object, &codecPL)
			if err != nil {
				logCodecError(ctx, app, d, err)
				return errors.Wrap(err, "unmarshal to codec payload error")
			}

			pl.Data, err = codecPL.EncodeToBytes()
			if err != nil {
				logCodecError(ctx, app, d, err)
				return errors.Wrap(err, "marshal codec payload to binary error")
			}
		}

		if _, err := EnqueueDownlinkPayload(ctx, tx, pl.DevEUI, pl.Confirmed, pl.FPort, pl.Data); err != nil {
			return errors.Wrap(err, "enqueue downlink device-queue item error")
		}

		return nil
	})
}

// EnqueueDownlinkPayload adds the downlink payload to the network-server
// device-queue.
func EnqueueDownlinkPayload(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64, confirmed bool, fPort uint8, data []byte) (uint32, error) {
	// get network-server and network-server api client
	n, err := storage.GetNetworkServerForDevEUI(ctx, db, devEUI)
	if err != nil {
		return 0, errors.Wrap(err, "get network-server error")
	}
	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return 0, errors.Wrap(err, "get network-server client error")
	}

	// get fCnt to use for encrypting and enqueueing
	resp, err := nsClient.GetNextDownlinkFCntForDevEUI(context.Background(), &ns.GetNextDownlinkFCntForDevEUIRequest{
		DevEui: devEUI[:],
	})
	if err != nil {
		return 0, errors.Wrap(err, "get next downlink fcnt for deveui error")
	}

	// get device
	d, err := storage.GetDevice(ctx, db, devEUI, false, true)
	if err != nil {
		return 0, errors.Wrap(err, "get device error")
	}

	// encrypt payload
	b, err := lorawan.EncryptFRMPayload(d.AppSKey, false, d.DevAddr, resp.FCnt, data)
	if err != nil {
		return 0, errors.Wrap(err, "encrypt frmpayload error")
	}

	// enqueue device-queue item
	_, err = nsClient.CreateDeviceQueueItem(ctx, &ns.CreateDeviceQueueItemRequest{
		Item: &ns.DeviceQueueItem{
			DevAddr:    d.DevAddr[:],
			DevEui:     devEUI[:],
			FrmPayload: b,
			FCnt:       resp.FCnt,
			FPort:      uint32(fPort),
			Confirmed:  confirmed,
		},
	})
	if err != nil {
		return 0, errors.Wrap(err, "create device-queue item error")
	}

	log.WithFields(log.Fields{
		"f_cnt":     resp.FCnt,
		"dev_eui":   devEUI,
		"confirmed": confirmed,
	}).Info("downlink device-queue item handled")

	return resp.FCnt, nil
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

	if err := eventlog.LogEventForDevice(d.DevEUI, eventlog.Error, &errEvent); err != nil {
		log.WithError(err).WithField("ctx_id", ctx.Value(logging.ContextIDKey)).Error("log event for device error")
	}

	if err := integration.Integration().SendErrorNotification(ctx, vars, errEvent); err != nil {
		log.WithError(err).WithField("ctx_id", ctx.Value(logging.ContextIDKey)).Error("send error event to integration error")
	}
}
