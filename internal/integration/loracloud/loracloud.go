package loracloud

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-api/go/v3/as/integration"
	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-api/go/v3/common"
	gw "github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/integration/loracloud/client/das"
	"github.com/brocaar/chirpstack-application-server/internal/integration/loracloud/client/geolocation"
	"github.com/brocaar/chirpstack-application-server/internal/integration/loracloud/client/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/gps"
)

// Config contains the LoRaCloud integration configuration.
type Config struct {
	// Geolocation.
	Geolocation                 bool   `json:"geolocation"`
	GeolocationToken            string `json:"geolocationToken"`
	GeolocationBufferTTL        int    `json:"geolocationBufferTTL"`
	GeolocationMinBufferSize    int    `json:"geolocationMinBufferSize"`
	GeolocationTDOA             bool   `json:"geolocationTDOA"`
	GeolocationRSSI             bool   `json:"geolocationRSSI"`
	GeolocationGNSS             bool   `json:"geolocationGNSS"`
	GeolocationGNSSPayloadField string `json:"geolocationGNSSPayloadField"`
	GeolocationGNSSUseRxTime    bool   `json:"geolicationGNSSUseRxTime"`
	GeolocationWifi             bool   `json:"geolocationWifi"`
	GeolocationWifiPayloadField string `json:"geolocationWifiPayloadField"`

	// Device Application Services.
	DAS                          bool   `json:"das"`
	DASToken                     string `json:"dasToken"`
	DASModemPort                 uint8  `json:"dasModemPort"`
	DASGNSSPort                  uint8  `json:"dasGNSSPort"`
	DASGNSSUseRxTime             bool   `json:"dasGNSSUseRxTime"`
	DASStreamingGeolocWorkaround bool   `json:"dasStreamingGeolocWorkaround"`
}

// Integration implements a LoRaCloud Integration.
type Integration struct {
	config         Config
	geolocationURI string
	dasURI         string
}

// New creates a new LoRaCloud integration.
func New(conf Config) (*Integration, error) {
	return &Integration{
		config:         conf,
		geolocationURI: os.Getenv("LORACLOUD_GEOLOCATION_URI"), // for testing
		dasURI:         os.Getenv("LORACLOUD_DAS_URI"),         // for testing
	}, nil
}

func (i *Integration) getGeolocationURI() string {
	// for testing
	if i.geolocationURI != "" {
		return i.geolocationURI
	}
	return "https://mgs.loracloud.com"
}

func (i *Integration) getDasURI() string {
	// for testing
	if i.dasURI != "" {
		return i.dasURI
	}
	return "https://mgs.loracloud.com"
}

// HandleUplinkEvent handles the Uplinkevent.
func (i *Integration) HandleUplinkEvent(ctx context.Context, ii models.Integration, vars map[string]string, pl pb.UplinkEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	// handle geolocation
	err := func() error {
		// update and get geoloc buffer
		geolocBuffer, err := i.updateGeolocBuffer(ctx, devEUI, pl)
		if err != nil {
			return errors.Wrap(err, "update geolocation buffer error")
		}

		// do geolocation
		uplinkIDs, loc, err := i.geolocation(ctx, devEUI, geolocBuffer, pl)
		if err != nil {
			return errors.Wrap(err, "geolocation error")
		}

		// if it resolved to a location, send it to integrations
		if loc != nil {
			var fCnt uint32
			if len(uplinkIDs) == 0 {
				fCnt = pl.FCnt
			}
			if err := ii.HandleLocationEvent(ctx, vars, pb.LocationEvent{
				ApplicationId:   pl.ApplicationId,
				ApplicationName: pl.ApplicationName,
				DeviceName:      pl.DeviceName,
				DevEui:          pl.DevEui,
				Tags:            pl.Tags,
				Location:        loc,
				UplinkIds:       uplinkIDs,
				FCnt:            fCnt,
				PublishedAt:     ptypes.TimestampNow(),
			}); err != nil {
				return errors.Wrap(err, "handle location event error")
			}
		}

		return nil
	}()
	if err != nil {
		log.WithError(err).Error("integration/loracloud: geolocation error")
	}

	// handle das
	if i.config.DAS {
		err := func() error {
			if pl.FPort == uint32(i.config.DASModemPort) {
				// handle DAS modem message
				if err := i.dasModem(ctx, vars, devEUI, pl, ii); err != nil {
					return errors.Wrap(err, "das modem message error")
				}
			} else if pl.FPort == uint32(i.config.DASGNSSPort) {
				// handle DAS gnss message
				if err := i.dasGNSS(ctx, vars, devEUI, pl, ii); err != nil {
					return errors.Wrap(err, "das gnss message error")
				}
			} else {
				// uplink meta-data
				if err := i.dasUplinkMetaData(ctx, vars, devEUI, pl, ii); err != nil {
					return errors.Wrap(err, "das meta-data message error")
				}
			}

			return nil
		}()
		if err != nil {
			log.WithError(err).Error("integration/loracloud: das error")
		}
	}

	return nil
}

// HandleJoinEvent is not implemented.
func (i *Integration) HandleJoinEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.JoinEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	// handle das
	if i.config.DAS {
		if err := i.dasJoin(ctx, devEUI, pl); err != nil {
			log.WithError(err).Error("integration/loracloud: das error")
		}
	}

	return nil
}

// HandleAckEvent is not implemented.
func (i *Integration) HandleAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.AckEvent) error {
	return nil
}

// HandleErrorEvent is not implemented.
func (i *Integration) HandleErrorEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.ErrorEvent) error {
	return nil
}

// HandleStatusEvent is not implemented.
func (i *Integration) HandleStatusEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.StatusEvent) error {
	return nil
}

// HandleLocationEvent is not implemented.
func (i *Integration) HandleLocationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.LocationEvent) error {
	return nil
}

// HandleTxAckEvent is not implemented.
func (i *Integration) HandleTxAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.TxAckEvent) error {
	return nil
}

// HandleIntegrationEvent is not implemented.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.IntegrationEvent) error {
	return nil
}

// DataDownChan returns nil.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return nil
}

// Close is not implemented.
func (i *Integration) Close() error {
	return nil
}

func (i *Integration) updateGeolocBuffer(ctx context.Context, devEUI lorawan.EUI64, pl pb.UplinkEvent) ([][]*gw.UplinkRXInfo, error) {
	// Do not trigger geolocation if there are less than 3 gateways.
	if len(pl.RxInfo) < 3 {
		return nil, nil
	}

	// read the geoloc buffer
	geolocBuffer, err := GetGeolocBuffer(ctx, devEUI, time.Duration(i.config.GeolocationBufferTTL)*time.Second)
	if err != nil {
		return nil, errors.Wrap(err, "get geoloc buffer error")
	}
	geolocBuffer = append(geolocBuffer, pl.RxInfo)

	if err := SaveGeolocBuffer(ctx, devEUI, geolocBuffer, time.Duration(i.config.GeolocationBufferTTL)*time.Second); err != nil {
		return nil, errors.Wrap(err, "save geoloc buffer error")
	}

	return geolocBuffer, nil
}

func (i *Integration) geolocation(ctx context.Context, devEUI lorawan.EUI64, geolocBuffer [][]*gw.UplinkRXInfo, pl pb.UplinkEvent) ([][]byte, *common.Location, error) {
	if i.config.GeolocationGNSS {
		gnssPL, err := getBytesFromJSONObject(i.config.GeolocationGNSSPayloadField, pl.ObjectJson)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"dev_eui":       devEUI,
				"ctx_id":        ctx.Value(logging.ContextIDKey),
				"payload_field": i.config.GeolocationGNSSPayloadField,
			}).Error("integration/loracloud: get gnss bytes from object error")
			return nil, nil, nil
		}

		if len(gnssPL) == 0 {
			log.WithFields(log.Fields{
				"dev_eui":       devEUI,
				"ctx_id":        ctx.Value(logging.ContextIDKey),
				"payload_field": i.config.GeolocationGNSSPayloadField,
			}).Debug("integration/loracloud: no gnss bytes found in object")
		} else {
			loc, err := i.gnssLR1110Geolocation(ctx, devEUI, pl.RxInfo, gnssPL)
			return nil, loc, err
		}
	}

	if i.config.GeolocationWifi {
		wifiAPs, err := getWifiAccessPointsFromJSONObject(i.config.GeolocationWifiPayloadField, pl.ObjectJson)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"dev_eui":       devEUI,
				"ctx_id":        ctx.Value(logging.ContextIDKey),
				"payload_field": i.config.GeolocationWifiPayloadField,
			}).Error("integration/loracloud: get wifi access-points from object error")
			return nil, nil, nil
		}

		if len(wifiAPs) == 0 {
			log.WithFields(log.Fields{
				"dev_eui":       devEUI,
				"ctx_id":        ctx.Value(logging.ContextIDKey),
				"payload_field": i.config.GeolocationWifiPayloadField,
			}).Debug("integration/loracloud: no wifi access-points found in object")
		} else {
			loc, err := i.wifiTDOAGeolocation(ctx, devEUI, pl.RxInfo, wifiAPs)
			return nil, loc, err
		}
	}

	if i.config.GeolocationTDOA {
		tdoaFiltered := filterOnFineTimestamp(geolocBuffer, 3)
		if len(tdoaFiltered) == 0 || len(tdoaFiltered) < i.config.GeolocationMinBufferSize {
			log.WithFields(log.Fields{
				"dev_eui":              devEUI,
				"ctx_id":               ctx.Value(logging.ContextIDKey),
				"tdoa_suitable_frames": len(tdoaFiltered),
			}).Debug("integration/loracloud: not enough meta-data for tdoa geolocation")
		} else {
			var uplinkIDs [][]byte
			for i := range tdoaFiltered {
				for j := range tdoaFiltered[i] {
					uplinkIDs = append(uplinkIDs, tdoaFiltered[i][j].GetUplinkId())
				}
			}

			loc, err := i.tdoaGeolocation(ctx, devEUI, tdoaFiltered)
			return uplinkIDs, loc, err
		}
	}

	if i.config.GeolocationRSSI {
		if len(geolocBuffer) == 0 || len(geolocBuffer) < i.config.GeolocationMinBufferSize {
			log.WithFields(log.Fields{
				"dev_eui": devEUI,
				"ctx_id":  ctx.Value(logging.ContextIDKey),
				"frames":  len(geolocBuffer),
			}).Debug("integration/loracloud: not enough meta-data for rssi geolocation")
		} else {
			var uplinkIDs [][]byte
			for i := range geolocBuffer {
				for j := range geolocBuffer[i] {
					uplinkIDs = append(uplinkIDs, geolocBuffer[i][j].GetUplinkId())
				}
			}
			loc, err := i.rssiGeolocation(ctx, devEUI, geolocBuffer)
			return uplinkIDs, loc, err
		}
	}

	return nil, nil, nil
}

func (i *Integration) tdoaGeolocation(ctx context.Context, devEUI lorawan.EUI64, geolocBuffer [][]*gw.UplinkRXInfo) (*common.Location, error) {
	token := i.config.GeolocationToken
	migrated := false
	if i.config.DASToken != "" {
		token = i.config.DASToken
		migrated = true
	}

	client := geolocation.New(migrated, i.getGeolocationURI(), token)
	start := time.Now()

	var loc common.Location
	var err error

	if len(geolocBuffer) == 1 {
		// single-frame geoloc
		loc, err = client.TDOASingleFrame(ctx, geolocBuffer[0])
		loRaCloudAPIDuration("v2_tdoa_single").Observe(float64(time.Since(start)) / float64(time.Second))

	} else {
		// multi-frame geoloc
		loc, err = client.TDOAMultiFrame(ctx, geolocBuffer)
		loRaCloudAPIDuration("v2_tdoa_multi").Observe(float64(time.Since(start)) / float64(time.Second))
	}

	if err != nil {
		if err == geolocation.ErrNoLocation {
			return nil, nil
		}

		return nil, errors.Wrap(err, "geolocation error")
	}

	return &loc, nil
}

func (i *Integration) rssiGeolocation(ctx context.Context, devEUI lorawan.EUI64, geolocBuffer [][]*gw.UplinkRXInfo) (*common.Location, error) {
	token := i.config.GeolocationToken
	migrated := false
	if i.config.DASToken != "" {
		token = i.config.DASToken
		migrated = true
	}

	client := geolocation.New(migrated, i.getGeolocationURI(), token)
	start := time.Now()

	var loc common.Location
	var err error

	if len(geolocBuffer) == 1 {
		// single-frame geoloc
		loc, err = client.RSSISingleFrame(ctx, geolocBuffer[0])
		loRaCloudAPIDuration("v2_rssi_single").Observe(float64(time.Since(start)) / float64(time.Second))

	} else {
		// multi-frame geoloc
		loc, err = client.RSSIMultiFrame(ctx, geolocBuffer)
		loRaCloudAPIDuration("v2_rssi_multi").Observe(float64(time.Since(start)) / float64(time.Second))

	}

	if err != nil {
		if err == geolocation.ErrNoLocation {
			return nil, nil
		}

		return nil, errors.Wrap(err, "geolocation error")
	}

	return &loc, nil
}

func (i *Integration) gnssLR1110Geolocation(ctx context.Context, devEUI lorawan.EUI64, rxInfo []*gw.UplinkRXInfo, pl []byte) (*common.Location, error) {
	token := i.config.GeolocationToken
	migrated := false
	if i.config.DASToken != "" {
		token = i.config.DASToken
		migrated = true
	}

	client := geolocation.New(migrated, i.getGeolocationURI(), token)
	start := time.Now()

	loc, err := client.GNSSLR1110SingleFrame(ctx, rxInfo, i.config.GeolocationGNSSUseRxTime, pl)
	if err != nil {
		if err == geolocation.ErrNoLocation {
			return nil, nil
		}

		return nil, errors.Wrap(err, "geolocation error")
	}

	loRaCloudAPIDuration("v3_gnss_rl1110_single").Observe(float64(time.Since(start)) / float64(time.Second))

	return &loc, nil
}

func (i *Integration) wifiTDOAGeolocation(ctx context.Context, devEUI lorawan.EUI64, rxInfo []*gw.UplinkRXInfo, aps []geolocation.WifiAccessPoint) (*common.Location, error) {
	token := i.config.GeolocationToken
	migrated := false
	if i.config.DASToken != "" {
		token = i.config.DASToken
		migrated = true
	}

	client := geolocation.New(migrated, i.getGeolocationURI(), token)
	start := time.Now()

	loc, err := client.WifiTDOASingleFrame(ctx, rxInfo, aps)
	if err != nil {
		if err == geolocation.ErrNoLocation {
			return nil, nil
		}

		return nil, errors.Wrap(err, "geolocation error")
	}

	loRaCloudAPIDuration("v2_wifi_tdoa_single").Observe(float64(time.Since(start)) / float64(time.Second))

	return &loc, nil
}

func (i *Integration) dasJoin(ctx context.Context, devEUI lorawan.EUI64, pl pb.JoinEvent) error {
	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/das: forwarding join notification")

	client := das.New(i.getDasURI(), i.config.DASToken)
	start := time.Now()

	_, err := client.UplinkSend(ctx, das.UplinkRequest{
		helpers.EUI64(devEUI): das.UplinkMsgJoining{
			MsgType:   "joining",
			Timestamp: float64(helpers.GetTimestamp(pl.RxInfo).UnixNano()) / float64(time.Second),
			DR:        uint8(pl.Dr),
			Freq:      pl.GetTxInfo().Frequency,
		},
	})
	if err != nil {
		return errors.Wrap(err, "das error")
	}

	loRaCloudAPIDuration("das_v1_uplink_send").Observe(float64(time.Since(start)) / float64(time.Second))

	return nil
}

func (i *Integration) dasModem(ctx context.Context, vars map[string]string, devEUI lorawan.EUI64, pl pb.UplinkEvent, ii models.Integration) error {
	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/loracloud: forwarding das modem message")

	client := das.New(i.getDasURI(), i.config.DASToken)
	start := time.Now()

	resp, err := client.UplinkSend(ctx, das.UplinkRequest{
		helpers.EUI64(devEUI): das.UplinkMsgModem{
			MsgType:   "modem",
			Payload:   helpers.HEXBytes(pl.Data),
			FCnt:      pl.FCnt,
			Timestamp: float64(helpers.GetTimestamp(pl.RxInfo).UnixNano()) / float64(time.Second),
			DR:        uint8(pl.Dr),
			Freq:      pl.GetTxInfo().Frequency,
		},
	})
	if err != nil {
		return errors.Wrap(err, "das error")
	}

	loRaCloudAPIDuration("das_v1_uplink_send").Observe(float64(time.Since(start)) / float64(time.Second))

	err = i.handleDASResponse(ctx, vars, devEUI, resp, ii, pl, common.LocationSource_UNKNOWN)
	if err != nil {
		return errors.Wrap(err, "handle das response error")
	}

	return nil
}

func (i *Integration) dasGNSS(ctx context.Context, vars map[string]string, devEUI lorawan.EUI64, pl pb.UplinkEvent, ii models.Integration) error {
	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/loracloud: forwarding das gnss message")

	client := das.New(i.getDasURI(), i.config.DASToken)
	start := time.Now()

	msg := das.UplinkMsgGNSS{
		MsgType:   "gnss",
		Payload:   helpers.HEXBytes(pl.Data),
		Timestamp: float64(helpers.GetTimestamp(pl.RxInfo).UnixNano()) / float64(time.Second),
	}

	if i.config.DASGNSSUseRxTime {
		acc := float64(15)
		msg.GNSSCaptureTimeAccuracy = acc

		if gpsTime := helpers.GetTimeSinceGPSEpoch(pl.RxInfo); gpsTime != nil {
			t := (float64(*gpsTime) / float64(time.Second)) - 6
			msg.GNSSCaptureTime = t
		} else {
			gpsTime := gps.Time(time.Now()).TimeSinceGPSEpoch()
			t := (float64(gpsTime) / float64(time.Second)) - 6
			msg.GNSSCaptureTime = t
		}
	}

	if loc := helpers.GetStartLocation(pl.RxInfo); loc != nil {
		msg.GNSSAssistPosition = []float64{loc.Latitude, loc.Longitude}
		msg.GNSSAssistAltitude = loc.Altitude
	}

	resp, err := client.UplinkSend(ctx, das.UplinkRequest{
		helpers.EUI64(devEUI): msg,
	})
	if err != nil {
		return errors.Wrap(err, "das error")
	}

	loRaCloudAPIDuration("das_v1_uplink_send").Observe(float64(time.Since(start)) / float64(time.Second))

	err = i.handleDASResponse(ctx, vars, devEUI, resp, ii, pl, common.LocationSource_GEO_RESOLVER_GNSS)
	if err != nil {
		return errors.Wrap(err, "handle das response error")
	}

	return nil
}

func (i *Integration) dasUplinkMetaData(ctx context.Context, vars map[string]string, devEUI lorawan.EUI64, pl pb.UplinkEvent, ii models.Integration) error {
	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/das: forwarding uplink meta-data to das")

	client := das.New(i.getDasURI(), i.config.DASToken)
	start := time.Now()

	resp, err := client.UplinkSend(ctx, das.UplinkRequest{
		helpers.EUI64(devEUI): das.UplinkMsg{
			MsgType:   "updf",
			FCnt:      pl.FCnt,
			Timestamp: float64(helpers.GetTimestamp(pl.RxInfo).UnixNano()) / float64(time.Second),
			Port:      uint8(pl.FPort),
			DR:        uint8(pl.Dr),
			Freq:      pl.GetTxInfo().Frequency,
		},
	})
	if err != nil {
		return errors.Wrap(err, "das error")
	}

	err = i.handleDASResponse(ctx, vars, devEUI, resp, ii, pl, common.LocationSource_UNKNOWN)
	if err != nil {
		return errors.Wrap(err, "handle das response error")
	}

	loRaCloudAPIDuration("das_v1_uplink_send").Observe(float64(time.Since(start)) / float64(time.Second))

	return nil
}

func (i *Integration) handleDASResponse(ctx context.Context, vars map[string]string, devEUI lorawan.EUI64, dasResp das.UplinkResponse, ii models.Integration, pl integration.UplinkEvent, ls common.LocationSource) error {
	devResp, ok := dasResp.Result[helpers.EUI64(devEUI)]
	if !ok {
		return errors.New("no response for deveui")
	}

	if devResp.Error != "" {
		return fmt.Errorf("das api returned error: %s", devResp.Error)
	}

	// workaround!
	if i.config.DASStreamingGeolocWorkaround && len(devResp.Result.StreamRecords) != 0 {
		if err := i.streamGeolocWorkaround(ctx, vars, devEUI, ii, pl, devResp.Result.StreamRecords); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"dev_eui": devEUI,
				"ctx_id":  ctx.Value(logging.ContextIDKey),
			}).Error("integration/loracloud: streaming geoloc workaround error")
		}
	}

	b, err := json.Marshal(devResp.Result)
	if err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	intPL := integration.IntegrationEvent{
		ApplicationId:   pl.ApplicationId,
		ApplicationName: pl.ApplicationName,
		DeviceName:      pl.DeviceName,
		DevEui:          pl.DevEui,
		IntegrationName: "loracloud",
		EventType:       "DAS_UplinkResponse",
		ObjectJson:      string(b),
		PublishedAt:     ptypes.TimestampNow(),
	}
	if err := ii.HandleIntegrationEvent(ctx, nil, intPL); err != nil {
		log.WithError(err).Error("integration/loracloud: handle integration event error")
	}

	if dl := devResp.Result.Downlink; dl != nil {
		// Work-around for assistance position
		if dl.Port == 0 {
			dl.Port = 150
		}

		fCnt, err := storage.EnqueueDownlinkPayload(ctx, storage.DB(), devEUI, false, dl.Port, dl.Payload[:])
		if err != nil {
			log.WithError(err).Error("integration/loracloud: enqueue downlink payload error")
		} else {
			log.WithFields(log.Fields{
				"dev_eui": devEUI,
				"f_cnt":   fCnt,
				"ctx_id":  ctx.Value(logging.ContextIDKey),
			}).Info("integration/loracloud: enqueued das downlink payload")
		}
	}

	if ps := devResp.Result.PositionSolution; ps != nil {
		if len(ps.LLH) != 3 {
			return errors.New("position_solution.llh must contain exactly 3 items")
		}

		locPL := integration.LocationEvent{
			ApplicationId:   pl.ApplicationId,
			ApplicationName: pl.ApplicationName,
			DeviceName:      pl.DeviceName,
			DevEui:          pl.DevEui,
			FCnt:            pl.FCnt,
			Location: &common.Location{
				Latitude:  ps.LLH[0],
				Longitude: ps.LLH[1],
				Altitude:  ps.LLH[2],
				Source:    ls,
				Accuracy:  uint32(ps.Accuracy),
			},
			PublishedAt: ptypes.TimestampNow(),
		}
		if err := ii.HandleLocationEvent(ctx, vars, locPL); err != nil {
			log.WithError(err).Error("integration/loracloud: handle  location event error")
		}
	}

	return nil
}

func (i *Integration) streamGeolocWorkaround(ctx context.Context, vars map[string]string, devEUI lorawan.EUI64, ii models.Integration, pl integration.UplinkEvent, records das.StreamUpdate) error {
	var payloads [][]byte

	for _, r := range records {
		// sanity check as 0 = index, 1 = payload
		if len(r) != 2 {
			continue
		}

		// sanity check to make sure that pl is of type string
		hexPL, ok := r[1].(string)
		if !ok {
			continue
		}

		// decode pl from hex
		b, err := hex.DecodeString(hexPL)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"dev_eui": devEUI,
				"ctx_id":  ctx.Value(logging.ContextIDKey),
			}).Error("integration/loracloud: could not hex decode stream record")
			continue
		}

		payloads = append(payloads, b)
	}

	for _, p := range payloads {
		// there must be at least 2 bytes to read (tag and length)
		for index := 0; len(p)-index >= 2; {
			// get tag
			t := p[index]
			// get length
			l := int(p[index+1])

			// validate that we can at least read 'l' data
			if len(p)-index-2 < l {
				return errors.New("invalid TLV payload")
			}
			// get v
			v := p[index+2 : index+2+l]

			// increment index (2 bytes for t and l bytes + length of v)
			index = index + 2 + l

			if t == 0x06 || t == 0x07 {
				// gnss with pcb || gnss with patch antenna
				msg := das.UplinkMsgGNSS{
					MsgType:   "gnss",
					Payload:   helpers.HEXBytes(v),
					Timestamp: float64(helpers.GetTimestamp(pl.RxInfo).UnixNano()) / float64(time.Second),
				}

				// Note: we must rely on the embedded gnss timestamp, as the frame
				// is de-fragmented and we can not assume the scan time from the
				// rx timestamp.

				if loc := helpers.GetStartLocation(pl.RxInfo); loc != nil {
					msg.GNSSAssistPosition = []float64{loc.Latitude, loc.Longitude}
					msg.GNSSAssistAltitude = loc.Altitude
				}

				client := das.New(i.getDasURI(), i.config.DASToken)
				resp, err := client.UplinkSend(ctx, das.UplinkRequest{
					helpers.EUI64(devEUI): msg,
				})
				if err != nil {
					return errors.Wrap(err, "das error")
				}

				err = i.handleDASResponse(ctx, vars, devEUI, resp, ii, pl, common.LocationSource_GEO_RESOLVER_GNSS)
				if err != nil {
					return errors.Wrap(err, "handle das response error")
				}
			} else if t == 0x08 {
				// wifi (legacy)
				msg := das.UplinkMsgWifi{
					MsgType:   "wifi",
					Payload:   helpers.HEXBytes(append([]byte{0x01}, v...)),
					Timestamp: float64(helpers.GetTimestamp(pl.RxInfo).UnixNano()) / float64(time.Second),
				}

				client := das.New(i.getDasURI(), i.config.DASToken)
				resp, err := client.UplinkSend(ctx, das.UplinkRequest{
					helpers.EUI64(devEUI): msg,
				})
				if err != nil {
					return errors.Wrap(err, "das error")
				}

				err = i.handleDASResponse(ctx, vars, devEUI, resp, ii, pl, common.LocationSource_GEO_RESOLVER_WIFI)
				if err != nil {
					return errors.Wrap(err, "handle das response error")
				}
			} else if t == 0x0e {
				// we have to skip first 5 bytes
				if len(v) < 5 {
					continue
				}

				// wifi
				msg := das.UplinkMsgWifi{
					MsgType:   "wifi",
					Payload:   helpers.HEXBytes(append([]byte{0x01}, v[5:]...)),
					Timestamp: float64(helpers.GetTimestamp(pl.RxInfo).UnixNano()) / float64(time.Second),
				}

				client := das.New(i.getDasURI(), i.config.DASToken)
				resp, err := client.UplinkSend(ctx, das.UplinkRequest{
					helpers.EUI64(devEUI): msg,
				})
				if err != nil {
					return errors.Wrap(err, "das error")
				}

				err = i.handleDASResponse(ctx, vars, devEUI, resp, ii, pl, common.LocationSource_GEO_RESOLVER_WIFI)
				if err != nil {
					return errors.Wrap(err, "handle das response error")
				}
			}
		}
	}

	return nil
}

// filterOnFineTimestamp filters the given frame RXInfo slices on the presence
// of a plain fine-timestamp. Per frame it filters on the availability of
// minPerFrame.
func filterOnFineTimestamp(geolocBuffer [][]*gw.UplinkRXInfo, minPerFrame int) [][]*gw.UplinkRXInfo {
	var out [][]*gw.UplinkRXInfo

	for i := range geolocBuffer {
		var f []*gw.UplinkRXInfo

		for j := range geolocBuffer[i] {
			if geolocBuffer[i][j].GetFineTimestamp() != nil {
				f = append(f, geolocBuffer[i][j])
			}
		}

		if len(f) >= minPerFrame {
			out = append(out, f)
		}
	}

	return out
}

// getBytesFromJSONObject returns a slice of bytes from the decoded object,
// using the given name.
func getBytesFromJSONObject(field, jsonStr string) ([]byte, error) {
	if jsonStr == "" {
		return nil, nil
	}

	v := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonStr), &v); err != nil {
		return nil, errors.Wrap(err, "unmarshal json error")
	}

	vv, ok := v[field]
	if !ok {
		return nil, nil
	}

	str, ok := vv.(string)
	if !ok {
		return nil, fmt.Errorf("expected string, got %T", vv)
	}

	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, errors.Wrap(err, "base64 decode error")
	}

	return b, nil
}

// getWifiAccessPointsFromJSONObject returns a slice of Wifi APs from the
// decoded object, using the given name.
func getWifiAccessPointsFromJSONObject(field, jsonStr string) ([]geolocation.WifiAccessPoint, error) {
	if jsonStr == "" {
		return nil, nil
	}

	v := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonStr), &v); err != nil {
		return nil, errors.Wrap(err, "unmarshal json error")
	}

	vv, ok := v[field]
	if !ok {
		return nil, nil
	}

	aps, ok := vv.([]interface{})
	if !ok {
		return nil, fmt.Errorf("field content must be a list of objects, got: %T", vv)
	}

	var out []geolocation.WifiAccessPoint

	for i := range aps {
		vvv, ok := aps[i].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected key / value map, got: %T", aps[i])
		}

		var ap geolocation.WifiAccessPoint
		bssid, ok := vvv["macAddress"].(string)
		if !ok {
			return nil, fmt.Errorf("macAddress must be a string, got: %T", vvv["macAddress"])
		}
		b, err := base64.StdEncoding.DecodeString(bssid)
		if err != nil {
			return nil, errors.Wrap(err, "base64 decode error")
		}
		copy(ap.MacAddress[:], b)

		ss, ok := vvv["signalStrength"].(float64)
		if !ok {
			return nil, fmt.Errorf("signalStrength must be a float64, got: %T", vvv["signalStrength"])
		}
		ap.SignalStrength = int(ss)
		out = append(out, ap)
	}

	return out, nil
}
