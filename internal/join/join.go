package join

import (
	"crypto/aes"
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/handler"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

type context struct {
	joinReqPayload backend.JoinReqPayload
	joinAnsPayload backend.JoinAnsPayload
	phyPayload     lorawan.PHYPayload
	device         storage.Device
	application    storage.Application
	deviceKeys     storage.DeviceKeys
	appNonce       lorawan.AppNonce
	nwkSKey        lorawan.AES128Key
	appSKey        lorawan.AES128Key
	netID          lorawan.NetID
}

type task func(*context) error

type flow struct {
	joinRequestTasks []task
}

func (f *flow) run(pl backend.JoinReqPayload) (backend.JoinAnsPayload, error) {
	ctx := context{
		joinReqPayload: pl,
	}

	for _, t := range f.joinRequestTasks {
		if err := t(&ctx); err != nil {
			return ctx.joinAnsPayload, err
		}
	}

	return ctx.joinAnsPayload, nil
}

var joinFlow = &flow{
	joinRequestTasks: []task{
		setPHYPayload,
		getDevice,
		getApplication,
		getDeviceKeys,
		validateMIC,
		setAppNonce,
		setNetID,
		setSessionKeys,
		createDeviceActivationRecord,
		flushDeviceQueueMapping,
		sendJoinNotification,
		createJoinAnsPayload,
	},
}

// HandleJoinRequest handles a given join-request and returns a join-answer
// payload.
func HandleJoinRequest(pl backend.JoinReqPayload) backend.JoinAnsPayload {
	basePayload := backend.BasePayload{
		ProtocolVersion: backend.ProtocolVersion1_0,
		SenderID:        pl.ReceiverID,
		ReceiverID:      pl.SenderID,
		TransactionID:   pl.TransactionID,
		MessageType:     backend.JoinAns,
	}

	jaPL, err := joinFlow.run(pl)
	if err != nil {
		var resCode backend.ResultCode

		switch errors.Cause(err) {
		case storage.ErrDoesNotExist:
			resCode = backend.UnknownDevEUI
		case ErrInvalidMIC:
			resCode = backend.MICFailed
		default:
			resCode = backend.Other
		}

		jaPL = backend.JoinAnsPayload{
			BasePayload: basePayload,
			Result: backend.Result{
				ResultCode:  resCode,
				Description: err.Error(),
			},
		}
	}

	jaPL.BasePayload = basePayload
	return jaPL
}

func setPHYPayload(ctx *context) error {
	if err := ctx.phyPayload.UnmarshalBinary(ctx.joinReqPayload.PHYPayload[:]); err != nil {
		return errors.Wrap(err, "unmarshal phypayload error")
	}
	return nil
}

func getDevice(ctx *context) error {
	d, err := storage.GetDevice(config.C.PostgreSQL.DB, ctx.joinReqPayload.DevEUI)
	if err != nil {
		return errors.Wrap(err, "get device error")
	}

	ctx.device = d
	return nil
}

func getApplication(ctx *context) error {
	a, err := storage.GetApplication(config.C.PostgreSQL.DB, ctx.device.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get application error")
	}

	ctx.application = a
	return nil
}

func getDeviceKeys(ctx *context) error {
	dk, err := storage.GetDeviceKeys(config.C.PostgreSQL.DB, ctx.device.DevEUI)
	if err != nil {
		return errors.Wrap(err, "get device-keys error")
	}
	ctx.deviceKeys = dk
	return nil
}

func validateMIC(ctx *context) error {
	ok, err := ctx.phyPayload.ValidateMIC(ctx.deviceKeys.AppKey)
	if err != nil {
		return errors.Wrap(err, "validate mic error")
	}
	if !ok {
		return ErrInvalidMIC
	}
	return nil
}

func setAppNonce(ctx *context) error {
	ctx.deviceKeys.JoinNonce++
	if ctx.deviceKeys.JoinNonce > (2<<23)-1 {
		return errors.New("join-nonce overflow")
	}

	if err := storage.UpdateDeviceKeys(config.C.PostgreSQL.DB, &ctx.deviceKeys); err != nil {
		return errors.Wrap(err, "update device-keys error")
	}

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(ctx.deviceKeys.JoinNonce))
	copy(ctx.appNonce[:], b[0:3])

	return nil
}

func setNetID(ctx *context) error {
	if err := ctx.netID.UnmarshalText([]byte(ctx.joinReqPayload.SenderID)); err != nil {
		return errors.Wrap(err, "unmarshal netid error")
	}
	return nil
}

func setSessionKeys(ctx *context) error {
	var err error

	jrPL, ok := ctx.phyPayload.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		return fmt.Errorf("expected *lorawan.JoinRequestPayload, got %T", ctx.phyPayload.MACPayload)
	}

	ctx.nwkSKey, err = getNwkSKey(ctx.deviceKeys.AppKey, ctx.netID, ctx.appNonce, jrPL.DevNonce)
	if err != nil {
		return errors.Wrap(err, "get nwk_s_key error")
	}

	ctx.appSKey, err = getAppSKey(ctx.deviceKeys.AppKey, ctx.netID, ctx.appNonce, jrPL.DevNonce)
	if err != nil {
		return errors.Wrap(err, "get app_s_key error")
	}

	return nil
}

func createDeviceActivationRecord(ctx *context) error {
	da := storage.DeviceActivation{
		DevEUI:  ctx.device.DevEUI,
		DevAddr: ctx.joinReqPayload.DevAddr,
		AppSKey: ctx.appSKey,
		NwkSKey: ctx.nwkSKey,
	}

	if err := storage.CreateDeviceActivation(config.C.PostgreSQL.DB, &da); err != nil {
		return errors.Wrap(err, "create device-activation error")
	}

	return nil
}

func flushDeviceQueueMapping(ctx *context) error {
	if err := storage.FlushDeviceQueueMappingForDevEUI(config.C.PostgreSQL.DB, ctx.device.DevEUI); err != nil {
		return errors.Wrap(err, "flush device-queue mapping error")
	}
	return nil
}

func sendJoinNotification(ctx *context) error {
	err := config.C.ApplicationServer.Integration.Handler.SendJoinNotification(handler.JoinNotification{
		ApplicationID:   ctx.device.ApplicationID,
		ApplicationName: ctx.application.Name,
		DeviceName:      ctx.device.Name,
		DevEUI:          ctx.device.DevEUI,
		DevAddr:         ctx.joinReqPayload.DevAddr,
	})
	if err != nil {
		return errors.Wrap(err, "send join notification error")
	}
	return nil
}

func createJoinAnsPayload(ctx *context) error {
	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			AppNonce:   ctx.appNonce,
			NetID:      ctx.netID,
			DevAddr:    ctx.joinReqPayload.DevAddr,
			DLSettings: ctx.joinReqPayload.DLSettings,
			RXDelay:    uint8(ctx.joinReqPayload.RxDelay),
			CFList:     ctx.joinReqPayload.CFList,
		},
	}

	if err := phy.SetMIC(ctx.deviceKeys.AppKey); err != nil {
		return err
	}

	if err := phy.EncryptJoinAcceptPayload(ctx.deviceKeys.AppKey); err != nil {
		return err
	}

	b, err := phy.MarshalBinary()
	if err != nil {
		return err
	}

	ctx.joinAnsPayload = backend.JoinAnsPayload{
		PHYPayload: backend.HEXBytes(b),
		Result: backend.Result{
			ResultCode: backend.Success,
		},
		NwkSKey: &backend.KeyEnvelope{
			AESKey: ctx.nwkSKey,
		},
		// TODO: add AppSKey and Lifetime
	}

	return nil
}

// getNwkSKey returns the network session key.
func getNwkSKey(appkey lorawan.AES128Key, netID lorawan.NetID, appNonce [3]byte, devNonce [2]byte) (lorawan.AES128Key, error) {
	return getSKey(0x01, appkey, netID, appNonce, devNonce)
}

// getAppSKey returns the application session key.
func getAppSKey(appkey lorawan.AES128Key, netID lorawan.NetID, appNonce [3]byte, devNonce [2]byte) (lorawan.AES128Key, error) {
	return getSKey(0x02, appkey, netID, appNonce, devNonce)
}

func getSKey(typ byte, appkey lorawan.AES128Key, netID lorawan.NetID, appNonce [3]byte, devNonce [2]byte) (lorawan.AES128Key, error) {
	var key lorawan.AES128Key
	b := make([]byte, 0, 16)
	b = append(b, typ)

	// little endian
	for i := len(appNonce) - 1; i >= 0; i-- {
		b = append(b, appNonce[i])
	}
	for i := len(netID) - 1; i >= 0; i-- {
		b = append(b, netID[i])
	}
	for i := len(devNonce) - 1; i >= 0; i-- {
		b = append(b, devNonce[i])
	}
	pad := make([]byte, 7)
	b = append(b, pad...)

	block, err := aes.NewCipher(appkey[:])
	if err != nil {
		return key, err
	}
	if block.BlockSize() != len(b) {
		return key, fmt.Errorf("block-size of %d bytes is expected", len(b))
	}
	block.Encrypt(key[:], b)
	return key, nil
}
