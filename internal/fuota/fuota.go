package fuota

import (
	"context"
	"crypto/aes"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/downlink"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/chirpstack-application-server/internal/multicast"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-api/go/ns"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/applayer/fragmentation"
	"github.com/brocaar/lorawan/applayer/multicastsetup"
)

var (
	interval                          = time.Second
	batchSize                         = 1
	mcGroupID                         int
	fragIndex                         int
	remoteMulticastSetupRetries       int
	remoteFragmentationSessionRetries int
	routingProfileID                  uuid.UUID
)

// Setup configures the package.
func Setup(conf config.Config) error {
	var err error

	routingProfileID, err = uuid.FromString(conf.ApplicationServer.ID)
	if err != nil {
		return errors.Wrap(err, "application-server id to uuid error")
	}

	mcGroupID = conf.ApplicationServer.FUOTADeployment.McGroupID
	fragIndex = conf.ApplicationServer.FUOTADeployment.FragIndex
	remoteMulticastSetupRetries = conf.ApplicationServer.RemoteMulticastSetup.SyncRetries
	remoteFragmentationSessionRetries = conf.ApplicationServer.FragmentationSession.SyncRetries

	go fuotaDeploymentLoop()

	return nil
}

func fuotaDeploymentLoop() {
	for {
		ctxID, err := uuid.NewV4()
		if err != nil {
			log.WithError(err).Error("new uuid error")
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, logging.ContextIDKey, ctxID)

		err = storage.Transaction(func(tx sqlx.Ext) error {
			return fuotaDeployments(ctx, tx)
		})
		if err != nil {
			log.WithError(err).Error("fuota deployment error")
		}
		time.Sleep(interval)
	}
}

func fuotaDeployments(ctx context.Context, db sqlx.Ext) error {
	items, err := storage.GetPendingFUOTADeployments(ctx, db, batchSize)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := fuotaDeployment(ctx, db, item); err != nil {
			return errors.Wrap(err, "fuota deployment error")
		}
	}

	return nil
}

func fuotaDeployment(ctx context.Context, db sqlx.Ext, item storage.FUOTADeployment) error {
	switch item.State {
	case storage.FUOTADeploymentMulticastCreate:
		return stepMulticastCreate(ctx, db, item)
	case storage.FUOTADeploymentMulticastSetup:
		return stepMulticastSetup(ctx, db, item)
	case storage.FUOTADeploymentFragmentationSessSetup:
		return stepFragmentationSessSetup(ctx, db, item)
	case storage.FUOTADeploymentMulticastSessCSetup:
		return stepMulticastSessCSetup(ctx, db, item)
	case storage.FUOTADeploymentEnqueue:
		return stepEnqueue(ctx, db, item)
	case storage.FUOTADeploymentStatusRequest:
		return stepStatusRequest(ctx, db, item)
	case storage.FUOTADeploymentSetDeviceStatus:
		return stepSetDeviceStatus(ctx, db, item)
	case storage.FUOTADeploymentCleanup:
		return stepCleanup(ctx, db, item)
	default:
		return fmt.Errorf("unexpected state: %s", item.State)
	}
}

func stepMulticastCreate(ctx context.Context, db sqlx.Ext, item storage.FUOTADeployment) error {
	var devAddr lorawan.DevAddr
	if _, err := rand.Read(devAddr[:]); err != nil {
		return errors.Wrap(err, "read random bytes error")
	}

	var mcKey lorawan.AES128Key
	if _, err := rand.Read(mcKey[:]); err != nil {
		return errors.Wrap(err, "read random bytes error")
	}

	mcAppSKey, err := multicastsetup.GetMcAppSKey(mcKey, devAddr)
	if err != nil {
		return errors.Wrap(err, "get McAppSKey error")
	}

	mcNetSKey, err := multicastsetup.GetMcNetSKey(mcKey, devAddr)
	if err != nil {
		return errors.Wrap(err, "get McNetSKey error")
	}

	spID, err := storage.GetServiceProfileIDForFUOTADeployment(ctx, db, item.ID)
	if err != nil {
		return errors.Wrap(err, "get service-profile for fuota deployment error")
	}

	mg := storage.MulticastGroup{
		Name:             fmt.Sprintf("fuota-%s", item.ID),
		MCAppSKey:        mcAppSKey,
		MCKey:            mcKey,
		ServiceProfileID: spID,
		MulticastGroup: ns.MulticastGroup{
			McAddr:           devAddr[:],
			McNwkSKey:        mcNetSKey[:],
			FCnt:             0,
			Dr:               uint32(item.DR),
			Frequency:        uint32(item.Frequency),
			PingSlotPeriod:   uint32(item.PingSlotPeriod),
			ServiceProfileId: spID.Bytes(),
			RoutingProfileId: routingProfileID.Bytes(),
		},
	}

	switch item.GroupType {
	case storage.FUOTADeploymentGroupTypeB:
		mg.MulticastGroup.GroupType = ns.MulticastGroupType_CLASS_B
	case storage.FUOTADeploymentGroupTypeC:
		mg.MulticastGroup.GroupType = ns.MulticastGroupType_CLASS_C
	default:
		return fmt.Errorf("unkonwn group-type: %s", item.GroupType)
	}

	err = storage.CreateMulticastGroup(ctx, db, &mg)
	if err != nil {
		return errors.Wrap(err, "create multicast-group error")
	}

	var mgID uuid.UUID
	copy(mgID[:], mg.MulticastGroup.Id)

	item.MulticastGroupID = &mgID
	item.State = storage.FUOTADeploymentMulticastSetup
	item.NextStepAfter = time.Now()

	err = storage.UpdateFUOTADeployment(ctx, db, &item)
	if err != nil {
		return errors.Wrap(err, "update fuota deployment error")
	}

	return nil
}

func stepMulticastSetup(ctx context.Context, db sqlx.Ext, item storage.FUOTADeployment) error {
	if item.MulticastGroupID == nil {
		return errors.New("MulticastGroupID must not be nil")
	}

	mcg, err := storage.GetMulticastGroup(ctx, db, *item.MulticastGroupID, false, false)
	if err != nil {
		return errors.Wrap(err, "get multicast group error")
	}

	// query all device-keys that relate to this FUOTA deployment
	var deviceKeys []storage.DeviceKeys
	err = sqlx.Select(db, &deviceKeys, `
		select
			dk.*
		from
			fuota_deployment_device dd
		inner join
			device_keys dk
			on dd.dev_eui = dk.dev_eui
		where
			dd.fuota_deployment_id = $1`,
		item.ID,
	)
	if err != nil {
		return errors.Wrap(err, "sql select error")
	}

	for _, dk := range deviceKeys {
		var nullKey lorawan.AES128Key

		// get the encrypted McKey.
		var mcKeyEncrypted, mcRootKey lorawan.AES128Key
		if dk.AppKey != nullKey {
			mcRootKey, err = multicastsetup.GetMcRootKeyForAppKey(dk.AppKey)
			if err != nil {
				return errors.Wrap(err, "get McRootKey for AppKey error")
			}
		} else {
			mcRootKey, err = multicastsetup.GetMcRootKeyForGenAppKey(dk.GenAppKey)
			if err != nil {
				return errors.Wrap(err, "get McRootKey for GenAppKey error")
			}
		}

		mcKEKey, err := multicastsetup.GetMcKEKey(mcRootKey)
		if err != nil {
			return errors.Wrap(err, "get McKEKey error")
		}

		block, err := aes.NewCipher(mcKEKey[:])
		if err != nil {
			return errors.Wrap(err, "new cipher error")
		}
		block.Decrypt(mcKeyEncrypted[:], mcg.MCKey[:])

		// create remote multicast setup record for device
		rms := storage.RemoteMulticastSetup{
			DevEUI:           dk.DevEUI,
			MulticastGroupID: *item.MulticastGroupID,
			McGroupID:        mcGroupID,
			McKeyEncrypted:   mcKeyEncrypted,
			MinMcFCnt:        0,
			MaxMcFCnt:        (1 << 32) - 1,
			State:            storage.RemoteMulticastSetupSetup,
			RetryInterval:    item.UnicastTimeout,
		}
		copy(rms.McAddr[:], mcg.MulticastGroup.McAddr)

		err = storage.CreateRemoteMulticastSetup(ctx, db, &rms)
		if err != nil {
			return errors.Wrap(err, "create remote multicast setup error")
		}
	}

	item.State = storage.FUOTADeploymentFragmentationSessSetup
	item.NextStepAfter = time.Now().Add(time.Duration(remoteMulticastSetupRetries) * item.UnicastTimeout)

	err = storage.UpdateFUOTADeployment(ctx, db, &item)
	if err != nil {
		return errors.Wrap(err, "update fuota deployment error")
	}

	return nil
}

func stepFragmentationSessSetup(ctx context.Context, db sqlx.Ext, item storage.FUOTADeployment) error {
	if item.MulticastGroupID == nil {
		return errors.New("MulticastGroupID must not be nil")
	}

	if item.FragSize == 0 {
		return errors.New("FragSize must not be 0")
	}

	// query all devices with complete multicast setup
	var devEUIs []lorawan.EUI64
	err := sqlx.Select(db, &devEUIs, `
		select
			dev_eui
		from
			remote_multicast_setup
		where
			multicast_group_id = $1
			and state = $2
			and state_provisioned = $3`,
		item.MulticastGroupID,
		storage.RemoteMulticastSetupSetup,
		true,
	)
	if err != nil {
		return errors.Wrap(err, "get devices with multicast setup error")
	}

	padding := (item.FragSize - (len(item.Payload) % item.FragSize)) % item.FragSize
	nbFrag := ((len(item.Payload) + padding) / item.FragSize)

	for _, devEUI := range devEUIs {
		// delete existing fragmentation session if it exist
		err = storage.DeleteRemoteFragmentationSession(ctx, db, devEUI, fragIndex)
		if err != nil && err != storage.ErrDoesNotExist {
			return errors.Wrap(err, "delete remote fragmentation session error")
		}

		fs := storage.RemoteFragmentationSession{
			DevEUI:              devEUI,
			FragIndex:           fragIndex,
			MCGroupIDs:          []int{mcGroupID},
			NbFrag:              nbFrag,
			FragSize:            item.FragSize,
			FragmentationMatrix: item.FragmentationMatrix,
			BlockAckDelay:       item.BlockAckDelay,
			Padding:             padding,
			Descriptor:          item.Descriptor,
			State:               storage.RemoteMulticastSetupSetup,
			RetryInterval:       item.UnicastTimeout,
		}
		err = storage.CreateRemoteFragmentationSession(ctx, db, &fs)
		if err != nil {
			return errors.Wrap(err, "create remote fragmentation session error")
		}
	}

	item.State = storage.FUOTADeploymentMulticastSessCSetup
	item.NextStepAfter = time.Now().Add(time.Duration(remoteFragmentationSessionRetries) * item.UnicastTimeout)

	err = storage.UpdateFUOTADeployment(ctx, db, &item)
	if err != nil {
		return errors.Wrap(err, "update fuota deployment error")
	}

	return nil
}

func stepMulticastSessCSetup(ctx context.Context, db sqlx.Ext, item storage.FUOTADeployment) error {
	if item.MulticastGroupID == nil {
		return errors.New("MulticastGroupID must not be nil")
	}

	mcg, err := storage.GetMulticastGroup(ctx, db, *item.MulticastGroupID, false, false)
	if err != nil {
		return errors.Wrap(err, "get multicast group error")
	}

	// query all devices with complete fragmentation session setup
	var devEUIs []lorawan.EUI64
	err = sqlx.Select(db, &devEUIs, `
		select
			rms.dev_eui
		from
			remote_multicast_setup rms
		inner join
			remote_fragmentation_session rfs
		on
			rfs.dev_eui = rms.dev_eui
			and rfs.frag_index = $1
		where
			rms.multicast_group_id = $2
			and rms.state = $3
			and rms.state_provisioned = $4
			and rfs.state = $3
			and rms.state_provisioned = $4`,
		fragIndex,
		item.MulticastGroupID,
		storage.RemoteMulticastSetupSetup,
		true,
	)
	if err != nil {
		return errors.Wrap(err, "get devices with fragmentation session setup error")
	}

	for _, devEUI := range devEUIs {
		rmccs := storage.RemoteMulticastClassCSession{
			DevEUI:           devEUI,
			MulticastGroupID: *item.MulticastGroupID,
			McGroupID:        mcGroupID,
			DLFrequency:      int(mcg.MulticastGroup.Frequency),
			DR:               int(mcg.MulticastGroup.Dr),
			SessionTime:      time.Now().Add(time.Duration(remoteMulticastSetupRetries) * item.UnicastTimeout),
			SessionTimeOut:   item.MulticastTimeout,
			RetryInterval:    item.UnicastTimeout,
		}
		err = storage.CreateRemoteMulticastClassCSession(ctx, db, &rmccs)
		if err != nil {
			return errors.Wrap(err, "create remote multicast class-c session error")
		}
	}

	item.State = storage.FUOTADeploymentEnqueue
	item.NextStepAfter = time.Now().Add(time.Duration(remoteMulticastSetupRetries) * item.UnicastTimeout)

	err = storage.UpdateFUOTADeployment(ctx, db, &item)
	if err != nil {
		return errors.Wrap(err, "update fuota deployment error")
	}

	return nil
}

func stepEnqueue(ctx context.Context, db sqlx.Ext, item storage.FUOTADeployment) error {
	if item.MulticastGroupID == nil {
		return errors.New("MulticastGroupID must not be nil")
	}

	// fragment the payload
	padding := (item.FragSize - (len(item.Payload) % item.FragSize)) % item.FragSize
	fragments, err := fragmentation.Encode(append(item.Payload, make([]byte, padding)...), item.FragSize, item.Redundancy)
	if err != nil {
		return errors.Wrap(err, "fragment payload error")
	}

	// wrap the payloads into data-fragment payloads
	var payloads [][]byte
	for i := range fragments {
		cmd := fragmentation.Command{
			CID: fragmentation.DataFragment,
			Payload: &fragmentation.DataFragmentPayload{
				IndexAndN: fragmentation.DataFragmentPayloadIndexAndN{
					FragIndex: uint8(fragIndex),
					N:         uint16(i + 1),
				},
				Payload: fragments[i],
			},
		}
		b, err := cmd.MarshalBinary()
		if err != nil {
			return errors.Wrap(err, "marshal binary error")
		}

		payloads = append(payloads, b)
	}

	// enqueue the payloads
	_, err = multicast.EnqueueMultiple(ctx, db, *item.MulticastGroupID, fragmentation.DefaultFPort, payloads)
	if err != nil {
		return errors.Wrap(err, "enqueue multiple error")
	}

	item.State = storage.FUOTADeploymentStatusRequest

	switch item.GroupType {
	case storage.FUOTADeploymentGroupTypeC:
		item.NextStepAfter = time.Now().Add(time.Second * time.Duration(1<<uint(item.MulticastTimeout)))
	default:
		return fmt.Errorf("group-type not implemented: %s", item.GroupType)
	}

	err = storage.UpdateFUOTADeployment(ctx, db, &item)
	if err != nil {
		return errors.Wrap(err, "update fuota deployment error")
	}

	return nil
}

func stepStatusRequest(ctx context.Context, db sqlx.Ext, item storage.FUOTADeployment) error {
	if item.MulticastGroupID == nil {
		return errors.New("MulticastGroupID must not be nil")
	}

	// query all devices with complete fragmentation session setup
	var devEUIs []lorawan.EUI64
	err := sqlx.Select(db, &devEUIs, `
		select
			rms.dev_eui
		from
			remote_multicast_setup rms
		inner join
			remote_fragmentation_session rfs
		on
			rfs.dev_eui = rms.dev_eui
			and rfs.frag_index = $1
		where
			rms.multicast_group_id = $2
			and rms.state = $3
			and rms.state_provisioned = $4
			and rfs.state = $3
			and rfs.state_provisioned = $4`,
		fragIndex,
		item.MulticastGroupID,
		storage.RemoteMulticastSetupSetup,
		true,
	)
	if err != nil {
		return errors.Wrap(err, "get devices with fragmentation session setup error")
	}

	for _, devEUI := range devEUIs {
		cmd := fragmentation.Command{
			CID: fragmentation.FragSessionStatusReq,
			Payload: &fragmentation.FragSessionStatusReqPayload{
				FragStatusReqParam: fragmentation.FragSessionStatusReqPayloadFragStatusReqParam{
					FragIndex:    uint8(fragIndex),
					Participants: true,
				},
			},
		}
		b, err := cmd.MarshalBinary()
		if err != nil {
			return errors.Wrap(err, "marshal binary error")
		}

		_, err = downlink.EnqueueDownlinkPayload(ctx, db, devEUI, false, fragmentation.DefaultFPort, b)
		if err != nil {
			return errors.Wrap(err, "enqueue downlink payload error")
		}
	}

	item.State = storage.FUOTADeploymentSetDeviceStatus
	item.NextStepAfter = time.Now().Add(item.UnicastTimeout)

	err = storage.UpdateFUOTADeployment(ctx, db, &item)
	if err != nil {
		return errors.Wrap(err, "update fuota deployment error")
	}

	return nil
}

func stepSetDeviceStatus(ctx context.Context, db sqlx.Ext, item storage.FUOTADeployment) error {
	if item.MulticastGroupID == nil {
		return errors.New("MulticastGroupID must not be nil")
	}

	// set remote multicast session error
	_, err := db.Exec(`
		update
			fuota_deployment_device fdd
		set
			state = $5,
			error_message = $6
		from
			remote_multicast_setup rms
		where
			fdd.fuota_deployment_id = $1
			and rms.multicast_group_id = $2

			and fdd.state = $3
			and rms.state_provisioned = $4

			-- join the two tables
			and fdd.dev_eui = rms.dev_eui`,

		item.ID,
		*item.MulticastGroupID,
		storage.FUOTADeploymentDevicePending,
		false,
		storage.FUOTADeploymentDeviceError,
		"The device failed to provision the remote multicast setup.",
	)
	if err != nil {
		return errors.Wrap(err, "set remote multicast setup error error")
	}

	// set remote fragmentation session error
	_, err = db.Exec(`
		update
			fuota_deployment_device fdd
		set
			state = $5,
			error_message = $6
		from
			remote_fragmentation_session rfs
		where
			fdd.fuota_deployment_id = $1
			and rfs.frag_index = $2

			and fdd.state = $3
			and rfs.state_provisioned = $4

			-- join the two tables
			and fdd.dev_eui = rfs.dev_eui`,
		item.ID,
		fragIndex,
		storage.FUOTADeploymentDevicePending,
		false,
		storage.FUOTADeploymentDeviceError,
		"The device failed to provision the fragmentation session setup.",
	)
	if err != nil {
		return errors.Wrap(err, "set fragmentation session setup error error")
	}

	// set remaining errors
	_, err = db.Exec(`
		update
			fuota_deployment_device
		set
			state = $3,
			error_message = $4
		where
			fuota_deployment_id = $1
			and state = $2`,
		item.ID,
		storage.FUOTADeploymentDevicePending,
		storage.FUOTADeploymentDeviceError,
		"Device did not complete the FUOTA deployment or did not confirm that it completed the FUOTA deployment.",
	)
	if err != nil {
		return errors.Wrap(err, "set incomplete fuota deployment error")
	}

	item.State = storage.FUOTADeploymentCleanup
	item.NextStepAfter = time.Now()

	err = storage.UpdateFUOTADeployment(ctx, db, &item)
	if err != nil {
		return errors.Wrap(err, "update fuota deployment error")
	}

	return nil
}

func stepCleanup(ctx context.Context, db sqlx.Ext, item storage.FUOTADeployment) error {
	if item.MulticastGroupID != nil {
		if err := storage.DeleteMulticastGroup(ctx, db, *item.MulticastGroupID); err != nil {
			return errors.Wrap(err, "delete multicast group error")
		}
	}

	item.MulticastGroupID = nil
	item.State = storage.FUOTADeploymentDone

	err := storage.UpdateFUOTADeployment(ctx, db, &item)
	if err != nil {
		return errors.Wrap(err, "update fuota deployment error")
	}

	return nil
}
