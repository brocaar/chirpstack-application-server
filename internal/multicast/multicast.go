package multicast

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

// Enqueue adds the given payload to the multicast-group queue.
func Enqueue(db sqlx.Ext, multicastGroupID uuid.UUID, fPort uint8, data []byte) (uint32, error) {
	// Get and lock multicast-group, the lock is to make sure there are no
	// concurrent enqueue actions for the same multicast-group, which would
	// result in the re-use of the same frame-counter.
	mg, err := storage.GetMulticastGroup(db, multicastGroupID, true, false)
	if err != nil {
		return 0, errors.Wrap(err, "get multicast-group error")
	}

	// get network-server / client
	n, err := storage.GetNetworkServerForMulticastGroupID(db, multicastGroupID)
	if err != nil {
		return 0, errors.Wrap(err, "get network-server error")
	}
	nsClient, err := networkserver.GetPool().Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
	if err != nil {
		return 0, errors.Wrap(err, "get network-server client error")
	}

	var devAddr lorawan.DevAddr
	copy(devAddr[:], mg.MulticastGroup.McAddr)

	// encrypt payload
	b, err := lorawan.EncryptFRMPayload(mg.MCAppSKey, false, devAddr, mg.MulticastGroup.FCnt, data)
	if err != nil {
		return 0, errors.Wrap(err, "encrypt frmpayload error")
	}

	_, err = nsClient.EnqueueMulticastQueueItem(context.Background(), &ns.EnqueueMulticastQueueItemRequest{
		MulticastQueueItem: &ns.MulticastQueueItem{
			MulticastGroupId: multicastGroupID.Bytes(),
			FrmPayload:       b,
			FCnt:             mg.MulticastGroup.FCnt,
			FPort:            uint32(fPort),
		},
	})

	return mg.MulticastGroup.FCnt, nil
}
