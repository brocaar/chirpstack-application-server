package api

import (
	"context"
	"crypto/aes"
	"crypto/rand"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/as"
	"github.com/brocaar/lorawan"
)

// ApplicationServerAPI implements the as.ApplicationServerServer interface.
type ApplicationServerAPI struct {
	ctx common.Context
}

// NewApplicationServerAPI returns a new ApplicationServerAPI.
func NewApplicationServerAPI(ctx common.Context) *ApplicationServerAPI {
	return &ApplicationServerAPI{
		ctx: ctx,
	}
}

// JoinRequest handles a join-request.
func (a *ApplicationServerAPI) JoinRequest(ctx context.Context, req *as.JoinRequestRequest) (*as.JoinRequestResponse, error) {
	var phy lorawan.PHYPayload

	if err := phy.UnmarshalBinary(req.PhyPayload); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	jrPL, ok := phy.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		return nil, grpc.Errorf(codes.InvalidArgument, "PHYPayload does not contain a JoinRequestPayload")
	}

	var netID lorawan.NetID
	var devAddr lorawan.DevAddr

	copy(netID[:], req.NetID)
	copy(devAddr[:], req.DevAddr)

	// get the node from the db and validate the AppEUI
	node, err := storage.GetNode(a.ctx.DB, jrPL.DevEUI)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	if node.AppEUI != jrPL.AppEUI {
		return nil, grpc.Errorf(codes.Unknown, "DevEUI exists, but with a different AppEUI")
	}

	// validate MIC
	ok, err = phy.ValidateMIC(node.AppKey)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	if !ok {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid MIC")
	}

	// validate that the DevNonce hasn't been used before
	if !node.ValidateDevNonce(jrPL.DevNonce) {
		return nil, grpc.Errorf(codes.InvalidArgument, "DevNonce has already been used")
	}

	// get app nonce
	appNonce, err := getAppNonce()
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "get AppNonce error: %s", err)
	}

	// get the (optional) CFList
	cFList, err := storage.GetCFListForNode(a.ctx.DB, node)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	// get keys
	nwkSKey, err := getNwkSKey(node.AppKey, netID, appNonce, jrPL.DevNonce)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	appSKey, err := getAppSKey(node.AppKey, netID, appNonce, jrPL.DevNonce)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	// update the node
	node.DevAddr = devAddr
	node.NwkSKey = nwkSKey
	node.AppSKey = appSKey
	if err = storage.UpdateNode(a.ctx.DB, node); err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	// construct response
	jaPHY := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			AppNonce: appNonce,
			NetID:    netID,
			DevAddr:  devAddr,
			RXDelay:  node.RXDelay,
			DLSettings: lorawan.DLSettings{
				RX2DataRate: uint8(node.RX2DR),
				RX1DROffset: node.RX1DROffset,
			},
			CFList: cFList,
		},
	}
	if err = jaPHY.SetMIC(node.AppKey); err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	if err = jaPHY.EncryptJoinAcceptPayload(node.AppKey); err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	b, err := jaPHY.MarshalBinary()
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}

	resp := as.JoinRequestResponse{
		PhyPayload:  b,
		NwkSKey:     nwkSKey[:],
		RxDelay:     uint32(node.RXDelay),
		Rx1DROffset: uint32(node.RX1DROffset),
		RxWindow:    as.RXWindow(node.RXWindow),
		Rx2DR:       uint32(node.RX2DR),
	}

	if cFList != nil {
		resp.CFList = cFList[:]
	}

	return &resp, nil
}

// HandleDataUp handles incoming (uplink) data.
func (a *ApplicationServerAPI) HandleDataUp(ctx context.Context, req *as.HandleDataUpRequest) (*as.HandleDataUpResponse, error) {
	panic("not implemented")
}

// GetDataDown returns the first payload from the datadown queue.
func (a *ApplicationServerAPI) GetDataDown(ctx context.Context, req *as.GetDataDownRequest) (*as.GetDataDownResponse, error) {
	panic("not implemented")
}

// HandleDataDownACK handles an ack on a downlink transmission.
func (a *ApplicationServerAPI) HandleDataDownACK(ctx context.Context, req *as.HandleDataDownACKRequest) (*as.HandleDataDownACKResponse, error) {
	panic("not implemented")
}

// HandleError handles an incoming error.
func (a *ApplicationServerAPI) HandleError(ctx context.Context, req *as.HandleErrorRequest) (*as.HandleErrorResponse, error) {
	panic("not implemented")
}

// getAppNonce returns a random application nonce (used for OTAA).
func getAppNonce() ([3]byte, error) {
	var b [3]byte
	if _, err := rand.Read(b[:]); err != nil {
		return b, err
	}
	return b, nil
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
