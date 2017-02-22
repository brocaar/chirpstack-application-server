package downlink

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lora-app-server/internal/storage"
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

	if node.ApplicationID != pl.ApplicationID {
		return errors.New("enqueue data-down payload: node does not exist for given application")
	}

	queueItem := storage.DownlinkQueueItem{
		Reference: pl.Reference,
		DevEUI:    pl.DevEUI,
		Confirmed: pl.Confirmed,
		FPort:     pl.FPort,
		Data:      pl.Data,
	}
	err = storage.CreateDownlinkQueueItem(ctx.DB, &queueItem)
	if err != nil {
		return fmt.Errorf("create downlink queue item error: %s", err)
	}

	return nil
}
