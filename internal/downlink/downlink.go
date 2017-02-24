package downlink

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

// HandleDataDownPayloads handles received downlink payloads to be emitted to the
// nodes.
func HandleDataDownPayloads(ctx common.Context, plChan chan handler.DataDownPayload) {
	for pl := range plChan {
		go func(pl handler.DataDownPayload) {
			if err := handleDataDownPayload(ctx, pl); err != nil {
				log.WithFields(log.Fields{
					"dev_eui":        pl.DevEUI,
					"application_id": pl.ApplicationID,
					"reference":      pl.Reference,
				}).Errorf("handle data-down payload error: %s", err)
			}
		}(pl)
	}
}

func handleDataDownPayload(ctx common.Context, pl handler.DataDownPayload) error {
	node, err := storage.GetNode(ctx.DB, pl.DevEUI)
	if err != nil {
		return fmt.Errorf("get node error: %s", err)
	}

	// Validate that the ApplicationID matches the actual DevEUI.
	// This is needed as authorisation might be performed on MQTT topic level
	// where it is unknown if the given ApplicationID matches the given
	// DevEUI.
	if node.ApplicationID != pl.ApplicationID {
		return errors.New("enqueue data-down payload: node does not exist for given application")
	}

	qi := storage.DownlinkQueueItem{
		Reference: pl.Reference,
		DevEUI:    pl.DevEUI,
		Confirmed: pl.Confirmed,
		FPort:     pl.FPort,
		Data:      pl.Data,
	}

	return HandleDownlinkQueueItem(ctx, node, &qi)
}

// HandleDownlinkQueueItem handles a DownlinkQueueItem to be emitted to the node.
// In case of class-c, it will send the payload directly to the network-server.
// In any other case, it will be enqueued.
func HandleDownlinkQueueItem(ctx common.Context, node storage.Node, qi *storage.DownlinkQueueItem) error {
	if node.IsClassC && qi.Confirmed {
		qi.Pending = true
	}

	// In case of a class-c device, we directly push the payload to the
	// network-server.
	// Before pushing, we purge the queue to make sure we have always a single
	// item in the queue in case of confirmed data.
	if node.IsClassC {
		if err := storage.DeleteDownlinkQueueItemsForDevEUI(ctx.DB, node.DevEUI); err != nil {
			return err
		}

		if err := pushDataDown(ctx, node, qi); err != nil {
			return err
		}
	}

	// save the queue-item in every case, except when the node is a class-c
	// device and the data is unconfirmed.
	if !(node.IsClassC && !qi.Confirmed) {
		if err := storage.CreateDownlinkQueueItem(ctx.DB, qi); err != nil {
			return fmt.Errorf("create downlink queue item error: %s", err)
		}
	}

	return nil
}

func pushDataDown(ctx common.Context, node storage.Node, qi *storage.DownlinkQueueItem) error {
	nsResp, err := ctx.NetworkServer.GetNodeSession(context.Background(), &ns.GetNodeSessionRequest{
		DevEUI: node.DevEUI[:],
	})
	if err != nil {
		return fmt.Errorf("get node session error: %s", err)
	}

	b, err := lorawan.EncryptFRMPayload(node.AppSKey, false, node.DevAddr, nsResp.FCntDown, qi.Data)
	if err != nil {
		return fmt.Errorf("encrypt frmpayload error: %s", err)
	}

	_, err = ctx.NetworkServer.PushDataDown(context.Background(), &ns.PushDataDownRequest{
		DevEUI:    qi.DevEUI[:],
		Data:      b,
		Confirmed: qi.Confirmed,
		FPort:     uint32(qi.FPort),
		FCnt:      nsResp.FCntDown,
	})
	if err != nil {
		return fmt.Errorf("push data-down error: %s", err)
	}
	return nil
}
