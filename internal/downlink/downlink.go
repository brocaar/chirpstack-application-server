package downlink

import (
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/codec"
	"github.com/brocaar/lora-app-server/internal/eventlog"
	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

// HandleDataDownPayloads handles received downlink payloads to be emitted to the
// devices.
func HandleDataDownPayloads() {
	for pl := range integration.Integration().DataDownChan() {
		go func(pl integration.DataDownPayload) {
			if err := handleDataDownPayload(pl); err != nil {
				log.WithFields(log.Fields{
					"dev_eui":        pl.DevEUI,
					"application_id": pl.ApplicationID,
				}).Errorf("handle data-down payload error: %s", err)
			}
		}(pl)
	}
}

func handleDataDownPayload(pl integration.DataDownPayload) error {
	return storage.Transaction(func(tx sqlx.Ext) error {
		// lock the device so that a concurrent Enqueue action will block
		// until this transaction has been completed
		d, err := storage.GetDevice(tx, pl.DevEUI, true, true)
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
		if pl.Object != nil {
			app, err := storage.GetApplication(tx, d.ApplicationID)
			if err != nil {
				return errors.Wrap(err, "get application error")
			}

			// get the codec payload configured for the application
			codecPL := codec.NewPayload(app.PayloadCodec, pl.FPort, app.PayloadEncoderScript, app.PayloadDecoderScript)
			if codecPL == nil {
				logCodecError(app, d, errors.New("no or invalid codec configured for application"))
				return errors.New("no or invalid codec configured for application")
			}

			err = json.Unmarshal(pl.Object, &codecPL)
			if err != nil {
				logCodecError(app, d, err)
				return errors.Wrap(err, "unmarshal to codec payload error")
			}

			pl.Data, err = codecPL.EncodeToBytes()
			if err != nil {
				logCodecError(app, d, err)
				return errors.Wrap(err, "marshal codec payload to binary error")
			}
		}

		if _, err := EnqueueDownlinkPayload(tx, pl.DevEUI, pl.Confirmed, pl.FPort, pl.Data); err != nil {
			return errors.Wrap(err, "enqueue downlink device-queue item error")
		}

		return nil
	})
}

// EnqueueDownlinkPayload adds the downlink payload to the network-server
// device-queue.
func EnqueueDownlinkPayload(db sqlx.Ext, devEUI lorawan.EUI64, confirmed bool, fPort uint8, data []byte) (uint32, error) {
	// get network-server and network-server api client
	n, err := storage.GetNetworkServerForDevEUI(db, devEUI)
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

	// get current device-activation for AppSKey
	da, err := storage.GetLastDeviceActivationForDevEUI(db, devEUI)
	if err != nil {
		return 0, errors.Wrap(err, "get last device-activation error")
	}

	// encrypt payload
	b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, resp.FCnt, data)
	if err != nil {
		return 0, errors.Wrap(err, "encrypt frmpayload error")
	}

	// enqueue device-queue item
	_, err = nsClient.CreateDeviceQueueItem(context.Background(), &ns.CreateDeviceQueueItemRequest{
		Item: &ns.DeviceQueueItem{
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

func logCodecError(a storage.Application, d storage.Device, err error) {
	errNotification := integration.ErrorNotification{
		ApplicationID:   a.ID,
		ApplicationName: a.Name,
		DeviceName:      d.Name,
		DevEUI:          d.DevEUI,
		Type:            "CODEC",
		Error:           err.Error(),
	}

	if err := eventlog.LogEventForDevice(d.DevEUI, eventlog.EventLog{
		Type:    eventlog.Error,
		Payload: errNotification,
	}); err != nil {
		log.WithError(err).Error("log event for device error")
	}

	if err := integration.Integration().SendErrorNotification(errNotification); err != nil {
		log.WithError(err).Error("send error notification to integration error")
	}
}
