package api

import (
	"crypto/aes"
	"crypto/rand"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/gusseleet/lora-app-server/internal/codec"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/gwping"
	"github.com/gusseleet/lora-app-server/internal/handler"
	"github.com/gusseleet/lora-app-server/internal/storage"
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

// HandleUplinkData handles incoming (uplink) data.
func (a *ApplicationServerAPI) HandleUplinkData(ctx context.Context, req *as.HandleUplinkDataRequest) (*as.HandleUplinkDataResponse, error) {
	var appEUI, devEUI lorawan.EUI64
	copy(appEUI[:], req.AppEUI)
	copy(devEUI[:], req.DevEUI)

	d, err := storage.GetDevice(config.C.PostgreSQL.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get device error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	app, err := storage.GetApplication(config.C.PostgreSQL.DB, d.ApplicationID)
	if err != nil {
		errStr := fmt.Sprintf("get application error: %s", err)
		log.WithField("id", d.ApplicationID).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	da, err := storage.GetLastDeviceActivationForDevEUI(config.C.PostgreSQL.DB, d.DevEUI)
	if err != nil {
		errStr := fmt.Sprintf("get device-activation error: %s", err)
		log.WithField("dev_eui", d.DevEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	now := time.Now()
	d.LastSeenAt = &now
	d.DeviceStatusBattery = nil
	if req.DeviceStatusBattery != 256 {
		batt := int(req.DeviceStatusBattery)
		d.DeviceStatusBattery = &batt
	}
	d.DeviceStatusMargin = nil
	if req.DeviceStatusMargin != 256 {
		marg := int(req.DeviceStatusMargin)
		d.DeviceStatusMargin = &marg
	}
	err = storage.UpdateDevice(config.C.PostgreSQL.DB, &d)
	if err != nil {
		errStr := fmt.Sprintf("update device error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
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

	codecPL := codec.NewPayload(app.PayloadCodec, uint8(req.FPort), app.PayloadEncoderScript, app.PayloadDecoderScript)
	if codecPL != nil {
		if err := codecPL.UnmarshalBinary(b); err != nil {
			log.WithFields(log.Fields{
				"codec":          app.PayloadCodec,
				"application_id": app.ID,
				"f_port":         req.FPort,
				"f_cnt":          req.FCnt,
				"dev_eui":        d.DevEUI,
			}).WithError(err).Error("decode payload error")
		}
	}

	pl := handler.DataUpPayload{
		ApplicationID:       app.ID,
		ApplicationName:     app.Name,
		DeviceName:          d.Name,
		DevEUI:              devEUI,
		DeviceStatusBattery: d.DeviceStatusBattery,
		DeviceStatusMargin:  d.DeviceStatusMargin,
		RXInfo:              []handler.RXInfo{},
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
		FCnt:   req.FCnt,
		FPort:  uint8(req.FPort),
		Data:   b,
		Object: codecPL,
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

	err = config.C.ApplicationServer.Integration.Handler.SendDataUp(pl)
	if err != nil {
		errStr := fmt.Sprintf("send data up to handler error: %s", err)
		log.Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	return &as.HandleUplinkDataResponse{}, nil
}

// HandleDownlinkACK handles an ack on a downlink transmission.
func (a *ApplicationServerAPI) HandleDownlinkACK(ctx context.Context, req *as.HandleDownlinkACKRequest) (*as.HandleDownlinkACKResponse, error) {
	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEUI)

	d, err := storage.GetDevice(config.C.PostgreSQL.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get device error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}
	app, err := storage.GetApplication(config.C.PostgreSQL.DB, d.ApplicationID)
	if err != nil {
		errStr := fmt.Sprintf("get application error: %s", err)
		log.WithField("id", d.ApplicationID).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	dqm, err := storage.GetDeviceQueueMappingForDevEUIAndFCnt(config.C.PostgreSQL.DB, devEUI, req.FCnt)
	if err != nil {
		return nil, errToRPCError(err)
	}

	if err := storage.DeleteDeviceQueueMapping(config.C.PostgreSQL.DB, dqm.ID); err != nil {
		return nil, errToRPCError(err)
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
	}).Info("downlink device-queue item acknowledged")

	err = config.C.ApplicationServer.Integration.Handler.SendACKNotification(handler.ACKNotification{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		DeviceName:      d.Name,
		DevEUI:          devEUI,
		Reference:       dqm.Reference,
		Acknowledged:    req.Acknowledged,
		FCnt:            req.FCnt,
	})
	if err != nil {
		log.Errorf("send ack notification to handler error: %s", err)
	}

	return &as.HandleDownlinkACKResponse{}, nil
}

// HandleError handles an incoming error.
func (a *ApplicationServerAPI) HandleError(ctx context.Context, req *as.HandleErrorRequest) (*as.HandleErrorResponse, error) {
	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEUI)

	d, err := storage.GetDevice(config.C.PostgreSQL.DB, devEUI)
	if err != nil {
		errStr := fmt.Sprintf("get device error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}
	app, err := storage.GetApplication(config.C.PostgreSQL.DB, d.ApplicationID)
	if err != nil {
		errStr := fmt.Sprintf("get application error: %s", err)
		log.WithField("id", d.ApplicationID).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	log.WithFields(log.Fields{
		"type":    req.Type,
		"dev_eui": devEUI,
	}).Error(req.Error)

	err = config.C.ApplicationServer.Integration.Handler.SendErrorNotification(handler.ErrorNotification{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		DeviceName:      d.Name,
		DevEUI:          devEUI,
		Type:            req.Type.String(),
		Error:           req.Error,
		FCnt:            req.FCnt,
	})
	if err != nil {
		errStr := fmt.Sprintf("send error notification to handler error: %s", err)
		log.Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	return &as.HandleErrorResponse{}, nil
}

// HandleProprietaryUplink handles proprietary uplink payloads.
func (a *ApplicationServerAPI) HandleProprietaryUplink(ctx context.Context, req *as.HandleProprietaryUplinkRequest) (*as.HandleProprietaryUplinkResponse, error) {
	err := gwping.HandleReceivedPing(req)
	if err != nil {
		errStr := fmt.Sprintf("handle received ping error: %s", err)
		log.Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	return &as.HandleProprietaryUplinkResponse{}, nil
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
