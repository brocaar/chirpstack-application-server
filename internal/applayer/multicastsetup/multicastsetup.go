package multicastsetup

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
	"github.com/brocaar/lorawan/applayer/multicastsetup"
	"github.com/brocaar/lorawan/gps"
)

var (
	syncInterval  time.Duration
	syncRetries   int
	syncBatchSize int
)

// Setup configures the package.
func Setup(conf config.Config) error {
	syncInterval = conf.ApplicationServer.RemoteMulticastSetup.SyncInterval
	syncBatchSize = conf.ApplicationServer.RemoteMulticastSetup.SyncBatchSize
	syncRetries = conf.ApplicationServer.RemoteMulticastSetup.SyncRetries

	go SyncRemoteMulticastSetupLoop()
	go SyncRemoteMulticastClassCSessionLoop()

	return nil
}

// SyncRemoteMulticastSetupLoop syncs the multicast setup with the devices.
func SyncRemoteMulticastSetupLoop() {
	for {
		ctxID, err := uuid.NewV4()
		if err != nil {
			log.WithError(err).Error("new uuid error")
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, logging.ContextIDKey, ctxID)

		err = storage.Transaction(func(tx sqlx.Ext) error {
			return syncRemoteMulticastSetup(ctx, tx)
		})

		if err != nil {
			log.WithError(err).Error("sync remote multicast setup error")
		}
		time.Sleep(syncInterval)
	}
}

// SyncRemoteMulticastClassCSessionLoop syncs the multicast Class-C session
// with the devices.
func SyncRemoteMulticastClassCSessionLoop() {
	for {
		ctxID, err := uuid.NewV4()
		if err != nil {
			log.WithError(err).Error("new uuid error")
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, logging.ContextIDKey, ctxID)

		err = storage.Transaction(func(tx sqlx.Ext) error {
			return syncRemoteMulticastClassCSession(ctx, tx)
		})

		if err != nil {
			log.WithError(err).Error("sync remote multicast class-c session error")
		}
		time.Sleep(syncInterval)
	}
}

// HandleRemoteMulticastSetupCommand handles an uplink remote multicast setup command.
func HandleRemoteMulticastSetupCommand(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64, b []byte) error {
	var cmd multicastsetup.Command

	if err := cmd.UnmarshalBinary(true, b); err != nil {
		return errors.Wrap(err, "unmarshal command error")
	}

	switch cmd.CID {
	case multicastsetup.McGroupSetupAns:
		pl, ok := cmd.Payload.(*multicastsetup.McGroupSetupAnsPayload)
		if !ok {
			return fmt.Errorf("expected *multicastsetup.McGroupSetupAnsPayload, got: %T", cmd.Payload)
		}
		if err := handleMcGroupSetupAns(ctx, db, devEUI, pl); err != nil {
			return errors.Wrap(err, "handle McGroupSetupAns error")
		}
	case multicastsetup.McGroupDeleteAns:
		pl, ok := cmd.Payload.(*multicastsetup.McGroupDeleteAnsPayload)
		if !ok {
			return fmt.Errorf("expected *multicastsetup.McGroupDeleteAnsPayload, got: %T", cmd.Payload)
		}
		if err := handleMcGroupDeleteAns(ctx, db, devEUI, pl); err != nil {
			return errors.Wrap(err, "handle McGroupDeleteAns error")
		}
	case multicastsetup.McClassCSessionAns:
		pl, ok := cmd.Payload.(*multicastsetup.McClassCSessionAnsPayload)
		if !ok {
			return fmt.Errorf("expected *multicastsetup.McClassCSessionAnsPayload, got: %T", cmd.Payload)
		}
		if err := handleMcClassCSessionAns(ctx, db, devEUI, pl); err != nil {
			return errors.Wrap(err, "handle McClassCSessionAns error")
		}
	default:
		return fmt.Errorf("CID not implemented: %s", cmd.CID)
	}

	return nil
}

func handleMcGroupSetupAns(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64, pl *multicastsetup.McGroupSetupAnsPayload) error {
	log.WithFields(log.Fields{
		"dev_eui":     devEUI,
		"id_error":    pl.McGroupIDHeader.IDError,
		"mc_group_id": pl.McGroupIDHeader.McGroupID,
		"ctx_id":      ctx.Value(logging.ContextIDKey),
	}).Info("McGroupSetupAns received")

	if pl.McGroupIDHeader.IDError {
		return fmt.Errorf("IDError for McGroupID: %d", pl.McGroupIDHeader.McGroupID)
	}

	rms, err := storage.GetRemoteMulticastSetupByGroupID(ctx, db, devEUI, int(pl.McGroupIDHeader.McGroupID), true)
	if err != nil {
		return errors.Wrap(err, "get remote multicast-setup by group id error")
	}

	rms.StateProvisioned = true
	if err := storage.UpdateRemoteMulticastSetup(ctx, db, &rms); err != nil {
		return errors.Wrap(err, "update remote multicast-setup error")
	}

	return nil
}

func handleMcGroupDeleteAns(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64, pl *multicastsetup.McGroupDeleteAnsPayload) error {
	log.WithFields(log.Fields{
		"dev_eui":            devEUI,
		"mc_group_id":        pl.McGroupIDHeader.McGroupID,
		"mc_group_undefined": pl.McGroupIDHeader.McGroupUndefined,
		"ctx_id":             ctx.Value(logging.ContextIDKey),
	}).Info("McGroupDeleteAns received")

	if pl.McGroupIDHeader.McGroupUndefined {
		return fmt.Errorf("McGroupUndefined for McGroupID: %d", pl.McGroupIDHeader.McGroupID)
	}

	rms, err := storage.GetRemoteMulticastSetupByGroupID(ctx, db, devEUI, int(pl.McGroupIDHeader.McGroupID), true)
	if err != nil {
		return errors.Wrap(err, "get remote multicast-setup by group id error")
	}

	rms.StateProvisioned = true
	if err := storage.UpdateRemoteMulticastSetup(ctx, db, &rms); err != nil {
		return errors.Wrap(err, "update remote multicast-setup error")
	}

	if err := storage.RemoveDeviceFromMulticastGroup(ctx, db, rms.MulticastGroupID, devEUI); err != nil {
		if err == storage.ErrDoesNotExist {
			log.WithFields(log.Fields{
				"dev_eui":            devEUI,
				"multicast_group_id": rms.MulticastGroupID,
				"ctx_id":             ctx.Value(logging.ContextIDKey),
			}).Info("applayer/multicastsetup: removing device from multicast group, but device does not exist")
		} else {
			return errors.Wrap(err, "remove device from multicast group error")
		}
	}

	return nil
}

func handleMcClassCSessionAns(ctx context.Context, db sqlx.Ext, devEUI lorawan.EUI64, pl *multicastsetup.McClassCSessionAnsPayload) error {
	log.WithFields(log.Fields{
		"dev_eui":            devEUI,
		"time_to_start":      pl.TimeToStart,
		"mc_group_undefined": pl.StatusAndMcGroupID.McGroupUndefined,
		"freq_error":         pl.StatusAndMcGroupID.FreqError,
		"dr_error":           pl.StatusAndMcGroupID.DRError,
		"mc_group_id":        pl.StatusAndMcGroupID.McGroupID,
		"ctx_id":             ctx.Value(logging.ContextIDKey),
	}).Info("McClassCSessionAns received")

	if pl.StatusAndMcGroupID.DRError || pl.StatusAndMcGroupID.FreqError || pl.StatusAndMcGroupID.McGroupUndefined {
		return fmt.Errorf("DRError: %t, FreqError: %t, McGroupUndefined: %t for McGroupID: %d", pl.StatusAndMcGroupID.DRError, pl.StatusAndMcGroupID.FreqError, pl.StatusAndMcGroupID.McGroupUndefined, pl.StatusAndMcGroupID.McGroupID)
	}

	sess, err := storage.GetRemoteMulticastClassCSessionByGroupID(ctx, db, devEUI, int(pl.StatusAndMcGroupID.McGroupID), true)
	if err != nil {
		return errors.Wrap(err, "get remote multicast class-c session error")
	}

	sess.StateProvisioned = true
	if err := storage.UpdateRemoteMulticastClassCSession(ctx, db, &sess); err != nil {
		return errors.Wrap(err, "update remote multicast class-c session error")
	}

	if err := storage.AddDeviceToMulticastGroup(ctx, db, sess.MulticastGroupID, devEUI); err != nil {
		if err == storage.ErrAlreadyExists {
			log.WithFields(log.Fields{
				"dev_eui":            devEUI,
				"multicast_group_id": sess.MulticastGroupID,
				"ctx_id":             ctx.Value(logging.ContextIDKey),
			}).Warning("applayer/multicastsetup: adding device to multicast group, but device was already added")
		} else {
			return errors.Wrap(err, "add device to multicast group error")
		}
	}

	return nil
}

func syncRemoteMulticastSetup(ctx context.Context, db sqlx.Ext) error {
	items, err := storage.GetPendingRemoteMulticastSetupItems(ctx, db, syncBatchSize, syncRetries)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := syncRemoteMulticastSetupItem(ctx, db, item); err != nil {
			return errors.Wrap(err, "sync remote multicast-setup error")
		}
	}

	return nil
}

func syncRemoteMulticastSetupItem(ctx context.Context, db sqlx.Ext, item storage.RemoteMulticastSetup) error {
	var cmd multicastsetup.Command

	switch item.State {
	case storage.RemoteMulticastSetupSetup:
		cmd = multicastsetup.Command{
			CID: multicastsetup.McGroupSetupReq,
			Payload: &multicastsetup.McGroupSetupReqPayload{
				McGroupIDHeader: multicastsetup.McGroupSetupReqPayloadMcGroupIDHeader{
					McGroupID: uint8(item.McGroupID),
				},
				McAddr:         item.McAddr,
				McKeyEncrypted: item.McKeyEncrypted,
				MinMcFCnt:      item.MinMcFCnt,
				MaxMcFCnt:      item.MaxMcFCnt,
			},
		}
	case storage.RemoteMulticastSetupDelete:
		cmd = multicastsetup.Command{
			CID: multicastsetup.McGroupDeleteReq,
			Payload: &multicastsetup.McGroupDeleteReqPayload{
				McGroupIDHeader: multicastsetup.McGroupDeleteReqPayloadMcGroupIDHeader{
					McGroupID: uint8(item.McGroupID),
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

	_, err = downlink.EnqueueDownlinkPayload(ctx, db, item.DevEUI, false, multicastsetup.DefaultFPort, b)
	if err != nil {
		return errors.Wrap(err, "enqueue downlink payload error")
	}

	log.WithFields(log.Fields{
		"dev_eui":     item.DevEUI,
		"mc_group_id": item.McGroupID,
		"ctx_id":      ctx.Value(logging.ContextIDKey),
	}).Infof("%s enqueued", cmd.CID)

	item.RetryCount++
	item.RetryAfter = time.Now().Add(item.RetryInterval)

	err = storage.UpdateRemoteMulticastSetup(ctx, db, &item)
	if err != nil {
		return errors.Wrap(err, "update remote multicast-setup error")
	}

	return nil
}

func syncRemoteMulticastClassCSession(ctx context.Context, db sqlx.Ext) error {
	items, err := storage.GetPendingRemoteMulticastClassCSessions(ctx, db, syncBatchSize, syncRetries)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := syncRemoteMulticastClassCSessionItem(ctx, db, item); err != nil {
			return errors.Wrap(err, "sync remote multicast class-c session error")
		}
	}

	return nil
}

func syncRemoteMulticastClassCSessionItem(ctx context.Context, db sqlx.Ext, item storage.RemoteMulticastClassCSession) error {
	cmd := multicastsetup.Command{
		CID: multicastsetup.McClassCSessionReq,
		Payload: &multicastsetup.McClassCSessionReqPayload{
			McGroupIDHeader: multicastsetup.McClassCSessionReqPayloadMcGroupIDHeader{
				McGroupID: uint8(item.McGroupID),
			},
			SessionTime: uint32((gps.Time(item.SessionTime).TimeSinceGPSEpoch() / time.Second) % (1 << 32)),
			SessionTimeOut: multicastsetup.McClassCSessionReqPayloadSessionTimeOut{
				TimeOut: uint8(item.SessionTimeOut),
			},
			DLFrequency: uint32(item.DLFrequency),
			DR:          uint8(item.DR),
		},
	}

	b, err := cmd.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "marshal binary error")
	}

	_, err = downlink.EnqueueDownlinkPayload(ctx, db, item.DevEUI, false, multicastsetup.DefaultFPort, b)
	if err != nil {
		return errors.Wrap(err, "enqueue downlink payload error")
	}

	log.WithFields(log.Fields{
		"dev_eui":     item.DevEUI,
		"mc_group_id": item.McGroupID,
	}).Infof("%s enqueued", cmd.CID)

	item.RetryCount++
	item.RetryAfter = time.Now().Add(item.RetryInterval)

	err = storage.UpdateRemoteMulticastClassCSession(ctx, db, &item)
	if err != nil {
		return errors.Wrap(err, "update remote multicast class-c session error")
	}

	return nil
}
