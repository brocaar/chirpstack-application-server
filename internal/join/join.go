package join

import (
	"crypto/aes"
	"fmt"

	"github.com/pkg/errors"

	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

type context struct {
	joinReqPayload   backend.JoinReqPayload
	joinAnsPayload   backend.JoinAnsPayload
	rejoinReqPayload backend.RejoinReqPayload
	rejoinAnsPaylaod backend.RejoinAnsPayload
	joinType         lorawan.JoinType
	phyPayload       lorawan.PHYPayload
	application      storage.Application
	deviceKeys       storage.DeviceKeys
	devNonce         lorawan.DevNonce
	joinNonce        lorawan.JoinNonce
	netID            lorawan.NetID
	devEUI           lorawan.EUI64
	joinEUI          lorawan.EUI64
	fNwkSIntKey      lorawan.AES128Key
	appSKey          lorawan.AES128Key
	sNwkSIntKey      lorawan.AES128Key
	nwkSEncKey       lorawan.AES128Key
}

var joinTasks = []func(*context) error{
	setJoinContext,
	getDeviceKeys,
	validateMIC,
	setJoinNonce,
	setSessionKeys,
	createJoinAnsPayload,
}

var rejoinTasks = []func(*context) error{
	setRejoinContext,
	getDeviceKeys,
	setJoinNonce,
	setSessionKeys,
	createRejoinAnsPayload,
}

func handleJoinRequest(pl backend.JoinReqPayload) (backend.JoinAnsPayload, error) {
	ctx := context{
		joinReqPayload: pl,
	}

	for _, t := range joinTasks {
		if err := t(&ctx); err != nil {
			return ctx.joinAnsPayload, err
		}
	}

	return ctx.joinAnsPayload, nil
}

func handleRejoinRequest(pl backend.RejoinReqPayload) (backend.RejoinAnsPayload, error) {
	ctx := context{
		rejoinReqPayload: pl,
	}

	for _, t := range rejoinTasks {
		if err := t(&ctx); err != nil {
			return ctx.rejoinAnsPaylaod, err
		}
	}

	return ctx.rejoinAnsPaylaod, nil
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

	jaPL, err := handleJoinRequest(pl)
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

// HandleRejoinRequest handles a given rejoin-request and returns a
// rejoin-answer payload.
func HandleRejoinRequest(pl backend.RejoinReqPayload) backend.RejoinAnsPayload {
	basePayload := backend.BasePayload{
		ProtocolVersion: backend.ProtocolVersion1_0,
		SenderID:        pl.ReceiverID,
		ReceiverID:      pl.SenderID,
		TransactionID:   pl.TransactionID,
		MessageType:     backend.RejoinAns,
	}

	rjaPL, err := handleRejoinRequest(pl)
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

		rjaPL = backend.RejoinAnsPayload{
			BasePayload: basePayload,
			Result: backend.Result{
				ResultCode:  resCode,
				Description: err.Error(),
			},
		}
	}

	rjaPL.BasePayload = basePayload
	return rjaPL
}

func setJoinContext(ctx *context) error {
	if err := ctx.phyPayload.UnmarshalBinary(ctx.joinReqPayload.PHYPayload[:]); err != nil {
		return errors.Wrap(err, "unmarshal phypayload error")
	}

	if err := ctx.netID.UnmarshalText([]byte(ctx.joinReqPayload.SenderID)); err != nil {
		return errors.Wrap(err, "unmarshal netid error")
	}

	if err := ctx.joinEUI.UnmarshalText([]byte(ctx.joinReqPayload.ReceiverID)); err != nil {
		return errors.Wrap(err, "unmarshal joineui error")
	}

	ctx.devEUI = ctx.joinReqPayload.DevEUI
	ctx.joinType = lorawan.JoinRequestType

	switch v := ctx.phyPayload.MACPayload.(type) {
	case *lorawan.JoinRequestPayload:
		ctx.devNonce = v.DevNonce
	default:
		return fmt.Errorf("expected *lorawan.JoinRequestPayload, got %T", ctx.phyPayload.MACPayload)
	}

	return nil
}

func setRejoinContext(ctx *context) error {
	if err := ctx.phyPayload.UnmarshalBinary(ctx.rejoinReqPayload.PHYPayload[:]); err != nil {
		return errors.Wrap(err, "unmarshal phypayload error")
	}

	if err := ctx.netID.UnmarshalText([]byte(ctx.rejoinReqPayload.SenderID)); err != nil {
		return errors.Wrap(err, "unmarshal netid error")
	}

	if err := ctx.joinEUI.UnmarshalText([]byte(ctx.rejoinReqPayload.ReceiverID)); err != nil {
		return errors.Wrap(err, "unmarshal joineui error")
	}

	switch v := ctx.phyPayload.MACPayload.(type) {
	case *lorawan.RejoinRequestType02Payload:
		ctx.joinType = v.RejoinType
		ctx.devNonce = lorawan.DevNonce(v.RJCount0)
	case *lorawan.RejoinRequestType1Payload:
		ctx.joinType = v.RejoinType
		ctx.devNonce = lorawan.DevNonce(v.RJCount1)
	default:
		return fmt.Errorf("expected rejoin payload, got %T", ctx.phyPayload.MACPayload)
	}

	ctx.devEUI = ctx.rejoinReqPayload.DevEUI

	return nil
}

func getDeviceKeys(ctx *context) error {
	dk, err := storage.GetDeviceKeys(storage.DB(), ctx.devEUI)
	if err != nil {
		return errors.Wrap(err, "get device-keys error")
	}
	ctx.deviceKeys = dk
	return nil
}

func validateMIC(ctx *context) error {
	ok, err := ctx.phyPayload.ValidateUplinkJoinMIC(ctx.deviceKeys.NwkKey)
	if err != nil {
		return errors.Wrap(err, "validate mic error")
	}
	if !ok {
		return ErrInvalidMIC
	}
	return nil
}

func setJoinNonce(ctx *context) error {
	ctx.deviceKeys.JoinNonce++
	if ctx.deviceKeys.JoinNonce > (1<<24)-1 {
		return errors.New("join-nonce overflow")
	}
	ctx.joinNonce = lorawan.JoinNonce(ctx.deviceKeys.JoinNonce)

	if err := storage.UpdateDeviceKeys(storage.DB(), &ctx.deviceKeys); err != nil {
		return errors.Wrap(err, "update device-keys error")
	}

	return nil
}

func setSessionKeys(ctx *context) error {
	var err error

	ctx.fNwkSIntKey, err = getFNwkSIntKey(ctx.joinReqPayload.DLSettings.OptNeg, ctx.deviceKeys.NwkKey, ctx.netID, ctx.joinEUI, ctx.joinNonce, ctx.devNonce)
	if err != nil {
		return errors.Wrap(err, "get FNwkSIntKey error")
	}

	if ctx.joinReqPayload.DLSettings.OptNeg {
		ctx.appSKey, err = getAppSKey(ctx.joinReqPayload.DLSettings.OptNeg, ctx.deviceKeys.AppKey, ctx.netID, ctx.joinEUI, ctx.joinNonce, ctx.devNonce)
		if err != nil {
			return errors.Wrap(err, "get AppSKey error")
		}
	} else {
		ctx.appSKey, err = getAppSKey(ctx.joinReqPayload.DLSettings.OptNeg, ctx.deviceKeys.NwkKey, ctx.netID, ctx.joinEUI, ctx.joinNonce, ctx.devNonce)
		if err != nil {
			return errors.Wrap(err, "get AppSKey error")
		}
	}

	ctx.sNwkSIntKey, err = getSNwkSIntKey(ctx.joinReqPayload.DLSettings.OptNeg, ctx.deviceKeys.NwkKey, ctx.netID, ctx.joinEUI, ctx.joinNonce, ctx.devNonce)
	if err != nil {
		return errors.Wrap(err, "get SNwkSIntKey error")
	}

	ctx.nwkSEncKey, err = getNwkSEncKey(ctx.joinReqPayload.DLSettings.OptNeg, ctx.deviceKeys.NwkKey, ctx.netID, ctx.joinEUI, ctx.joinNonce, ctx.devNonce)
	if err != nil {
		return errors.Wrap(err, "get NwkSEncKey error")
	}

	return nil
}

func createJoinAnsPayload(ctx *context) error {
	var cFList *lorawan.CFList
	if len(ctx.joinReqPayload.CFList[:]) != 0 {
		cFList = new(lorawan.CFList)
		if err := cFList.UnmarshalBinary(ctx.joinReqPayload.CFList[:]); err != nil {
			return errors.Wrap(err, "unmarshal cflist error")
		}
	}

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			JoinNonce:  ctx.joinNonce,
			HomeNetID:  ctx.netID,
			DevAddr:    ctx.joinReqPayload.DevAddr,
			DLSettings: ctx.joinReqPayload.DLSettings,
			RXDelay:    uint8(ctx.joinReqPayload.RxDelay),
			CFList:     cFList,
		},
	}

	if ctx.joinReqPayload.DLSettings.OptNeg {
		jsIntKey, err := getJSIntKey(ctx.deviceKeys.NwkKey, ctx.devEUI)
		if err != nil {
			return err
		}
		if err := phy.SetDownlinkJoinMIC(ctx.joinType, ctx.joinEUI, ctx.devNonce, jsIntKey); err != nil {
			return err
		}
	} else {
		if err := phy.SetDownlinkJoinMIC(ctx.joinType, ctx.joinEUI, ctx.devNonce, ctx.deviceKeys.NwkKey); err != nil {
			return err
		}
	}

	if err := phy.EncryptJoinAcceptPayload(ctx.deviceKeys.NwkKey); err != nil {
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
		// TODO add Lifetime?
	}

	ctx.joinAnsPayload.AppSKey, err = getASKeyEnvelope(ctx.appSKey)
	if err != nil {
		return err
	}

	if ctx.joinReqPayload.DLSettings.OptNeg {
		// LoRaWAN 1.1+
		ctx.joinAnsPayload.FNwkSIntKey, err = getNSKeyEnvelope(ctx.netID, ctx.fNwkSIntKey)
		if err != nil {
			return err
		}
		ctx.joinAnsPayload.SNwkSIntKey, err = getNSKeyEnvelope(ctx.netID, ctx.sNwkSIntKey)
		if err != nil {
			return err
		}
		ctx.joinAnsPayload.NwkSEncKey, err = getNSKeyEnvelope(ctx.netID, ctx.nwkSEncKey)
		if err != nil {
			return err
		}
	} else {
		// LoRaWAN 1.0.x
		ctx.joinAnsPayload.NwkSKey, err = getNSKeyEnvelope(ctx.netID, ctx.fNwkSIntKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func createRejoinAnsPayload(ctx *context) error {
	var cFList *lorawan.CFList
	if len(ctx.rejoinReqPayload.CFList[:]) != 0 {
		cFList = new(lorawan.CFList)
		if err := cFList.UnmarshalBinary(ctx.rejoinReqPayload.CFList[:]); err != nil {
			return errors.Wrap(err, "unmarshal cflist error")
		}
	}

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			JoinNonce:  ctx.joinNonce,
			HomeNetID:  ctx.netID,
			DevAddr:    ctx.rejoinReqPayload.DevAddr,
			DLSettings: ctx.rejoinReqPayload.DLSettings,
			RXDelay:    uint8(ctx.rejoinReqPayload.RxDelay),
			CFList:     cFList,
		},
	}

	jsIntKey, err := getJSIntKey(ctx.deviceKeys.NwkKey, ctx.devEUI)
	if err != nil {
		return err
	}

	jsEncKey, err := getJSEncKey(ctx.deviceKeys.NwkKey, ctx.devEUI)
	if err != nil {
		return err
	}

	if err := phy.SetDownlinkJoinMIC(ctx.joinType, ctx.joinEUI, ctx.devNonce, jsIntKey); err != nil {
		return err
	}

	if err := phy.EncryptJoinAcceptPayload(jsEncKey); err != nil {
		return err
	}

	b, err := phy.MarshalBinary()
	if err != nil {
		return err
	}

	// as the rejoin-request is only implemented for LoRaWAN1.1+ there is no
	// need to check the OptNeg flag
	ctx.rejoinAnsPaylaod = backend.RejoinAnsPayload{
		Result: backend.Result{
			ResultCode: backend.Success,
		},
		PHYPayload: backend.HEXBytes(b),
		// TODO: add Lifetime?
	}

	ctx.rejoinAnsPaylaod.AppSKey, err = getASKeyEnvelope(ctx.appSKey)
	if err != nil {
		return err
	}

	ctx.rejoinAnsPaylaod.FNwkSIntKey, err = getNSKeyEnvelope(ctx.netID, ctx.fNwkSIntKey)
	if err != nil {
		return err
	}
	ctx.rejoinAnsPaylaod.SNwkSIntKey, err = getNSKeyEnvelope(ctx.netID, ctx.sNwkSIntKey)
	if err != nil {
		return err
	}
	ctx.rejoinAnsPaylaod.NwkSEncKey, err = getNSKeyEnvelope(ctx.netID, ctx.nwkSEncKey)
	if err != nil {
		return err
	}

	return nil
}

// getFNwkSIntKey returns the FNwkSIntKey.
// For LoRaWAN 1.0: SNwkSIntKey = NwkSEncKey = FNwkSIntKey = NwkSKey
func getFNwkSIntKey(optNeg bool, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	return getSKey(optNeg, 0x01, nwkKey, netID, joinEUI, joinNonce, devNonce)
}

// getAppSKey returns appSKey.
func getAppSKey(optNeg bool, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	return getSKey(optNeg, 0x02, nwkKey, netID, joinEUI, joinNonce, devNonce)
}

// getSNwkSIntKey returns the NwkSIntKey.
func getSNwkSIntKey(optNeg bool, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	return getSKey(optNeg, 0x03, nwkKey, netID, joinEUI, joinNonce, devNonce)
}

// getNwkSEncKey returns the NwkSEncKey.
func getNwkSEncKey(optNeg bool, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	return getSKey(optNeg, 0x04, nwkKey, netID, joinEUI, joinNonce, devNonce)
}

// getJSIntKey returns the JSIntKey.
func getJSIntKey(nwkKey lorawan.AES128Key, devEUI lorawan.EUI64) (lorawan.AES128Key, error) {
	return getJSKey(0x06, devEUI, nwkKey)
}

// getJSEncKey returns the JSEncKey.
func getJSEncKey(nwkKey lorawan.AES128Key, devEUI lorawan.EUI64) (lorawan.AES128Key, error) {
	return getJSKey(0x05, devEUI, nwkKey)
}

func getSKey(optNeg bool, typ byte, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	var key lorawan.AES128Key
	b := make([]byte, 16)
	b[0] = typ

	netIDB, err := netID.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	joinEUIB, err := joinEUI.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	joinNonceB, err := joinNonce.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	devNonceB, err := devNonce.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	if optNeg {
		copy(b[1:4], joinNonceB)
		copy(b[4:12], joinEUIB)
		copy(b[12:14], devNonceB)
	} else {
		copy(b[1:4], joinNonceB)
		copy(b[4:7], netIDB)
		copy(b[7:9], devNonceB)
	}

	block, err := aes.NewCipher(nwkKey[:])
	if err != nil {
		return key, err
	}
	if block.BlockSize() != len(b) {
		return key, fmt.Errorf("block-size of %d bytes is expected", len(b))
	}
	block.Encrypt(key[:], b)

	return key, nil
}

func getJSKey(typ byte, devEUI lorawan.EUI64, nwkKey lorawan.AES128Key) (lorawan.AES128Key, error) {
	var key lorawan.AES128Key
	b := make([]byte, 16)

	b[0] = typ

	devB, err := devEUI.MarshalBinary()
	if err != nil {
		return key, err
	}
	copy(b[1:9], devB[:])

	block, err := aes.NewCipher(nwkKey[:])
	if err != nil {
		return key, err
	}
	if block.BlockSize() != len(b) {
		return key, fmt.Errorf("block-size of %d bytes is expected", len(b))
	}
	block.Encrypt(key[:], b)
	return key, nil
}
