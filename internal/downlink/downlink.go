package downlink

import (
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/gusseleet/lora-app-server/internal/codec"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/handler"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

// HandleDataDownPayloads handles received downlink payloads to be emitted to the
// devices.
func HandleDataDownPayloads() {
	for pl := range config.C.ApplicationServer.Integration.Handler.DataDownChan() {
		go func(pl handler.DataDownPayload) {
			if err := handleDataDownPayload(pl); err != nil {
				log.WithFields(log.Fields{
					"dev_eui":        pl.DevEUI,
					"application_id": pl.ApplicationID,
					"reference":      pl.Reference,
				}).Errorf("handle data-down payload error: %s", err)
			}
		}(pl)
	}
}

func handleDataDownPayload(pl handler.DataDownPayload) error {
	d, err := storage.GetDevice(config.C.PostgreSQL.DB, pl.DevEUI)
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
		app, err := storage.GetApplication(config.C.PostgreSQL.DB, d.ApplicationID)
		if err != nil {
			return errors.Wrap(err, "get application error")
		}

		// get the codec payload configured for the application
		codecPL := codec.NewPayload(app.PayloadCodec, pl.FPort, app.PayloadEncoderScript, app.PayloadDecoderScript)
		if codecPL == nil {
			log.WithFields(log.Fields{
				"application_id": app.ID,
				"codec_type":     app.PayloadCodec,
			}).Error("no or invalid codec configured for application")
			return errors.New("no or invalid codec configured for application")
		}

		err = json.Unmarshal(pl.Object, &codecPL)
		if err != nil {
			return errors.Wrap(err, "unmarshal to codec payload error")
		}

		pl.Data, err = codecPL.MarshalBinary()
		if err != nil {
			return errors.Wrap(err, "marshal codec payload to binary error")
		}
	}

	return storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		if err := EnqueueDownlinkPayload(tx, pl.DevEUI, pl.Reference, pl.Confirmed, pl.FPort, pl.Data); err != nil {
			return errors.Wrap(err, "enqueue downlink device-queue item error")
		}
		return nil
	})
}

// EnqueueDownlinkPayload adds the downlink payload to the network-server
// device-queue.
func EnqueueDownlinkPayload(db sqlx.Ext, devEUI lorawan.EUI64, reference string, confirmed bool, fPort uint8, data []byte) error {
	// get network-server and network-server api client
	n, err := storage.GetNetworkServerForDevEUI(db, devEUI)
	if err != nil {
		return errors.Wrap(err, "get network-server error")
	}
	nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return errors.Wrap(err, "get network-server client error")
	}

	// get fCnt to use for encrypting and enqueueing
	resp, err := nsClient.GetNextDownlinkFCntForDevEUI(context.Background(), &ns.GetNextDownlinkFCntForDevEUIRequest{
		DevEUI: devEUI[:],
	})
	if err != nil {
		return errors.Wrap(err, "get next downlink fcnt for deveui error")
	}

	// get current device-activation for AppSKey
	da, err := storage.GetLastDeviceActivationForDevEUI(db, devEUI)
	if err != nil {
		return errors.Wrap(err, "get last device-activation error")
	}

	// encrypt payload
	b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, resp.FCnt, data)
	if err != nil {
		return errors.Wrap(err, "encrypt frmpayload error")
	}

	// create device-queue mapping (for mapping a device-queue item to an
	// user-given reference)
	if confirmed == true {
		err = storage.CreateDeviceQueueMapping(db, &storage.DeviceQueueMapping{
			Reference: reference,
			DevEUI:    devEUI,
			FCnt:      resp.FCnt,
		})
		if err != nil {
			return errors.Wrap(err, "create device-queue mapping error")
		}
	}

	// enqueue device-queue item
	_, err = nsClient.CreateDeviceQueueItem(context.Background(), &ns.CreateDeviceQueueItemRequest{
		Item: &ns.DeviceQueueItem{
			DevEUI:     devEUI[:],
			FrmPayload: b,
			FCnt:       resp.FCnt,
			FPort:      uint32(fPort),
			Confirmed:  confirmed,
		},
	})
	if err != nil {
		return errors.Wrap(err, "create device-queue item error")
	}

	log.WithFields(log.Fields{
		"f_cnt":     resp.FCnt,
		"dev_eui":   devEUI,
		"reference": reference,
		"confirmed": confirmed,
	}).Info("downlink device-queue item handled")

	return nil
}
