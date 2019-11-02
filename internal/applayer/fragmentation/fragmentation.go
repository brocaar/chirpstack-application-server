package fragmentation

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/downlink"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/applayer/fragmentation"
)

var (
	syncInterval  time.Duration
	syncRetries   int
	syncBatchSize int
)

// Setup configures the package.
func Setup(conf config.Config) error {
	syncInterval = conf.ApplicationServer.FragmentationSession.SyncInterval
	syncBatchSize = conf.ApplicationServer.FragmentationSession.SyncBatchSize
	syncRetries = conf.ApplicationServer.FragmentationSession.SyncRetries

	go SyncRemoteFragmentationSessionsLoop()

	return nil
}

// SyncRemoteFragmentationSessions syncs the fragmentation sessions with the devices.
func SyncRemoteFragmentationSessionsLoop() {
	for {
		ctxID, err := uuid.NewV4()
		if err != nil {
			log.WithError(err).Error("new uuid error")
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, logging.ContextIDKey, ctxID)

		err = storage.Transaction(func(tx sqlx.Ext) error {
			return syncRemoteFragmentationSessions(ctx, tx)
		})
		if err != nil {
			log.WithError(err).Error("sync remote fragmentation setup error")
		}
		time.Sleep(syncInterval)
	}
}

// HandleRemoteFragmentationSessionCommand handles an uplink fragmentation session command.
func HandleRemoteFragmentationSessionCommand(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64, b []byte) error {
	var cmd fragmentation.Command

	if err := cmd.UnmarshalBinary(true, b); err != nil {
		return errors.Wrap(err, "unmarshal command error")
	}

	switch cmd.CID {
	case fragmentation.FragSessionSetupAns:
		pl, ok := cmd.Payload.(*fragmentation.FragSessionSetupAnsPayload)
		if !ok {
			return fmt.Errorf("expected *fragmentation.FragSessionSetupAnsPayload, got: %T", cmd.Payload)
		}
		if err := handleFragSessionSetupAns(ctx, db, devEUI, pl); err != nil {
			return errors.Wrap(err, "handle FragSessionSetupAns error")
		}
	case fragmentation.FragSessionDeleteAns:
		pl, ok := cmd.Payload.(*fragmentation.FragSessionDeleteAnsPayload)
		if !ok {
			return fmt.Errorf("expected *fragmentation.FragSessionDeleteAnsPayload, got: %T", cmd.Payload)
		}
		if err := handleFragSessionDeleteAns(ctx, db, devEUI, pl); err != nil {
			return errors.Wrap(err, "handle FragSessionDeleteAns error")
		}
	case fragmentation.FragSessionStatusAns:
		pl, ok := cmd.Payload.(*fragmentation.FragSessionStatusAnsPayload)
		if !ok {
			return fmt.Errorf("expected *fragmentation.FragSessionStatusAns, got: %T", cmd.Payload)
		}
		if err := handleFragSessionStatusAns(ctx, db, devEUI, pl); err != nil {
			return errors.Wrap(err, "handle FragSessionStatusAns error")
		}
	default:
		return fmt.Errorf("CID not implemented: %s", cmd.CID)
	}

	return nil
}

func syncRemoteFragmentationSessions(ctx context.Context, db sqlx.Ext) error {
	items, err := storage.GetPendingRemoteFragmentationSessions(ctx, db, syncBatchSize, syncRetries)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := syncRemoteFragmentationSession(ctx, db, item); err != nil {
			return errors.Wrap(err, "sync remote fragmentation session error")
		}
	}

	return nil
}

func syncRemoteFragmentationSession(ctx context.Context, db sqlx.Ext, item storage.RemoteFragmentationSession) error {
	var cmd fragmentation.Command

	switch item.State {
	case storage.RemoteMulticastSetupSetup:
		pl := fragmentation.FragSessionSetupReqPayload{
			FragSession: fragmentation.FragSessionSetupReqPayloadFragSession{
				FragIndex: uint8(item.FragIndex),
			},
			NbFrag:   uint16(item.NbFrag),
			FragSize: uint8(item.FragSize),
			Control: fragmentation.FragSessionSetupReqPayloadControl{
				FragmentationMatrix: item.FragmentationMatrix,
				BlockAckDelay:       uint8(item.BlockAckDelay),
			},
			Padding:    uint8(item.Padding),
			Descriptor: item.Descriptor,
		}

		for _, idx := range item.MCGroupIDs {
			if idx <= 3 {
				pl.FragSession.McGroupBitMask[idx] = true
			}
		}

		cmd = fragmentation.Command{
			CID:     fragmentation.FragSessionSetupReq,
			Payload: &pl,
		}
	case storage.RemoteMulticastSetupDelete:
		cmd = fragmentation.Command{
			CID: fragmentation.FragSessionDeleteReq,
			Payload: &fragmentation.FragSessionDeleteReqPayload{
				Param: fragmentation.FragSessionDeleteReqPayloadParam{
					FragIndex: uint8(item.FragIndex),
				},
			},
		}
	default:
		return fmt.Errorf("invalid state: %s", item.State)
	}

	b, err := cmd.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "marshal binary error")
	}

	_, err = downlink.EnqueueDownlinkPayload(ctx, db, item.DevEUI, false, fragmentation.DefaultFPort, b)
	if err != nil {
		return errors.Wrap(err, "enqueue downlink payload error")
	}

	log.WithFields(log.Fields{
		"dev_eui":    item.DevEUI,
		"frag_index": item.FragIndex,
		"ctx_id":     ctx.Value(logging.ContextIDKey),
	}).Infof("%s enqueued", cmd.CID)

	item.RetryCount++
	item.RetryAfter = time.Now().Add(item.RetryInterval)

	err = storage.UpdateRemoteFragmentationSession(ctx, db, &item)
	if err != nil {
		return errors.Wrap(err, "update remote fragmentation session error")
	}

	return nil
}

func handleFragSessionSetupAns(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64, pl *fragmentation.FragSessionSetupAnsPayload) error {
	log.WithFields(log.Fields{
		"dev_eui":                          devEUI,
		"frag_index":                       pl.StatusBitMask.FragIndex,
		"wrong_descriptor":                 pl.StatusBitMask.WrongDescriptor,
		"frag_session_index_not_supported": pl.StatusBitMask.FragSessionIndexNotSupported,
		"not_enough_memory":                pl.StatusBitMask.NotEnoughMemory,
		"encoding_unsupported":             pl.StatusBitMask.EncodingUnsupported,
		"ctx_id":                           ctx.Value(logging.ContextIDKey),
	}).Info("FragSessionSetupAns received")

	if pl.StatusBitMask.WrongDescriptor || pl.StatusBitMask.FragSessionIndexNotSupported || pl.StatusBitMask.NotEnoughMemory || pl.StatusBitMask.EncodingUnsupported {
		return fmt.Errorf("WrongDescriptor: %t, FragSessionIndexNotSupported: %t, NotEnoughMemory: %t, EncodingUnsupported: %t", pl.StatusBitMask.WrongDescriptor, pl.StatusBitMask.FragSessionIndexNotSupported, pl.StatusBitMask.NotEnoughMemory, pl.StatusBitMask.EncodingUnsupported)
	}

	rfs, err := storage.GetRemoteFragmentationSession(ctx, db, devEUI, int(pl.StatusBitMask.FragIndex), true)
	if err != nil {
		return errors.Wrap(err, "get remote fragmentation session error")
	}

	rfs.StateProvisioned = true
	if err := storage.UpdateRemoteFragmentationSession(ctx, db, &rfs); err != nil {
		return errors.Wrap(err, "update remote fragmentation session error")
	}

	return nil
}

func handleFragSessionDeleteAns(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64, pl *fragmentation.FragSessionDeleteAnsPayload) error {
	log.WithFields(log.Fields{
		"dev_eui":                devEUI,
		"frag_index":             pl.Status.FragIndex,
		"session_does_not_exist": pl.Status.SessionDoesNotExist,
		"ctx_id":                 ctx.Value(logging.ContextIDKey),
	}).Info("FragSessionDeleteAns received")

	if pl.Status.SessionDoesNotExist {
		return fmt.Errorf("FragIndex %d does not exist", pl.Status.FragIndex)
	}

	rfs, err := storage.GetRemoteFragmentationSession(ctx, db, devEUI, int(pl.Status.FragIndex), true)
	if err != nil {
		return errors.Wrap(err, "get remove fragmentation session error")
	}

	rfs.StateProvisioned = true
	if err := storage.UpdateRemoteFragmentationSession(ctx, db, &rfs); err != nil {
		return errors.Wrap(err, "update remote fragmentation session error")
	}

	return nil
}

func handleFragSessionStatusAns(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64, pl *fragmentation.FragSessionStatusAnsPayload) error {
	log.WithFields(log.Fields{
		"dev_eui":                  devEUI,
		"frag_index":               pl.ReceivedAndIndex.FragIndex,
		"missing_frag":             pl.MissingFrag,
		"nb_frag_received":         pl.ReceivedAndIndex.NbFragReceived,
		"not_enough_matrix_memory": pl.Status.NotEnoughMatrixMemory,
		"ctx_id":                   ctx.Value(logging.ContextIDKey),
	}).Info("FragSessionStatusAns received")

	fdd, err := storage.GetPendingFUOTADeploymentDevice(ctx, db, devEUI)
	if err != nil {
		return errors.Wrap(err, "get pending fuota deployment device error")
	}

	fdd.State = storage.FUOTADeploymentDeviceSuccess

	if pl.MissingFrag > 0 {
		fdd.State = storage.FUOTADeploymentDeviceError
		fdd.ErrorMessage = fmt.Sprintf("%d fragments missed (%d received).", pl.MissingFrag, pl.ReceivedAndIndex.NbFragReceived)
	}

	if pl.Status.NotEnoughMatrixMemory {
		fdd.State = storage.FUOTADeploymentDeviceError
		fdd.ErrorMessage = "Not enough matrix memory."
	}

	err = storage.UpdateFUOTADeploymentDevice(ctx, db, &fdd)
	if err != nil {
		return errors.Wrap(err, "update fuota deployment device error")
	}

	return nil
}
