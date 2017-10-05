package downlink

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

// HandleDataDownPayloads handles received downlink payloads to be emitted to the
// devices.
func HandleDataDownPayloads() {
	for pl := range common.Handler.DataDownChan() {
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
	d, err := storage.GetDevice(common.DB, pl.DevEUI)
	if err != nil {
		return fmt.Errorf("get device error: %s", err)
	}

	// Validate that the ApplicationID matches the actual DevEUI.
	// This is needed as authorisation might be performed on MQTT topic level
	// where it is unknown if the given ApplicationID matches the given
	// DevEUI.
	if d.ApplicationID != pl.ApplicationID {
		return errors.New("enqueue data-down payload: device does not exist for given application")
	}

	qi := storage.DeviceQueueItem{
		Reference: pl.Reference,
		DevEUI:    pl.DevEUI,
		Confirmed: pl.Confirmed,
		FPort:     pl.FPort,
		Data:      pl.Data,
	}

	return HandleDownlinkQueueItem(d, &qi)
}

// HandleDownlinkQueueItem handles a DownlinkQueueItem to be emitted to the device.
// In case of class-c, it will send the payload directly to the network-server.
// In any other case, it will be enqueued.
func HandleDownlinkQueueItem(d storage.Device, qi *storage.DeviceQueueItem) error {
	dp, err := storage.GetDeviceProfile(common.DB, d.DeviceProfileID)
	if err != nil {
		return errors.Wrap(err, "get device-profile error")
	}

	if dp.DeviceProfile.SupportsClassC && qi.Confirmed {
		qi.Pending = true
	}

	// In case of a class-c device, we directly push the payload to the
	// network-server.
	// Before pushing, we purge the queue to make sure we have always a single
	// item in the queue in case of confirmed data.
	if dp.DeviceProfile.SupportsClassC {
		if err := storage.DeleteDeviceQueueItemsForDevEUI(common.DB, d.DevEUI); err != nil {
			return err
		}

		if err := pushDataDown(d, qi); err != nil {
			return err
		}
	}

	// save the queue-item in every case, except when the node is a class-c
	// device and the data is unconfirmed.
	if !(dp.DeviceProfile.SupportsClassC && !qi.Confirmed) {
		if err := storage.CreateDeviceQueueItem(common.DB, qi); err != nil {
			return fmt.Errorf("create downlink queue item error: %s", err)
		}
	}

	return nil
}

func pushDataDown(d storage.Device, qi *storage.DeviceQueueItem) error {
	da, err := storage.GetLastDeviceActivationForDevEUI(common.DB, d.DevEUI)
	if err != nil {
		return errors.Wrap(err, "get device-activation error")
	}

	actRes, err := common.NetworkServer.GetDeviceActivation(context.Background(), &ns.GetDeviceActivationRequest{
		DevEUI: d.DevEUI[:],
	})
	if err != nil {
		return fmt.Errorf("get device activation error: %s", err)
	}

	b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, actRes.FCntDown, qi.Data)
	if err != nil {
		return fmt.Errorf("encrypt frmpayload error: %s", err)
	}

	_, err = common.NetworkServer.SendDownlinkData(context.Background(), &ns.SendDownlinkDataRequest{
		DevEUI:    qi.DevEUI[:],
		Data:      b,
		Confirmed: qi.Confirmed,
		FPort:     uint32(qi.FPort),
		FCnt:      actRes.FCntDown,
	})
	if err != nil {
		return fmt.Errorf("push data-down error: %s", err)
	}
	return nil
}
