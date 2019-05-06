package fragmentation

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/downlink"
	"github.com/brocaar/lora-app-server/internal/storage"
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
		err := storage.Transaction(func(tx sqlx.Ext) error {
			return syncRemoteFragmentationSessions(tx)
		})
		if err != nil {
			log.WithError(err).Error("sync remote fragmentation setup error")
		}
		time.Sleep(syncInterval)
	}
}

// HandleRemoteFragmentationSessionCommand handles an uplink fragmentation session command.
func HandleRemoteFragmentationSessionCommand(db sqlx.Ext, devEUI lorawan.EUI64, b []byte) error {
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
		if err := handleFragSessionSetupAns(db, devEUI, pl); err != nil {
			return errors.Wrap(err, "handle FragSessionSetupAns error")
		}
	case fragmentation.FragSessionDeleteAns:
		pl, ok := cmd.Payload.(*fragmentation.FragSessionDeleteAnsPayload)
		if !ok {
			return fmt.Errorf("expected *fragmentation.FragSessionDeleteAnsPayload, got: %T", cmd.Payload)
		}
		if err := handleFragSessionDeleteAns(db, devEUI, pl); err != nil {
			return errors.Wrap(err, "handle FragSessionDeleteAns error")
		}
	case fragmentation.FragSessionStatusAns:
		pl, ok := cmd.Payload.(*fragmentation.FragSessionStatusAnsPayload)
		if !ok {
			return fmt.Errorf("expected *fragmentation.FragSessionStatusAns, got: %T", cmd.Payload)
		}
		if err := handleFragSessionStatusAns(db, devEUI, pl); err != nil {
			return errors.Wrap(err, "handle FragSessionStatusAns error")
		}
	default:
		return fmt.Errorf("CID not implemented: %s", cmd.CID)
	}

	return nil
}

func syncRemoteFragmentationSessions(db sqlx.Ext) error {
	items, err := storage.GetPendingRemoteFragmentationSessions(db, syncBatchSize, syncRetries)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := syncRemoteFragmentationSession(db, item); err != nil {
			return errors.Wrap(err, "sync remote fragmentation session error")
		}
	}

	return nil
}

func syncRemoteFragmentationSession(db sqlx.Ext, item storage.RemoteFragmentationSession) error {
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

	_, err = downlink.EnqueueDownlinkPayload(db, item.DevEUI, false, fragmentation.DefaultFPort, b)
	if err != nil {
		return errors.Wrap(err, "enqueue downlink payload error")
	}

	log.WithFields(log.Fields{
		"dev_eui":    item.DevEUI,
		"frag_index": item.FragIndex,
	}).Infof("%s enqueued", cmd.CID)

	item.RetryCount++
	item.RetryAfter = time.Now().Add(item.RetryInterval)

	err = storage.UpdateRemoteFragmentationSession(db, &item)
	if err != nil {
		return errors.Wrap(err, "update remote fragmentation session error")
	}

	return nil
}

func handleFragSessionSetupAns(db sqlx.Ext, devEUI lorawan.EUI64, pl *fragmentation.FragSessionSetupAnsPayload) error {
	log.WithFields(log.Fields{
		"dev_eui":                          devEUI,
		"frag_index":                       pl.StatusBitMask.FragIndex,
		"wrong_descriptor":                 pl.StatusBitMask.WrongDescriptor,
		"frag_session_index_not_supported": pl.StatusBitMask.FragSessionIndexNotSupported,
		"not_enough_memory":                pl.StatusBitMask.NotEnoughMemory,
		"encoding_unsupported":             pl.StatusBitMask.EncodingUnsupported,
	}).Info("FragSessionSetupAns received")

	if pl.StatusBitMask.WrongDescriptor || pl.StatusBitMask.FragSessionIndexNotSupported || pl.StatusBitMask.NotEnoughMemory || pl.StatusBitMask.EncodingUnsupported {
		return fmt.Errorf("WrongDescriptor: %t, FragSessionIndexNotSupported: %t, NotEnoughMemory: %t, EncodingUnsupported: %t", pl.StatusBitMask.WrongDescriptor, pl.StatusBitMask.FragSessionIndexNotSupported, pl.StatusBitMask.NotEnoughMemory, pl.StatusBitMask.EncodingUnsupported)
	}

	rfs, err := storage.GetRemoteFragmentationSession(db, devEUI, int(pl.StatusBitMask.FragIndex), true)
	if err != nil {
		return errors.Wrap(err, "get remote fragmentation session error")
	}

	rfs.StateProvisioned = true
	if err := storage.UpdateRemoteFragmentationSession(db, &rfs); err != nil {
		return errors.Wrap(err, "update remote fragmentation session error")
	}

	return nil
}

func handleFragSessionDeleteAns(db sqlx.Ext, devEUI lorawan.EUI64, pl *fragmentation.FragSessionDeleteAnsPayload) error {
	log.WithFields(log.Fields{
		"dev_eui":                devEUI,
		"frag_index":             pl.Status.FragIndex,
		"session_does_not_exist": pl.Status.SessionDoesNotExist,
	}).Info("FragSessionDeleteAns received")

	if pl.Status.SessionDoesNotExist {
		return fmt.Errorf("FragIndex %d does not exist", pl.Status.FragIndex)
	}

	rfs, err := storage.GetRemoteFragmentationSession(db, devEUI, int(pl.Status.FragIndex), true)
	if err != nil {
		return errors.Wrap(err, "get remove fragmentation session error")
	}

	rfs.StateProvisioned = true
	if err := storage.UpdateRemoteFragmentationSession(db, &rfs); err != nil {
		return errors.Wrap(err, "update remote fragmentation session error")
	}

	return nil
}

func handleFragSessionStatusAns(db sqlx.Ext, devEUI lorawan.EUI64, pl *fragmentation.FragSessionStatusAnsPayload) error {
	log.WithFields(log.Fields{
		"dev_eui":                  devEUI,
		"frag_index":               pl.ReceivedAndIndex.FragIndex,
		"missing_frag":             pl.MissingFrag,
		"nb_frag_received":         pl.ReceivedAndIndex.NbFragReceived,
		"not_enough_matrix_memory": pl.Status.NotEnoughMatrixMemory,
	}).Info("FragSessionStatusAns received")

	fdd, err := storage.GetPendingFUOTADeploymentDevice(db, devEUI)
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

	err = storage.UpdateFUOTADeploymentDevice(db, &fdd)
	if err != nil {
		return errors.Wrap(err, "update fuota deployment device error")
	}

	return nil
}
