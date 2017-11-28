package api

import (
	"crypto/aes"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/brocaar/lora-app-server/internal/gwping"

	log "github.com/sirupsen/logrus"
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
}

// NewApplicationServerAPI returns a new ApplicationServerAPI.
func NewApplicationServerAPI() *ApplicationServerAPI {
	return &ApplicationServerAPI{}
}

// HandleDataUp handles incoming (uplink) data.
func (a *ApplicationServerAPI) HandleDataUp(ctx context.Context, req *as.HandleDataUpRequest) (*as.HandleDataUpResponse, error) {
	var appEUI, devEUI lorawan.EUI64
	copy(appEUI[:], req.AppEUI)
	copy(devEUI[:], req.DevEUI)

	d, err := storage.GetDevice(common.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get device error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	app, err := storage.GetApplication(common.DB, d.ApplicationID)
	if err != nil {
		errStr := fmt.Sprintf("get application error: %s", err)
		log.WithField("id", d.ApplicationID).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	da, err := storage.GetLastDeviceActivationForDevEUI(common.DB, d.DevEUI)
	if err != nil {
		errStr := fmt.Sprintf("get device-activation error: %s", err)
		log.WithField("dev_eui", d.DevEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	b, err := lorawan.EncryptFRMPayload(da.AppSKey, true, da.DevAddr, req.FCnt, req.Data)
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
		NodeName:        d.Name,
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
			MAC:       mac,
			Time:      timestamp,
			RSSI:      int(rxInfo.Rssi),
			LoRaSNR:   rxInfo.LoRaSNR,
			Name:      rxInfo.Name,
			Latitude:  rxInfo.Latitude,
			Longitude: rxInfo.Longitude,
			Altitude:  rxInfo.Altitude,
		})
	}

	err = common.Handler.SendDataUp(pl)
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

	qi, err := storage.GetNextDeviceQueueItem(common.DB, devEUI, int(req.MaxPayloadSize))
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

	d, err := storage.GetDevice(common.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get device error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	da, err := storage.GetLastDeviceActivationForDevEUI(common.DB, d.DevEUI)
	if err != nil {
		errStr := fmt.Sprintf("get device-activation error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, req.FCnt, qi.Data)
	if err != nil {
		errStr := fmt.Sprintf("encrypt payload error: %s", err)
		log.WithFields(log.Fields{
			"dev_eui": devEUI,
			"id":      qi.ID,
		}).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	queueSize, err := storage.GetDeviceQueueItemCount(common.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get downlink queue size error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	if !qi.Confirmed {
		if err := storage.DeleteDeviceQueueItem(common.DB, qi.ID); err != nil {
			errStr := fmt.Sprintf("delete downlink queue item error: %s", err)
			log.WithFields(log.Fields{
				"dev_eui": devEUI,
				"id":      qi.ID,
			}).Error(errStr)
			return nil, grpc.Errorf(codes.Internal, errStr)
		}
	} else {
		qi.Pending = true
		if err := storage.UpdateDeviceQueueItem(common.DB, qi); err != nil {
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

	d, err := storage.GetDevice(common.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get device error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}
	app, err := storage.GetApplication(common.DB, d.ApplicationID)
	if err != nil {
		errStr := fmt.Sprintf("get application error: %s", err)
		log.WithField("id", d.ApplicationID).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	qi, err := storage.GetPendingDeviceQueueItem(common.DB, devEUI)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	if err := storage.DeleteDeviceQueueItem(common.DB, qi.ID); err != nil {
		return nil, grpc.Errorf(codes.Unknown, err.Error())
	}
	log.WithFields(log.Fields{
		"dev_eui": qi.DevEUI,
	}).Info("downlink queue item acknowledged")

	err = common.Handler.SendACKNotification(handler.ACKNotification{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		NodeName:        d.Name,
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

	d, err := storage.GetDevice(common.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get device error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}
	app, err := storage.GetApplication(common.DB, d.ApplicationID)
	if err != nil {
		errStr := fmt.Sprintf("get application error: %s", err)
		log.WithField("id", d.ApplicationID).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	log.WithFields(log.Fields{
		"type":    req.Type,
		"dev_eui": devEUI,
	}).Error(req.Error)

	err = common.Handler.SendErrorNotification(handler.ErrorNotification{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		NodeName:        d.Name,
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

// HandleProprietaryUp handles proprietary uplink payloads.
func (a *ApplicationServerAPI) HandleProprietaryUp(ctx context.Context, req *as.HandleProprietaryUpRequest) (*as.HandleProprietaryUpResponse, error) {
	err := gwping.HandleReceivedPing(req)
	if err != nil {
		errStr := fmt.Sprintf("handle received ping error: %s", err)
		log.Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	return &as.HandleProprietaryUpResponse{}, nil
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
