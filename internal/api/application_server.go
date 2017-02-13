package api

import (
	"crypto/aes"
	"crypto/rand"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/handler"
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
		log.Errorf("unmarshal join-request PHYPayload error: %s", err)
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	jrPL, ok := phy.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		log.Errorf("join-request PHYPayload does not contain a JoinRequestPayload")
		return nil, grpc.Errorf(codes.InvalidArgument, "PHYPayload does not contain a JoinRequestPayload")
	}

	var netID lorawan.NetID
	var devAddr lorawan.DevAddr

	copy(netID[:], req.NetID)
	copy(devAddr[:], req.DevAddr)

	// get the node and application from the db and validate the AppEUI
	node, err := storage.GetNode(a.ctx.DB, jrPL.DevEUI)
	if err != nil {
		log.WithFields(log.Fields{
			"dev_eui": jrPL.DevEUI,
		}).Errorf("join-request node does not exist")
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	app, err := storage.GetApplication(a.ctx.DB, node.ApplicationID)
	if err != nil {
		log.WithFields(log.Fields{
			"id": node.ApplicationID,
		}).Errorf("get application error: %s", err)
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	if node.AppEUI != jrPL.AppEUI {
		log.WithFields(log.Fields{
			"dev_eui":          node.DevEUI,
			"expected_app_eui": node.AppEUI,
			"request_app_eui":  jrPL.AppEUI,
		}).Error("join-request DevEUI exists, but with a different AppEUI")
		return nil, grpc.Errorf(codes.Unknown, "DevEUI exists, but with a different AppEUI")
	}

	// validate MIC
	ok, err = phy.ValidateMIC(node.AppKey)
	if err != nil {
		log.WithFields(log.Fields{
			"dev_eui": node.DevEUI,
			"app_eui": node.AppEUI,
		}).Errorf("join-request validate mic error: %s", err)
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	if !ok {
		log.WithFields(log.Fields{
			"dev_eui": node.DevEUI,
			"app_eui": node.AppEUI,
			"mic":     phy.MIC,
		}).Error("join-request invalid mic")
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid MIC")
	}

	// validate that the DevNonce hasn't been used before
	if !node.ValidateDevNonce(jrPL.DevNonce) {
		log.WithFields(log.Fields{
			"dev_eui":   node.DevEUI,
			"app_eui":   node.AppEUI,
			"dev_nonce": jrPL.DevNonce,
		}).Error("join-request DevNonce has already been used")
		return nil, grpc.Errorf(codes.InvalidArgument, "DevNonce has already been used")
	}

	// now we know the frame is valid, test if the node is allowed to OTAA
	if node.IsABP {
		return nil, grpc.Errorf(codes.FailedPrecondition, "node is ABP device")
	}

	// get app nonce
	appNonce, err := getAppNonce()
	if err != nil {
		log.Errorf("get AppNone error: %s", err)
		return nil, grpc.Errorf(codes.Unknown, "get AppNonce error: %s", err)
	}

	// get the (optional) CFList
	cFList, err := storage.GetCFListForNode(a.ctx.DB, node)
	if err != nil {
		log.WithFields(log.Fields{
			"dev_eui": node.DevEUI,
			"app_eui": node.AppEUI,
		}).Errorf("join-request get CFList error: %s", err)
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
		PhyPayload:         b,
		NwkSKey:            nwkSKey[:],
		RxDelay:            uint32(node.RXDelay),
		Rx1DROffset:        uint32(node.RX1DROffset),
		RxWindow:           as.RXWindow(node.RXWindow),
		Rx2DR:              uint32(node.RX2DR),
		RelaxFCnt:          node.RelaxFCnt,
		AdrInterval:        node.ADRInterval,
		InstallationMargin: node.InstallationMargin,
	}

	if cFList != nil {
		resp.CFList = cFList[:]
	}

	log.WithFields(log.Fields{
		"dev_eui":          node.DevEUI,
		"app_eui":          node.AppEUI,
		"dev_addr":         node.DevAddr,
		"application_name": app.Name,
		"node_name":        node.Name,
	}).Info("join-request accepted")

	err = a.ctx.Handler.SendJoinNotification(handler.JoinNotification{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		NodeName:        node.Name,
		DevAddr:         node.DevAddr,
		DevEUI:          node.DevEUI,
	})
	if err != nil {
		log.Errorf("send join notification to handler error: %s", err)
	}

	return &resp, nil
}

// HandleDataUp handles incoming (uplink) data.
func (a *ApplicationServerAPI) HandleDataUp(ctx context.Context, req *as.HandleDataUpRequest) (*as.HandleDataUpResponse, error) {
	if len(req.RxInfo) == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "RxInfo must have length > 0")
	}

	var appEUI, devEUI lorawan.EUI64
	copy(appEUI[:], req.AppEUI)
	copy(devEUI[:], req.DevEUI)

	node, err := storage.GetNode(a.ctx.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get node error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}
	app, err := storage.GetApplication(a.ctx.DB, node.ApplicationID)
	if err != nil {
		errStr := fmt.Sprintf("get application error: %s", err)
		log.WithField("id", node.ApplicationID).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	b, err := lorawan.EncryptFRMPayload(node.AppSKey, true, node.DevAddr, req.FCnt, req.Data)
	if err != nil {
		log.WithFields(log.Fields{
			"dev_eui": devEUI,
			"f_cnt":   req.FCnt,
		}).Errorf("decrypt payload error: %s", err)
		return nil, grpc.Errorf(codes.Internal, "decrypt payload error: %s", err)
	}

	pl := handler.DataUpPayload{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		NodeName:        node.Name,
		DevEUI:          devEUI,
		RXInfo:          []handler.RXInfo{},
		TXInfo: handler.TXInfo{
			Frequency: int(req.TxInfo.Frequency),
			DataRate: handler.DataRate{
				Modulation:   req.TxInfo.DataRate.Modulation,
				Bandwidth:    int(req.TxInfo.DataRate.BandWidth),
				SpreadFactor: int(req.TxInfo.DataRate.SpreadFactor),
				Bitrate:      int(req.TxInfo.DataRate.Bitrate),
			},
			ADR:      req.TxInfo.Adr,
			CodeRate: req.TxInfo.CodeRate,
		},
		FCnt:  req.FCnt,
		FPort: uint8(req.FPort),
		Data:  b,
	}

	for _, rxInfo := range req.RxInfo {
		var timestamp *time.Time
		var mac lorawan.EUI64
		copy(mac[:], rxInfo.Mac)

		if len(rxInfo.Time) > 0 {
			ts, err := time.Parse(time.RFC3339Nano, rxInfo.Time)
			if err != nil {
				log.WithFields(log.Fields{
					"dev_eui":  devEUI,
					"time_str": rxInfo.Time,
				}).Errorf("unmarshal time error: %s", err)
			} else if !ts.Equal(time.Time{}) {
				timestamp = &ts
			}
		}
		pl.RXInfo = append(pl.RXInfo, handler.RXInfo{
			MAC:     mac,
			Time:    timestamp,
			RSSI:    int(rxInfo.Rssi),
			LoRaSNR: rxInfo.LoRaSNR,
		})
	}

	err = a.ctx.Handler.SendDataUp(pl)
	if err != nil {
		errStr := fmt.Sprintf("send data up to handler error: %s", err)
		log.Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	return &as.HandleDataUpResponse{}, nil
}

// GetDataDown returns the first payload from the datadown queue.
func (a *ApplicationServerAPI) GetDataDown(ctx context.Context, req *as.GetDataDownRequest) (*as.GetDataDownResponse, error) {
	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEUI)

	qi, err := storage.GetNextDownlinkQueueItem(a.ctx.DB, devEUI, int(req.MaxPayloadSize))
	if err != nil {
		errStr := fmt.Sprintf("get next downlink queue item error: %s", err)
		log.WithFields(log.Fields{
			"dev_eui":          devEUI,
			"max_payload_size": req.MaxPayloadSize,
		}).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	// the queue is empty
	if qi == nil {
		log.WithField("dev_eui", devEUI).Info("data-down item requested by network-server, but queue is empty")
		return &as.GetDataDownResponse{}, nil
	}

	node, err := storage.GetNode(a.ctx.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get node error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	b, err := lorawan.EncryptFRMPayload(node.AppSKey, false, node.DevAddr, req.FCnt, qi.Data)
	if err != nil {
		errStr := fmt.Sprintf("encrypt payload error: %s", err)
		log.WithFields(log.Fields{
			"dev_eui": devEUI,
			"id":      qi.ID,
		}).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	queueSize, err := storage.GetDownlinkQueueSize(a.ctx.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get downlink queue size error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	if !qi.Confirmed {
		if err := storage.DeleteDownlinkQueueItem(a.ctx.DB, qi.ID); err != nil {
			errStr := fmt.Sprintf("delete downlink queue item error: %s", err)
			log.WithFields(log.Fields{
				"dev_eui": devEUI,
				"id":      qi.ID,
			}).Error(errStr)
			return nil, grpc.Errorf(codes.Internal, errStr)
		}
	} else {
		qi.Pending = true
		if err := storage.UpdateDownlinkQueueItem(a.ctx.DB, *qi); err != nil {
			errStr := fmt.Sprintf("update downlink queue item error: %s", err)
			log.WithFields(log.Fields{
				"dev_eui": devEUI,
				"id":      qi.ID,
			}).Error(errStr)
			return nil, grpc.Errorf(codes.Internal, errStr)
		}
	}

	log.WithFields(log.Fields{
		"dev_eui":   devEUI,
		"confirmed": qi.Confirmed,
		"id":        qi.ID,
		"fcnt":      req.FCnt,
	}).Info("data-down item requested by network-server")

	return &as.GetDataDownResponse{
		Data:      b,
		Confirmed: qi.Confirmed,
		FPort:     uint32(qi.FPort),
		MoreData:  queueSize > 1,
	}, nil

}

// HandleDataDownACK handles an ack on a downlink transmission.
func (a *ApplicationServerAPI) HandleDataDownACK(ctx context.Context, req *as.HandleDataDownACKRequest) (*as.HandleDataDownACKResponse, error) {
	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEUI)

	node, err := storage.GetNode(a.ctx.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get node error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}
	app, err := storage.GetApplication(a.ctx.DB, node.ApplicationID)
	if err != nil {
		errStr := fmt.Sprintf("get application error: %s", err)
		log.WithField("id", node.ApplicationID).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	qi, err := storage.GetPendingDownlinkQueueItem(a.ctx.DB, devEUI)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	if err := storage.DeleteDownlinkQueueItem(a.ctx.DB, qi.ID); err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	log.WithFields(log.Fields{
		"application_name": app.Name,
		"node_name":        node.Name,
		"dev_eui":          qi.DevEUI,
	}).Info("downlink queue item acknowledged")

	err = a.ctx.Handler.SendACKNotification(handler.ACKNotification{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		NodeName:        node.Name,
		DevEUI:          devEUI,
		Reference:       qi.Reference,
	})
	if err != nil {
		log.Errorf("send ack notification to handler error: %s", err)
	}

	return &as.HandleDataDownACKResponse{}, nil
}

// HandleError handles an incoming error.
func (a *ApplicationServerAPI) HandleError(ctx context.Context, req *as.HandleErrorRequest) (*as.HandleErrorResponse, error) {
	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEUI)

	node, err := storage.GetNode(a.ctx.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get node error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}
	app, err := storage.GetApplication(a.ctx.DB, node.ApplicationID)
	if err != nil {
		errStr := fmt.Sprintf("get application error: %s", err)
		log.WithField("id", node.ApplicationID).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	log.WithFields(log.Fields{
		"application_name": app.Name,
		"node_name":        node.Name,
		"type":             req.Type,
		"dev_eui":          devEUI,
	}).Error(req.Error)

	err = a.ctx.Handler.SendErrorNotification(handler.ErrorNotification{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		NodeName:        node.Name,
		DevEUI:          devEUI,
		Type:            req.Type.String(),
		Error:           req.Error,
	})
	if err != nil {
		errStr := fmt.Sprintf("send error notification to handler error: %s", err)
		log.Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	return &as.HandleErrorResponse{}, nil
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
