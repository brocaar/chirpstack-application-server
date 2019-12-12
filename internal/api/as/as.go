package as

import (
	"fmt"
	"math"
	"net"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/chirpstack-api/go/v3/as"
	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/api/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/eventlog"
	"github.com/brocaar/chirpstack-application-server/internal/events/uplink"
	"github.com/brocaar/chirpstack-application-server/internal/gwping"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
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
	if err := uplink.Handle(ctx, *req); err != nil {
		return nil, grpc.Errorf(codes.Internal, "handle uplink data error: %s", err)
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

	pl := pb.AckEvent{
		ApplicationId:   uint64(app.ID),
		ApplicationName: app.Name,
		DeviceName:      d.Name,
		DevEui:          devEUI[:],
		Acknowledged:    req.Acknowledged,
		FCnt:            req.FCnt,
		Tags:            make(map[string]string),
	}

	// set tags
	for k, v := range d.Tags.Map {
		if v.Valid {
			pl.Tags[k] = v.String
		}
	}

	vars := make(map[string]string)
	for k, v := range d.Variables.Map {
		if v.Valid {
			vars[k] = v.String
		}
	}

	err = eventlog.LogEventForDevice(devEUI, eventlog.ACK, &pl)
	if err != nil {
		log.WithError(err).Error("log event for device error")
	}

	err = integration.Integration().SendACKNotification(ctx, vars, pl)
	if err != nil {
		log.WithError(err).Error("send ack event error")
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

	var errType pb.ErrorType
	switch req.Type {
	case as.ErrorType_OTAA:
		errType = pb.ErrorType_OTAA
	case as.ErrorType_DATA_UP_FCNT:
		errType = pb.ErrorType_UPLINK_FCNT
	case as.ErrorType_DATA_UP_MIC:
		errType = pb.ErrorType_UPLINK_MIC
	case as.ErrorType_DEVICE_QUEUE_ITEM_SIZE:
		errType = pb.ErrorType_DOWNLINK_PAYLOAD_SIZE
	case as.ErrorType_DEVICE_QUEUE_ITEM_FCNT:
		errType = pb.ErrorType_DOWNLINK_FCNT
	}

	pl := pb.ErrorEvent{
		ApplicationId:   uint64(app.ID),
		ApplicationName: app.Name,
		DeviceName:      d.Name,
		DevEui:          devEUI[:],
		Type:            errType,
		Error:           req.Error,
		FCnt:            req.FCnt,
		Tags:            make(map[string]string),
	}

	// set tags
	for k, v := range d.Tags.Map {
		if v.Valid {
			pl.Tags[k] = v.String
		}
	}

	vars := make(map[string]string)
	for k, v := range d.Variables.Map {
		if v.Valid {
			vars[k] = v.String
		}
	}

	err = eventlog.LogEventForDevice(devEUI, eventlog.Error, &pl)
	if err != nil {
		log.WithError(err).Error("log event for device error")
	}

	err = integration.Integration().SendErrorNotification(ctx, vars, pl)
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

	pl := pb.StatusEvent{
		ApplicationId:           uint64(app.ID),
		ApplicationName:         app.Name,
		DeviceName:              d.Name,
		DevEui:                  d.DevEUI[:],
		Margin:                  uint32(req.Margin),
		ExternalPowerSource:     req.ExternalPowerSource,
		BatteryLevel:            float32(math.Round(float64(req.BatteryLevel*100))) / 100,
		BatteryLevelUnavailable: req.BatteryLevelUnavailable,
		Tags:                    make(map[string]string),
	}

	// set tags
	for k, v := range d.Tags.Map {
		if v.Valid {
			pl.Tags[k] = v.String
		}
	}

	vars := make(map[string]string)
	for k, v := range d.Variables.Map {
		if v.Valid {
			vars[k] = v.String
		}
	}

	err = eventlog.LogEventForDevice(d.DevEUI, eventlog.Status, &pl)
	if err != nil {
		log.WithError(err).Error("log event for device error")
	}

	err = integration.Integration().SendStatusNotification(ctx, vars, pl)
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

	pl := pb.LocationEvent{
		ApplicationId:   uint64(app.ID),
		ApplicationName: app.Name,
		DeviceName:      d.Name,
		DevEui:          d.DevEUI[:],
		Location:        req.Location,
		Tags:            make(map[string]string),
	}

	// set tags
	for k, v := range d.Tags.Map {
		if v.Valid {
			pl.Tags[k] = v.String
		}
	}

	vars := make(map[string]string)
	for k, v := range d.Variables.Map {
		if v.Valid {
			vars[k] = v.String
		}
	}

	err = eventlog.LogEventForDevice(d.DevEUI, eventlog.Location, &pl)
	if err != nil {
		log.WithError(err).Error("log event for device error")
	}

	err = integration.Integration().SendLocationNotification(ctx, vars, pl)
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
