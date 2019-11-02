package as

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"net"
	"time"

	keywrap "github.com/NickBall/go-aes-key-wrap"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/chirpstack-application-server/internal/api/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/applayer/clocksync"
	"github.com/brocaar/chirpstack-application-server/internal/applayer/fragmentation"
	"github.com/brocaar/chirpstack-application-server/internal/applayer/multicastsetup"
	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/eventlog"
	"github.com/brocaar/chirpstack-application-server/internal/gwping"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/loraserver/api/as"
	"github.com/brocaar/loraserver/api/common"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/gps"
)

var (
	bind    string
	caCert  string
	tlsCert string
	tlsKey  string
)

// Setup configures the package.
func Setup(conf config.Config) error {
	bind = conf.ApplicationServer.API.Bind
	caCert = conf.ApplicationServer.API.CACert
	tlsCert = conf.ApplicationServer.API.TLSCert
	tlsKey = conf.ApplicationServer.API.TLSKey

	log.WithFields(log.Fields{
		"bind":     bind,
		"ca_cert":  caCert,
		"tls_cert": tlsCert,
		"tls_key":  tlsKey,
	}).Info("api/as: starting application-server api")

	grpcOpts := helpers.GetgRPCServerOptions()
	if caCert != "" && tlsCert != "" && tlsKey != "" {
		creds, err := helpers.GetTransportCredentials(caCert, tlsCert, tlsKey, true)
		if err != nil {
			return errors.Wrap(err, "get transport credentials error")
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}
	server := grpc.NewServer(grpcOpts...)
	as.RegisterApplicationServerServiceServer(server, NewApplicationServerAPI())

	ln, err := net.Listen("tcp", bind)
	if err != nil {
		return errors.Wrap(err, "start application-server api listener error")
	}
	go server.Serve(ln)

	return nil
}

// ApplicationServerAPI implements the as.ApplicationServerServer interface.
type ApplicationServerAPI struct {
}

// NewApplicationServerAPI returns a new ApplicationServerAPI.
func NewApplicationServerAPI() *ApplicationServerAPI {
	return &ApplicationServerAPI{}
}

// HandleUplinkData handles incoming (uplink) data.
func (a *ApplicationServerAPI) HandleUplinkData(ctx context.Context, req *as.HandleUplinkDataRequest) (*empty.Empty, error) {
	if req.TxInfo == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "tx_info must not be nil")
	}

	var err error
	var d storage.Device
	var appEUI, devEUI lorawan.EUI64
	copy(appEUI[:], req.JoinEui)
	copy(devEUI[:], req.DevEui)

	err = storage.Transaction(func(tx sqlx.Ext) error {
		d, err = storage.GetDevice(ctx, tx, devEUI, true, true)
		if err != nil {
			grpc.Errorf(codes.Internal, "get device error: %s", err)
		}

		now := time.Now()
		dr := int(req.Dr)

		d.LastSeenAt = &now
		d.DR = &dr
		err = storage.UpdateDevice(ctx, tx, &d, true)
		if err != nil {
			return grpc.Errorf(codes.Internal, "update device error: %s", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	app, err := storage.GetApplication(ctx, storage.DB(), d.ApplicationID)
	if err != nil {
		errStr := fmt.Sprintf("get application error: %s", err)
		log.WithField("id", d.ApplicationID).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	dp, err := storage.GetDeviceProfile(ctx, storage.DB(), d.DeviceProfileID, false, true)
	if err != nil {
		log.WithError(err).WithField("id", d.DeviceProfileID).Error("get device-profile error")
		return nil, grpc.Errorf(codes.Internal, "get device-profile error: %s", err)
	}

	if req.DeviceActivationContext != nil {
		if err := handleDeviceActivation(ctx, d, app, req); err != nil {
			return nil, helpers.ErrToRPCError(err)
		}
	}

	da, err := storage.GetLastDeviceActivationForDevEUI(ctx, storage.DB(), d.DevEUI)
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

	// payload is handled by the ChirpStack Application Server internal applayer
	var internalApplayer bool

	if err := storage.Transaction(func(db sqlx.Ext) error {
		switch req.FPort {
		case 200:
			internalApplayer = true
			if err := multicastsetup.HandleRemoteMulticastSetupCommand(ctx, db, d.DevEUI, b); err != nil {
				return grpc.Errorf(codes.Internal, "handle remote multicast setup command error: %s", err)
			}
		case 201:
			internalApplayer = true
			if err := fragmentation.HandleRemoteFragmentationSessionCommand(ctx, db, d.DevEUI, b); err != nil {
				return grpc.Errorf(codes.Internal, "handle remote fragmentation session command error: %s", err)
			}
		case 202:
			internalApplayer = true

			var timeSinceGPSEpoch time.Duration
			var timeField time.Time

			for _, rxInfo := range req.RxInfo {
				if rxInfo.TimeSinceGpsEpoch != nil {
					timeSinceGPSEpoch, err = ptypes.Duration(rxInfo.TimeSinceGpsEpoch)
					if err != nil {
						log.WithError(err).Error("time since gps epoch to duration error")
						continue
					}
				} else if rxInfo.Time != nil {
					timeField, err = ptypes.Timestamp(rxInfo.Time)
					if err != nil {
						log.WithError(err).Error("time to timestamp error")
						continue
					}
				}
			}

			// fallback on time field when time since GPS epoch is not available
			if timeSinceGPSEpoch == 0 {
				// fallback on current server time when time field is not available
				if timeField.IsZero() {
					timeField = time.Now()
				}
				timeSinceGPSEpoch = gps.Time(timeField).TimeSinceGPSEpoch()
			}

			if err := clocksync.HandleClockSyncCommand(ctx, db, d.DevEUI, timeSinceGPSEpoch, b); err != nil {
				return grpc.Errorf(codes.Internal, "handle clocksync command error: %s", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if internalApplayer {
		return &empty.Empty{}, nil
	}

	var object interface{}

	// TODO: in the next major release, remove this and always use the
	// device-profile codec fields.
	payloadCodec := app.PayloadCodec
	payloadEncoderScript := app.PayloadEncoderScript
	payloadDecoderScript := app.PayloadDecoderScript

	if dp.PayloadCodec != "" {
		payloadCodec = dp.PayloadCodec
		payloadEncoderScript = dp.PayloadEncoderScript
		payloadDecoderScript = dp.PayloadDecoderScript
	}

	codecPL := codec.NewPayload(payloadCodec, uint8(req.FPort), payloadEncoderScript, payloadDecoderScript)
	if codecPL != nil {
		start := time.Now()
		if err := codecPL.DecodeBytes(b); err != nil {
			log.WithFields(log.Fields{
				"codec":          app.PayloadCodec,
				"application_id": app.ID,
				"f_port":         req.FPort,
				"f_cnt":          req.FCnt,
				"dev_eui":        d.DevEUI,
			}).WithError(err).Error("decode payload error")

			errNotification := integration.ErrorNotification{
				ApplicationID:   d.ApplicationID,
				ApplicationName: app.Name,
				DeviceName:      d.Name,
				DevEUI:          d.DevEUI,
				Type:            "CODEC",
				Error:           err.Error(),
				FCnt:            req.FCnt,
				Tags:            make(map[string]string),
				Variables:       make(map[string]string),
			}

			for k, v := range d.Tags.Map {
				if v.Valid {
					errNotification.Tags[k] = v.String
				}
			}

			for k, v := range d.Variables.Map {
				if v.Valid {
					errNotification.Variables[k] = v.String
				}
			}

			if err := eventlog.LogEventForDevice(d.DevEUI, eventlog.EventLog{
				Type:    eventlog.Error,
				Payload: errNotification,
			}); err != nil {
				log.WithError(err).Error("log event for device error")
			}

			if err := integration.Integration().SendErrorNotification(ctx, errNotification); err != nil {
				log.WithError(err).Error("send error notification to integration error")
			}
		} else {
			log.WithFields(log.Fields{
				"application_id": app.ID,
				"codec":          app.PayloadCodec,
				"duration":       time.Since(start),
			}).Debug("payload codec completed Decode execution")
			object = codecPL.Object()
		}
	}

	pl := integration.DataUpPayload{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		DeviceName:      d.Name,
		DevEUI:          devEUI,
		RXInfo:          []integration.RXInfo{},
		TXInfo: integration.TXInfo{
			Frequency: int(req.TxInfo.Frequency),
			DR:        int(req.Dr),
		},
		ADR:       req.Adr,
		FCnt:      req.FCnt,
		FPort:     uint8(req.FPort),
		Data:      b,
		Object:    object,
		Tags:      make(map[string]string),
		Variables: make(map[string]string),
	}

	// set tags and variables
	for k, v := range d.Tags.Map {
		if v.Valid {
			pl.Tags[k] = v.String
		}
	}

	for k, v := range d.Variables.Map {
		if v.Valid {
			pl.Variables[k] = v.String
		}
	}

	// collect gateway data of receiving gateways (e.g. gateway name)
	var macs []lorawan.EUI64
	for _, rxInfo := range req.RxInfo {
		var mac lorawan.EUI64
		copy(mac[:], rxInfo.GatewayId)
		macs = append(macs, mac)
	}
	gws, err := storage.GetGatewaysForMACs(ctx, storage.DB(), macs)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "get gateways for macs error: %s", err)
	}

	for _, rxInfo := range req.RxInfo {
		var mac lorawan.EUI64
		copy(mac[:], rxInfo.GatewayId)

		var uplinkID uuid.UUID
		copy(uplinkID[:], rxInfo.UplinkId)

		row := integration.RXInfo{
			GatewayID: mac,
			UplinkID:  uplinkID,
			RSSI:      int(rxInfo.Rssi),
			LoRaSNR:   rxInfo.LoraSnr,
		}

		if rxInfo.Location != nil {
			row.Location = &integration.Location{
				Latitude:  rxInfo.Location.Latitude,
				Longitude: rxInfo.Location.Longitude,
				Altitude:  rxInfo.Location.Altitude,
			}
		}

		if gw, ok := gws[mac]; ok {
			row.Name = gw.Name
		}

		if rxInfo.Time != nil {
			ts, err := ptypes.Timestamp(rxInfo.Time)
			if err != nil {
				log.WithField("dev_eui", devEUI).WithError(err).Error("parse timestamp error")
			} else {
				row.Time = &ts
			}
		}

		pl.RXInfo = append(pl.RXInfo, row)
	}

	err = eventlog.LogEventForDevice(devEUI, eventlog.EventLog{
		Type:    eventlog.Uplink,
		Payload: pl,
	})
	if err != nil {
		log.WithError(err).Error("log event for device error")
	}

	err = integration.Integration().SendDataUp(ctx, pl)
	if err != nil {
		log.WithError(err).Error("send uplink data to integration error")
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	return &empty.Empty{}, nil
}

// HandleDownlinkACK handles an ack on a downlink transmission.
func (a *ApplicationServerAPI) HandleDownlinkACK(ctx context.Context, req *as.HandleDownlinkACKRequest) (*empty.Empty, error) {
	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEui)

	d, err := storage.GetDevice(ctx, storage.DB(), devEUI, false, true)
	if err != nil {
		errStr := fmt.Sprintf("get device error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}
	app, err := storage.GetApplication(ctx, storage.DB(), d.ApplicationID)
	if err != nil {
		errStr := fmt.Sprintf("get application error: %s", err)
		log.WithField("id", d.ApplicationID).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
	}).Info("downlink device-queue item acknowledged")

	pl := integration.ACKNotification{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		DeviceName:      d.Name,
		DevEUI:          devEUI,
		Acknowledged:    req.Acknowledged,
		FCnt:            req.FCnt,
		Tags:            make(map[string]string),
		Variables:       make(map[string]string),
	}

	// set tags and variables
	for k, v := range d.Tags.Map {
		if v.Valid {
			pl.Tags[k] = v.String
		}
	}
	for k, v := range d.Variables.Map {
		if v.Valid {
			pl.Variables[k] = v.String
		}
	}

	err = eventlog.LogEventForDevice(devEUI, eventlog.EventLog{
		Type:    eventlog.ACK,
		Payload: pl,
	})
	if err != nil {
		log.WithError(err).Error("log event for device error")
	}

	err = integration.Integration().SendACKNotification(ctx, pl)
	if err != nil {
		log.Errorf("send ack notification to integration error: %s", err)
	}

	return &empty.Empty{}, nil
}

// HandleError handles an incoming error.
func (a *ApplicationServerAPI) HandleError(ctx context.Context, req *as.HandleErrorRequest) (*empty.Empty, error) {
	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEui)

	d, err := storage.GetDevice(ctx, storage.DB(), devEUI, false, true)
	if err != nil {
		errStr := fmt.Sprintf("get device error: %s", err)
		log.WithField("dev_eui", devEUI).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}
	app, err := storage.GetApplication(ctx, storage.DB(), d.ApplicationID)
	if err != nil {
		errStr := fmt.Sprintf("get application error: %s", err)
		log.WithField("id", d.ApplicationID).Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	log.WithFields(log.Fields{
		"type":    req.Type,
		"dev_eui": devEUI,
	}).Error(req.Error)

	pl := integration.ErrorNotification{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		DeviceName:      d.Name,
		DevEUI:          devEUI,
		Type:            req.Type.String(),
		Error:           req.Error,
		FCnt:            req.FCnt,
		Tags:            make(map[string]string),
		Variables:       make(map[string]string),
	}

	// set tags and variables
	for k, v := range d.Tags.Map {
		if v.Valid {
			pl.Tags[k] = v.String
		}
	}

	for k, v := range d.Variables.Map {
		if v.Valid {
			pl.Variables[k] = v.String
		}
	}

	err = eventlog.LogEventForDevice(devEUI, eventlog.EventLog{
		Type:    eventlog.Error,
		Payload: pl,
	})
	if err != nil {
		log.WithError(err).Error("log event for device error")
	}

	err = integration.Integration().SendErrorNotification(ctx, pl)
	if err != nil {
		errStr := fmt.Sprintf("send error notification to integration error: %s", err)
		log.Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	return &empty.Empty{}, nil
}

// HandleProprietaryUplink handles proprietary uplink payloads.
func (a *ApplicationServerAPI) HandleProprietaryUplink(ctx context.Context, req *as.HandleProprietaryUplinkRequest) (*empty.Empty, error) {
	if req.TxInfo == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "tx_info must not be nil")
	}

	err := gwping.HandleReceivedPing(ctx, req)
	if err != nil {
		errStr := fmt.Sprintf("handle received ping error: %s", err)
		log.Error(errStr)
		return nil, grpc.Errorf(codes.Internal, errStr)
	}

	return &empty.Empty{}, nil
}

// SetDeviceStatus updates the device-status for the given device.
func (a *ApplicationServerAPI) SetDeviceStatus(ctx context.Context, req *as.SetDeviceStatusRequest) (*empty.Empty, error) {
	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEui)

	var d storage.Device
	var err error

	err = storage.Transaction(func(tx sqlx.Ext) error {
		d, err = storage.GetDevice(ctx, tx, devEUI, true, true)
		if err != nil {
			return helpers.ErrToRPCError(errors.Wrap(err, "get device error"))
		}

		marg := int(req.Margin)
		d.DeviceStatusMargin = &marg

		if req.BatteryLevelUnavailable {
			d.DeviceStatusBattery = nil
			d.DeviceStatusExternalPower = false
		} else if req.ExternalPowerSource {
			d.DeviceStatusExternalPower = true
			d.DeviceStatusBattery = nil
		} else {
			d.DeviceStatusExternalPower = false
			d.DeviceStatusBattery = &req.BatteryLevel
		}

		if err = storage.UpdateDevice(ctx, tx, &d, true); err != nil {
			return helpers.ErrToRPCError(errors.Wrap(err, "update device error"))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	app, err := storage.GetApplication(ctx, storage.DB(), d.ApplicationID)
	if err != nil {
		return nil, helpers.ErrToRPCError(errors.Wrap(err, "get application error"))
	}

	pl := integration.StatusNotification{
		ApplicationID:           app.ID,
		ApplicationName:         app.Name,
		DeviceName:              d.Name,
		DevEUI:                  d.DevEUI,
		Battery:                 int(req.Battery),
		Margin:                  int(req.Margin),
		ExternalPowerSource:     req.ExternalPowerSource,
		BatteryLevel:            float32(math.Round(float64(req.BatteryLevel*100))) / 100,
		BatteryLevelUnavailable: req.BatteryLevelUnavailable,
		Tags:                    make(map[string]string),
		Variables:               make(map[string]string),
	}

	// set tags and variables
	for k, v := range d.Tags.Map {
		if v.Valid {
			pl.Tags[k] = v.String
		}
	}
	for k, v := range d.Variables.Map {
		if v.Valid {
			pl.Variables[k] = v.String
		}
	}

	err = eventlog.LogEventForDevice(d.DevEUI, eventlog.EventLog{
		Type:    eventlog.Status,
		Payload: pl,
	})
	if err != nil {
		log.WithError(err).Error("log event for device error")
	}

	err = integration.Integration().SendStatusNotification(ctx, pl)
	if err != nil {
		return nil, helpers.ErrToRPCError(errors.Wrap(err, "send status notification to handler error"))
	}

	return &empty.Empty{}, nil
}

// SetDeviceLocation updates the device-location.
func (a *ApplicationServerAPI) SetDeviceLocation(ctx context.Context, req *as.SetDeviceLocationRequest) (*empty.Empty, error) {
	if req.Location == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "location must not be nil")
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEui)

	var d storage.Device
	var err error

	err = storage.Transaction(func(tx sqlx.Ext) error {
		d, err = storage.GetDevice(ctx, tx, devEUI, true, true)
		if err != nil {
			return helpers.ErrToRPCError(errors.Wrap(err, "get device error"))
		}

		d.Latitude = &req.Location.Latitude
		d.Longitude = &req.Location.Longitude
		d.Altitude = &req.Location.Altitude

		if err = storage.UpdateDevice(ctx, tx, &d, true); err != nil {
			return helpers.ErrToRPCError(errors.Wrap(err, "update device error"))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	app, err := storage.GetApplication(ctx, storage.DB(), d.ApplicationID)
	if err != nil {
		return nil, helpers.ErrToRPCError(errors.Wrap(err, "get application error"))
	}

	pl := integration.LocationNotification{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		DeviceName:      d.Name,
		DevEUI:          d.DevEUI,
		Location: integration.Location{
			Latitude:  req.Location.Latitude,
			Longitude: req.Location.Longitude,
			Altitude:  req.Location.Altitude,
		},
		Tags:      make(map[string]string),
		Variables: make(map[string]string),
	}

	// set tags and variables
	for k, v := range d.Tags.Map {
		if v.Valid {
			pl.Tags[k] = v.String
		}
	}
	for k, v := range d.Variables.Map {
		if v.Valid {
			pl.Variables[k] = v.String
		}
	}

	err = eventlog.LogEventForDevice(d.DevEUI, eventlog.EventLog{
		Type:    eventlog.Location,
		Payload: pl,
	})
	if err != nil {
		log.WithError(err).Error("log event for device error")
	}

	err = integration.Integration().SendLocationNotification(ctx, pl)
	if err != nil {
		return nil, helpers.ErrToRPCError(errors.Wrap(err, "send location notification to handler error"))
	}

	return &empty.Empty{}, nil
}

// HandleGatewayStats handles the given gateway stats.
func (a *ApplicationServerAPI) HandleGatewayStats(ctx context.Context, req *as.HandleGatewayStatsRequest) (*empty.Empty, error) {
	var gatewayID lorawan.EUI64
	copy(gatewayID[:], req.GatewayId)

	ts, err := ptypes.Timestamp(req.Time)
	if err != nil {
		return nil, helpers.ErrToRPCError(errors.Wrap(err, "time error"))
	}

	err = storage.Transaction(func(tx sqlx.Ext) error {
		gw, err := storage.GetGateway(ctx, tx, gatewayID, true)
		if err != nil {
			return helpers.ErrToRPCError(errors.Wrap(err, "get gateway error"))
		}

		if gw.FirstSeenAt == nil {
			gw.FirstSeenAt = &ts
		}
		gw.LastSeenAt = &ts

		if loc := req.Location; loc != nil {
			gw.Latitude = loc.Latitude
			gw.Longitude = loc.Longitude
			gw.Altitude = loc.Altitude
		}

		if err := storage.UpdateGateway(ctx, tx, &gw); err != nil {
			return helpers.ErrToRPCError(errors.Wrap(err, "update gateway error"))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	metrics := storage.MetricsRecord{
		Time: ts,
		Metrics: map[string]float64{
			"rx_count":    float64(req.RxPacketsReceived),
			"rx_ok_count": float64(req.RxPacketsReceivedOk),
			"tx_count":    float64(req.TxPacketsReceived),
			"tx_ok_count": float64(req.TxPacketsEmitted),
		},
	}
	if err := storage.SaveMetrics(ctx, storage.RedisPool(), fmt.Sprintf("gw:%s", gatewayID), metrics); err != nil {
		return nil, helpers.ErrToRPCError(errors.Wrap(err, "save metrics error"))
	}

	return &empty.Empty{}, nil
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

func handleDeviceActivation(ctx context.Context, d storage.Device, app storage.Application, req *as.HandleUplinkDataRequest) error {
	daCtx := req.DeviceActivationContext

	if daCtx.AppSKey == nil {
		return errors.New("AppSKey must not be nil")
	}

	key, err := unwrapASKey(daCtx.AppSKey)
	if err != nil {
		return errors.Wrap(err, "unwrap appSKey error")
	}

	da := storage.DeviceActivation{
		DevEUI:  d.DevEUI,
		AppSKey: key,
	}
	copy(da.DevAddr[:], daCtx.DevAddr)

	if err = storage.CreateDeviceActivation(ctx, storage.DB(), &da); err != nil {
		return errors.Wrap(err, "create device-activation error")
	}

	pl := integration.JoinNotification{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		DevEUI:          d.DevEUI,
		DeviceName:      d.Name,
		DevAddr:         da.DevAddr,
		RXInfo:          []integration.RXInfo{},
		TXInfo: integration.TXInfo{
			Frequency: int(req.TxInfo.Frequency),
			DR:        int(req.Dr),
		},
		Tags:      make(map[string]string),
		Variables: make(map[string]string),
	}

	// set tags and variables
	for k, v := range d.Tags.Map {
		if v.Valid {
			pl.Tags[k] = v.String
		}
	}
	for k, v := range d.Variables.Map {
		if v.Valid {
			pl.Variables[k] = v.String
		}
	}

	// collect gateway data of receiving gateways (e.g. gateway name)
	var macs []lorawan.EUI64
	for _, rxInfo := range req.RxInfo {
		var mac lorawan.EUI64
		copy(mac[:], rxInfo.GatewayId)
		macs = append(macs, mac)
	}
	gws, err := storage.GetGatewaysForMACs(ctx, storage.DB(), macs)
	if err != nil {
		return errors.Wrap(err, "get gateways for macs error")
	}

	for _, rxInfo := range req.RxInfo {
		var mac lorawan.EUI64
		var uplinkID uuid.UUID
		copy(mac[:], rxInfo.GatewayId)
		copy(uplinkID[:], rxInfo.UplinkId)

		row := integration.RXInfo{
			GatewayID: mac,
			UplinkID:  uplinkID,
			RSSI:      int(rxInfo.Rssi),
			LoRaSNR:   rxInfo.LoraSnr,
		}

		if rxInfo.Location != nil {
			row.Location = &integration.Location{
				Latitude:  rxInfo.Location.Latitude,
				Longitude: rxInfo.Location.Longitude,
				Altitude:  rxInfo.Location.Altitude,
			}
		}

		if gw, ok := gws[mac]; ok {
			row.Name = gw.Name
		}

		if rxInfo.Time != nil {
			ts, err := ptypes.Timestamp(rxInfo.Time)
			if err != nil {
				log.WithFields(log.Fields{
					"dev_eui": d.DevEUI,
					"ctx_id":  ctx.Value(logging.ContextIDKey),
				}).WithError(err).Error("parse timestamp error")
			} else {
				row.Time = &ts
			}
		}

		pl.RXInfo = append(pl.RXInfo, row)
	}

	err = eventlog.LogEventForDevice(d.DevEUI, eventlog.EventLog{
		Type:    eventlog.Join,
		Payload: pl,
	})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"ctx_id": ctx.Value(logging.ContextIDKey),
		}).Error("log event for device error")
	}

	err = integration.Integration().SendJoinNotification(ctx, pl)
	if err != nil {
		return errors.Wrap(err, "send join notification error")
	}

	return nil
}

func unwrapASKey(ke *common.KeyEnvelope) (lorawan.AES128Key, error) {
	var key lorawan.AES128Key

	if ke.KekLabel == "" {
		copy(key[:], ke.AesKey)
		return key, nil
	}

	for i := range config.C.JoinServer.KEK.Set {
		if config.C.JoinServer.KEK.Set[i].Label == ke.KekLabel {
			kek, err := hex.DecodeString(config.C.JoinServer.KEK.Set[i].KEK)
			if err != nil {
				return key, errors.Wrap(err, "decode kek error")
			}

			block, err := aes.NewCipher(kek)
			if err != nil {
				return key, errors.Wrap(err, "new cipher error")
			}

			b, err := keywrap.Unwrap(block, ke.AesKey)
			if err != nil {
				return key, errors.Wrap(err, "key unwrap error")
			}

			copy(key[:], b)
			return key, nil
		}
	}

	return key, fmt.Errorf("unknown kek label: %s", ke.KekLabel)
}
