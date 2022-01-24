package loracloud

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	nsmock "github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/integration/loracloud/client/das"
	"github.com/brocaar/chirpstack-application-server/internal/integration/loracloud/client/geolocation"
	"github.com/brocaar/chirpstack-application-server/internal/integration/loracloud/client/helpers"
	"github.com/brocaar/chirpstack-application-server/internal/integration/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
)

type LoRaCloudTestSuite struct {
	suite.Suite

	apiResponse string
	apiRequest  string
	server      *httptest.Server
	integration *mock.Integration
	loraCloud   *Integration

	nsClient *nsmock.Client
	device   storage.Device
}

func (ts *LoRaCloudTestSuite) SetupSuite() {
	assert := require.New(ts.T())
	conf := test.GetConfig()

	assert.NoError(storage.Setup(conf))
	assert.NoError(storage.MigrateDown(storage.DB().DB))
	assert.NoError(storage.MigrateUp(storage.DB().DB))

	ts.server = httptest.NewServer(http.HandlerFunc(ts.apiHandler))
	ts.integration = mock.New()
	ts.loraCloud, _ = New(Config{})
	ts.loraCloud.geolocationURI = ts.server.URL
	ts.loraCloud.dasURI = ts.server.URL

	ts.nsClient = nsmock.NewClient()
	networkserver.SetPool(nsmock.NewPool(ts.nsClient))

	ns := storage.NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:123",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &ns))

	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))

	sp := storage.ServiceProfile{
		Name:            "test-sp",
		NetworkServerID: ns.ID,
		OrganizationID:  org.ID,
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp))
	var spID uuid.UUID
	copy(spID[:], sp.ServiceProfile.Id)

	dp := storage.DeviceProfile{
		Name:            "test-dp",
		NetworkServerID: ns.ID,
		OrganizationID:  org.ID,
	}
	assert.NoError(storage.CreateDeviceProfile(context.Background(), storage.DB(), &dp))
	var dpID uuid.UUID
	copy(dpID[:], dp.DeviceProfile.Id)

	app := storage.Application{
		Name:             "test-app",
		OrganizationID:   org.ID,
		ServiceProfileID: spID,
	}
	assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &app))

	ts.device = storage.Device{
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		ApplicationID:   app.ID,
		DeviceProfileID: dpID,
		Name:            "test-dev",
	}
	assert.NoError(storage.CreateDevice(context.Background(), storage.DB(), &ts.device))
}

func (ts *LoRaCloudTestSuite) TearDownSuite() {
	ts.server.Close()
}

func (ts *LoRaCloudTestSuite) TestHandleUplinkEvent() {
	nowPB := ptypes.TimestampNow()
	altitude := float64(3.333)

	ts.T().Run("Geolocation", func(t *testing.T) {
		tests := []struct {
			name                string
			config              Config
			geolocBuffer        [][]*gw.UplinkRXInfo
			uplinkEvent         pb.UplinkEvent
			geolocationResponse interface{}

			expectedGeolocationRequest interface{}
			expectedLocationEvent      *pb.LocationEvent
		}{
			{
				name: "geolocation disabled",
				config: Config{
					Geolocation: false,
				},
				uplinkEvent: pb.UplinkEvent{
					ApplicationId:   1,
					ApplicationName: "test-app",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceName:      "test-device",
					RxInfo: []*gw.UplinkRXInfo{
						{
							UplinkId:  []byte{1},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 111,
									},
								},
							},
						},
						{
							UplinkId:  []byte{2},
							GatewayId: []byte{2, 2, 2, 2, 2, 2, 2, 2},
							Time:      nowPB,
							Rssi:      2,
							LoraSnr:   2.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 222,
									},
								},
							},
						},
						{
							UplinkId:  []byte{3},
							GatewayId: []byte{3, 3, 3, 3, 3, 3, 3, 3},
							Time:      nowPB,
							Rssi:      3,
							LoraSnr:   3.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 333,
									},
								},
							},
						},
					},
				},
			},
			{
				name: "geolocation enabled, single TDOA",
				config: Config{
					Geolocation:     true,
					GeolocationTDOA: true,
					GeolocationRSSI: false,
				},
				geolocationResponse: &geolocation.Response{
					Result: &geolocation.LocationResult{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Accuracy:  10,
					},
				},
				expectedGeolocationRequest: &geolocation.TDOASingleFrameRequest{
					LoRaWAN: []geolocation.UplinkTDOA{
						{
							GatewayID: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
							RSSI:      1,
							SNR:       1.1,
							TOA:       111,
							AntennaLocation: geolocation.AntennaLocation{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
						{
							GatewayID: lorawan.EUI64{2, 2, 2, 2, 2, 2, 2, 2},
							RSSI:      2,
							SNR:       2.1,
							TOA:       222,
							AntennaLocation: geolocation.AntennaLocation{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
						{
							GatewayID: lorawan.EUI64{3, 3, 3, 3, 3, 3, 3, 3},
							RSSI:      3,
							SNR:       3.1,
							TOA:       333,
							AntennaLocation: geolocation.AntennaLocation{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
					},
				},
				expectedLocationEvent: &pb.LocationEvent{
					ApplicationName: "test-app",
					ApplicationId:   1,
					DeviceName:      "test-device",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					UplinkIds:       [][]byte{{1}, {2}, {3}},
					Location: &common.Location{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Source:    common.LocationSource_GEO_RESOLVER_TDOA,
						Accuracy:  10,
					},
				},
				uplinkEvent: pb.UplinkEvent{
					ApplicationId:   1,
					ApplicationName: "test-app",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceName:      "test-device",
					RxInfo: []*gw.UplinkRXInfo{
						{
							UplinkId:  []byte{1},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 111,
									},
								},
							},
						},
						{
							UplinkId:  []byte{2},
							GatewayId: []byte{2, 2, 2, 2, 2, 2, 2, 2},
							Time:      nowPB,
							Rssi:      2,
							LoraSnr:   2.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 222,
									},
								},
							},
						},
						{
							UplinkId:  []byte{3},
							GatewayId: []byte{3, 3, 3, 3, 3, 3, 3, 3},
							Time:      nowPB,
							Rssi:      3,
							LoraSnr:   3.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 333,
									},
								},
							},
						},
					},
				},
			},
			{
				name: "geolocation enabled, single RSSI",
				config: Config{
					Geolocation:     true,
					GeolocationTDOA: false,
					GeolocationRSSI: true,
				},
				geolocationResponse: &geolocation.Response{
					Result: &geolocation.LocationResult{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Accuracy:  10,
					},
				},
				expectedGeolocationRequest: &geolocation.RSSISingleFrameRequest{
					LoRaWAN: []geolocation.UplinkRSSI{
						{
							GatewayID: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
							RSSI:      1,
							SNR:       1.1,
							AntennaLocation: geolocation.AntennaLocation{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
						{
							GatewayID: lorawan.EUI64{2, 2, 2, 2, 2, 2, 2, 2},
							RSSI:      2,
							SNR:       2.1,
							AntennaLocation: geolocation.AntennaLocation{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
						{
							GatewayID: lorawan.EUI64{3, 3, 3, 3, 3, 3, 3, 3},
							RSSI:      3,
							SNR:       3.1,
							AntennaLocation: geolocation.AntennaLocation{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
					},
				},
				expectedLocationEvent: &pb.LocationEvent{
					ApplicationName: "test-app",
					ApplicationId:   1,
					DeviceName:      "test-device",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					UplinkIds:       [][]byte{{1}, {2}, {3}},
					Location: &common.Location{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Source:    common.LocationSource_GEO_RESOLVER_RSSI,
						Accuracy:  10,
					},
				},
				uplinkEvent: pb.UplinkEvent{
					ApplicationId:   1,
					ApplicationName: "test-app",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceName:      "test-device",
					RxInfo: []*gw.UplinkRXInfo{
						{
							UplinkId:  []byte{1},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 111,
									},
								},
							},
						},
						{
							UplinkId:  []byte{2},
							GatewayId: []byte{2, 2, 2, 2, 2, 2, 2, 2},
							Time:      nowPB,
							Rssi:      2,
							LoraSnr:   2.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 222,
									},
								},
							},
						},
						{
							UplinkId:  []byte{3},
							GatewayId: []byte{3, 3, 3, 3, 3, 3, 3, 3},
							Time:      nowPB,
							Rssi:      3,
							LoraSnr:   3.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 333,
									},
								},
							},
						},
					},
				},
			},
			{
				name: "geolocation enabled, fallback to RSSI (only two fine-timestamps)",
				config: Config{
					Geolocation:     true,
					GeolocationTDOA: true,
					GeolocationRSSI: true,
				},
				geolocationResponse: &geolocation.Response{
					Result: &geolocation.LocationResult{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Accuracy:  10,
					},
				},
				expectedGeolocationRequest: &geolocation.RSSISingleFrameRequest{
					LoRaWAN: []geolocation.UplinkRSSI{
						{
							GatewayID: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
							RSSI:      1,
							SNR:       1.1,
							AntennaLocation: geolocation.AntennaLocation{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
						{
							GatewayID: lorawan.EUI64{2, 2, 2, 2, 2, 2, 2, 2},
							RSSI:      2,
							SNR:       2.1,
							AntennaLocation: geolocation.AntennaLocation{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
						{
							GatewayID: lorawan.EUI64{3, 3, 3, 3, 3, 3, 3, 3},
							RSSI:      3,
							SNR:       3.1,
							AntennaLocation: geolocation.AntennaLocation{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
					},
				},
				expectedLocationEvent: &pb.LocationEvent{
					ApplicationName: "test-app",
					ApplicationId:   1,
					DeviceName:      "test-device",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					UplinkIds:       [][]byte{{1}, {2}, {3}},
					Location: &common.Location{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Source:    common.LocationSource_GEO_RESOLVER_RSSI,
						Accuracy:  10,
					},
				},
				uplinkEvent: pb.UplinkEvent{
					ApplicationId:   1,
					ApplicationName: "test-app",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceName:      "test-device",
					RxInfo: []*gw.UplinkRXInfo{
						{
							UplinkId:  []byte{1},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
						{
							UplinkId:  []byte{2},
							GatewayId: []byte{2, 2, 2, 2, 2, 2, 2, 2},
							Time:      nowPB,
							Rssi:      2,
							LoraSnr:   2.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 222,
									},
								},
							},
						},
						{
							UplinkId:  []byte{3},
							GatewayId: []byte{3, 3, 3, 3, 3, 3, 3, 3},
							Time:      nowPB,
							Rssi:      3,
							LoraSnr:   3.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 333,
									},
								},
							},
						},
					},
				},
			},
			{
				name: "geoloc buffer too small",
				config: Config{
					Geolocation:              true,
					GeolocationTDOA:          true,
					GeolocationMinBufferSize: 2,
					GeolocationBufferTTL:     60,
				},
				uplinkEvent: pb.UplinkEvent{
					ApplicationId:   1,
					ApplicationName: "test-app",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceName:      "test-device",
					RxInfo: []*gw.UplinkRXInfo{
						{
							UplinkId:  []byte{1},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 111,
									},
								},
							},
						},
						{
							UplinkId:  []byte{2},
							GatewayId: []byte{2, 2, 2, 2, 2, 2, 2, 2},
							Time:      nowPB,
							Rssi:      2,
							LoraSnr:   2.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 222,
									},
								},
							},
						},
						{
							UplinkId:  []byte{3},
							GatewayId: []byte{3, 3, 3, 3, 3, 3, 3, 3},
							Time:      nowPB,
							Rssi:      3,
							LoraSnr:   3.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 333,
									},
								},
							},
						},
					},
				},
			},
			{
				name: "geolocation with buffer - TDOA",
				config: Config{
					Geolocation:              true,
					GeolocationTDOA:          true,
					GeolocationRSSI:          false,
					GeolocationMinBufferSize: 2,
					GeolocationBufferTTL:     60,
				},
				geolocationResponse: &geolocation.Response{
					Result: &geolocation.LocationResult{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Accuracy:  10,
					},
				},
				expectedGeolocationRequest: &geolocation.TDOAMultiFrameRequest{
					LoRaWAN: [][]geolocation.UplinkTDOA{
						{
							{
								GatewayID: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
								RSSI:      1,
								SNR:       1.1,
								TOA:       444,
								AntennaLocation: geolocation.AntennaLocation{
									Latitude:  1.111,
									Longitude: 2.222,
									Altitude:  3.333,
								},
							},
							{
								GatewayID: lorawan.EUI64{2, 2, 2, 2, 2, 2, 2, 2},
								RSSI:      2,
								SNR:       2.1,
								TOA:       555,
								AntennaLocation: geolocation.AntennaLocation{
									Latitude:  1.111,
									Longitude: 2.222,
									Altitude:  3.333,
								},
							},
							{
								GatewayID: lorawan.EUI64{3, 3, 3, 3, 3, 3, 3, 3},
								RSSI:      3,
								SNR:       3.1,
								TOA:       666,
								AntennaLocation: geolocation.AntennaLocation{
									Latitude:  1.111,
									Longitude: 2.222,
									Altitude:  3.333,
								},
							},
						},
						{
							{
								GatewayID: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
								RSSI:      1,
								SNR:       1.1,
								TOA:       111,
								AntennaLocation: geolocation.AntennaLocation{
									Latitude:  1.111,
									Longitude: 2.222,
									Altitude:  3.333,
								},
							},
							{
								GatewayID: lorawan.EUI64{2, 2, 2, 2, 2, 2, 2, 2},
								RSSI:      2,
								SNR:       2.1,
								TOA:       222,
								AntennaLocation: geolocation.AntennaLocation{
									Latitude:  1.111,
									Longitude: 2.222,
									Altitude:  3.333,
								},
							},
							{
								GatewayID: lorawan.EUI64{3, 3, 3, 3, 3, 3, 3, 3},
								RSSI:      3,
								SNR:       3.1,
								TOA:       333,
								AntennaLocation: geolocation.AntennaLocation{
									Latitude:  1.111,
									Longitude: 2.222,
									Altitude:  3.333,
								},
							},
						},
					},
				},
				expectedLocationEvent: &pb.LocationEvent{
					ApplicationName: "test-app",
					ApplicationId:   1,
					DeviceName:      "test-device",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					UplinkIds:       [][]byte{{4}, {5}, {6}, {1}, {2}, {3}},
					Location: &common.Location{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Source:    common.LocationSource_GEO_RESOLVER_TDOA,
						Accuracy:  10,
					},
				},
				geolocBuffer: [][]*gw.UplinkRXInfo{
					{
						{
							UplinkId:  []byte{4},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 444,
									},
								},
							},
						},
						{
							UplinkId:  []byte{5},
							GatewayId: []byte{2, 2, 2, 2, 2, 2, 2, 2},
							Time:      nowPB,
							Rssi:      2,
							LoraSnr:   2.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 555,
									},
								},
							},
						},
						{
							UplinkId:  []byte{6},
							GatewayId: []byte{3, 3, 3, 3, 3, 3, 3, 3},
							Time:      nowPB,
							Rssi:      3,
							LoraSnr:   3.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 666,
									},
								},
							},
						},
					},
				},
				uplinkEvent: pb.UplinkEvent{
					ApplicationId:   1,
					ApplicationName: "test-app",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceName:      "test-device",
					RxInfo: []*gw.UplinkRXInfo{
						{
							UplinkId:  []byte{1},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 111,
									},
								},
							},
						},
						{
							UplinkId:  []byte{2},
							GatewayId: []byte{2, 2, 2, 2, 2, 2, 2, 2},
							Time:      nowPB,
							Rssi:      2,
							LoraSnr:   2.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 222,
									},
								},
							},
						},
						{
							UplinkId:  []byte{3},
							GatewayId: []byte{3, 3, 3, 3, 3, 3, 3, 3},
							Time:      nowPB,
							Rssi:      3,
							LoraSnr:   3.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 333,
									},
								},
							},
						},
					},
				},
			},
			{
				name: "geolocation with buffer - RSSI",
				config: Config{
					Geolocation:              true,
					GeolocationTDOA:          false,
					GeolocationRSSI:          true,
					GeolocationMinBufferSize: 2,
					GeolocationBufferTTL:     60,
				},
				geolocationResponse: &geolocation.Response{
					Result: &geolocation.LocationResult{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Accuracy:  10,
					},
				},
				expectedGeolocationRequest: &geolocation.RSSIMultiFrameRequest{
					LoRaWAN: [][]geolocation.UplinkRSSI{
						{
							{
								GatewayID: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
								RSSI:      1,
								SNR:       1.1,
								AntennaLocation: geolocation.AntennaLocation{
									Latitude:  1.111,
									Longitude: 2.222,
									Altitude:  3.333,
								},
							},
							{
								GatewayID: lorawan.EUI64{2, 2, 2, 2, 2, 2, 2, 2},
								RSSI:      2,
								SNR:       2.1,
								AntennaLocation: geolocation.AntennaLocation{
									Latitude:  1.111,
									Longitude: 2.222,
									Altitude:  3.333,
								},
							},
							{
								GatewayID: lorawan.EUI64{3, 3, 3, 3, 3, 3, 3, 3},
								RSSI:      3,
								SNR:       3.1,
								AntennaLocation: geolocation.AntennaLocation{
									Latitude:  1.111,
									Longitude: 2.222,
									Altitude:  3.333,
								},
							},
						},
						{
							{
								GatewayID: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
								RSSI:      1,
								SNR:       1.1,
								AntennaLocation: geolocation.AntennaLocation{
									Latitude:  1.111,
									Longitude: 2.222,
									Altitude:  3.333,
								},
							},
							{
								GatewayID: lorawan.EUI64{2, 2, 2, 2, 2, 2, 2, 2},
								RSSI:      2,
								SNR:       2.1,
								AntennaLocation: geolocation.AntennaLocation{
									Latitude:  1.111,
									Longitude: 2.222,
									Altitude:  3.333,
								},
							},
							{
								GatewayID: lorawan.EUI64{3, 3, 3, 3, 3, 3, 3, 3},
								RSSI:      3,
								SNR:       3.1,
								AntennaLocation: geolocation.AntennaLocation{
									Latitude:  1.111,
									Longitude: 2.222,
									Altitude:  3.333,
								},
							},
						},
					},
				},
				expectedLocationEvent: &pb.LocationEvent{
					ApplicationName: "test-app",
					ApplicationId:   1,
					DeviceName:      "test-device",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					UplinkIds:       [][]byte{{4}, {5}, {6}, {1}, {2}, {3}},
					Location: &common.Location{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Source:    common.LocationSource_GEO_RESOLVER_RSSI,
						Accuracy:  10,
					},
				},
				geolocBuffer: [][]*gw.UplinkRXInfo{
					{
						{
							UplinkId:  []byte{4},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 444,
									},
								},
							},
						},
						{
							UplinkId:  []byte{5},
							GatewayId: []byte{2, 2, 2, 2, 2, 2, 2, 2},
							Time:      nowPB,
							Rssi:      2,
							LoraSnr:   2.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 555,
									},
								},
							},
						},
						{
							UplinkId:  []byte{6},
							GatewayId: []byte{3, 3, 3, 3, 3, 3, 3, 3},
							Time:      nowPB,
							Rssi:      3,
							LoraSnr:   3.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 666,
									},
								},
							},
						},
					},
				},
				uplinkEvent: pb.UplinkEvent{
					ApplicationId:   1,
					ApplicationName: "test-app",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceName:      "test-device",
					RxInfo: []*gw.UplinkRXInfo{
						{
							UplinkId:  []byte{1},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 111,
									},
								},
							},
						},
						{
							UplinkId:  []byte{2},
							GatewayId: []byte{2, 2, 2, 2, 2, 2, 2, 2},
							Time:      nowPB,
							Rssi:      2,
							LoraSnr:   2.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 222,
									},
								},
							},
						},
						{
							UplinkId:  []byte{3},
							GatewayId: []byte{3, 3, 3, 3, 3, 3, 3, 3},
							Time:      nowPB,
							Rssi:      3,
							LoraSnr:   3.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 333,
									},
								},
							},
						},
					},
				},
			},
			{
				name: "gnss geolocation",
				config: Config{
					Geolocation:                 true,
					GeolocationGNSS:             true,
					GeolocationGNSSPayloadField: "lr1110_gnss",
				},
				geolocationResponse: &geolocation.V3Response{
					Result: &geolocation.LocationSolverResult{
						LLH:      []float64{1.123, 2.123, 3.123},
						Accuracy: 10,
					},
				},
				expectedGeolocationRequest: &geolocation.GNSSLR1110SingleFrameRequest{
					Payload:            helpers.HEXBytes([]byte{1, 2, 3}),
					GNSSAssistPosition: []float64{1.111, 2.222},
					GNSSAssistAltitude: &altitude,
				},
				expectedLocationEvent: &pb.LocationEvent{
					ApplicationName: "test-app",
					ApplicationId:   1,
					DeviceName:      "test-device",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					Location: &common.Location{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Source:    common.LocationSource_GEO_RESOLVER_GNSS,
						Accuracy:  10,
					},
				},
				uplinkEvent: pb.UplinkEvent{
					ApplicationId:   1,
					ApplicationName: "test-app",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceName:      "test-device",
					ObjectJson:      `{"lr1110_gnss": "AQID"}`,
					RxInfo: []*gw.UplinkRXInfo{
						{
							UplinkId:  []byte{1},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 111,
									},
								},
							},
						},
					},
				},
			},
			{
				name: "gnss geolocation, no payload",
				config: Config{
					Geolocation:                 true,
					GeolocationGNSS:             true,
					GeolocationGNSSPayloadField: "lr1110_gnss",
				},
				uplinkEvent: pb.UplinkEvent{
					ApplicationId:   1,
					ApplicationName: "test-app",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceName:      "test-device",
					ObjectJson:      `{"different_field": "AQID"}`,
					RxInfo: []*gw.UplinkRXInfo{
						{
							UplinkId:  []byte{1},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: &timestamp.Timestamp{
										Nanos: 111,
									},
								},
							},
						},
					},
				},
			},
			{
				name: "wifi geolocation",
				config: Config{
					Geolocation:                 true,
					GeolocationWifi:             true,
					GeolocationWifiPayloadField: "wifi_aps",
				},
				geolocationResponse: &geolocation.Response{
					Result: &geolocation.LocationResult{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Accuracy:  10,
					},
				},
				expectedGeolocationRequest: &geolocation.WifiTDOASingleFrameRequest{
					LoRaWAN: []geolocation.UplinkTDOA{
						{
							GatewayID: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
							RSSI:      1,
							SNR:       1.1,
							AntennaLocation: geolocation.AntennaLocation{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
					},
					WifiAccessPoints: []geolocation.WifiAccessPoint{
						{
							MacAddress:     geolocation.BSSID{1, 1, 1, 1, 1, 1},
							SignalStrength: -10,
						},
						{
							MacAddress:     geolocation.BSSID{2, 2, 2, 2, 2, 2},
							SignalStrength: -20,
						},
						{
							MacAddress:     geolocation.BSSID{3, 3, 3, 3, 3, 3},
							SignalStrength: -30,
						},
					},
				},
				expectedLocationEvent: &pb.LocationEvent{
					ApplicationName: "test-app",
					ApplicationId:   1,
					DeviceName:      "test-device",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					Location: &common.Location{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
						Source:    common.LocationSource_GEO_RESOLVER_WIFI,
						Accuracy:  10,
					},
				},
				uplinkEvent: pb.UplinkEvent{
					ApplicationId:   1,
					ApplicationName: "test-app",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceName:      "test-device",
					ObjectJson: `{
						"wifi_aps": [
							{"macAddress": "AQEBAQEB", "signalStrength": -10},
							{"macAddress": "AgICAgIC", "signalStrength": -20},
							{"macAddress": "AwMDAwMD", "signalStrength": -30}
						]
					}`,
					RxInfo: []*gw.UplinkRXInfo{
						{
							UplinkId:  []byte{1},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
					},
				},
			},
			{
				name: "wifi geolocation, no payload",
				config: Config{
					Geolocation:                 true,
					GeolocationWifi:             true,
					GeolocationWifiPayloadField: "wifi_aps",
				},
				uplinkEvent: pb.UplinkEvent{
					ApplicationId:   1,
					ApplicationName: "test-app",
					DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceName:      "test-device",
					RxInfo: []*gw.UplinkRXInfo{
						{
							UplinkId:  []byte{1},
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Time:      nowPB,
							Rssi:      1,
							LoraSnr:   1.1,
							Location: &common.Location{
								Latitude:  1.111,
								Longitude: 2.222,
								Altitude:  3.333,
							},
						},
					},
				},
			},
		}

		for _, tst := range tests {
			t.Run(tst.name, func(t *testing.T) {
				assert := require.New(t)
				storage.RedisClient().FlushAll(context.Background())

				var devEUI lorawan.EUI64
				copy(devEUI[:], tst.uplinkEvent.DevEui)

				// set integration config
				ts.loraCloud.config = tst.config
				ts.apiRequest = ""

				// set geloc buffer
				assert.NoError(SaveGeolocBuffer(context.Background(), devEUI, tst.geolocBuffer, time.Duration(tst.config.GeolocationBufferTTL)*time.Second))

				// set api response
				if tst.geolocationResponse != nil {
					b, err := json.Marshal(tst.geolocationResponse)
					assert.NoError(err)
					ts.apiResponse = string(b)
				} else {
					ts.apiResponse = ""
				}

				// call LoRaCloud method
				assert.NoError(ts.loraCloud.HandleUplinkEvent(context.Background(), ts.integration, nil, tst.uplinkEvent))

				// assert request
				if tst.expectedGeolocationRequest != nil {
					b, err := json.Marshal(tst.expectedGeolocationRequest)
					assert.NoError(err)
					assert.Equal(string(b), ts.apiRequest)
				} else {
					assert.Equal("", ts.apiRequest)
				}

				// assert locationEvent
				if tst.expectedLocationEvent != nil {
					pl := <-ts.integration.SendLocationNotificationChan
					assert.NotNil(pl.PublishedAt)
					pl.PublishedAt = nil
					assert.Equal(*tst.expectedLocationEvent, pl)
				} else {
					assert.Len(ts.integration.SendLocationNotificationChan, 0)
				}
			})
		}
	})

	ts.T().Run("DAS uplink", func(t *testing.T) {
		assert := require.New(t)
		nowPB := ptypes.TimestampNow()
		now, err := ptypes.Timestamp(nowPB)
		assert.NoError(err)

		integrationB, err := json.Marshal(das.UplinkResponseResult{})
		assert.NoError(err)

		integrationLocB, err := json.Marshal(das.UplinkResponseResult{
			PositionSolution: &das.PositionSolution{
				LLH:      []float64{1.1, 2.2, 3.3},
				Accuracy: 10,
			},
		})

		integrationDownlinkB, err := json.Marshal(das.UplinkResponseResult{
			Downlink: &das.LoRaDownlink{
				Port:    10,
				Payload: helpers.HEXBytes{4, 5, 6},
			},
		})
		assert.NoError(err)

		tests := []struct {
			name        string
			config      Config
			uplinkEvent pb.UplinkEvent
			dasResponse *das.UplinkResponse

			expectedDASRequest       *das.UplinkRequest
			expectedDownlinkPayload  []byte
			expectedDownlinkFPort    uint8
			expectedIntegrationEvent *pb.IntegrationEvent
			expectedLocationEvent    *pb.LocationEvent
		}{
			{
				name: "modem uplink",
				config: Config{
					DAS:          true,
					DASModemPort: 199,
				},
				uplinkEvent: pb.UplinkEvent{
					DevEui: ts.device.DevEUI[:],
					RxInfo: []*gw.UplinkRXInfo{
						{
							Time: nowPB,
						},
					},
					TxInfo: &gw.UplinkTXInfo{
						Frequency: 868100000,
					},
					Dr:    3,
					FCnt:  10,
					FPort: 199,
					Data:  []byte{1, 2, 3},
				},
				dasResponse: &das.UplinkResponse{
					Result: das.UplinkDeviceMapResponse{
						helpers.EUI64(ts.device.DevEUI): das.UplinkResponseItem{
							Result: das.UplinkResponseResult{},
						},
					},
				},
				expectedDASRequest: &das.UplinkRequest{
					helpers.EUI64(ts.device.DevEUI): &das.UplinkMsgModem{
						MsgType:   "modem",
						Payload:   helpers.HEXBytes{1, 2, 3},
						FCnt:      10,
						Timestamp: float64(now.UnixNano()) / float64(time.Second),
						DR:        3,
						Freq:      868100000,
					},
				},
				expectedIntegrationEvent: &pb.IntegrationEvent{
					DevEui:          ts.device.DevEUI[:],
					IntegrationName: "loracloud",
					EventType:       "DAS_UplinkResponse",
					ObjectJson:      string(integrationB),
				},
			},
			{
				name: "modem uplink with downlink",
				config: Config{
					DAS:          true,
					DASModemPort: 199,
				},
				uplinkEvent: pb.UplinkEvent{
					DevEui: ts.device.DevEUI[:],
					RxInfo: []*gw.UplinkRXInfo{
						{
							Time: nowPB,
						},
					},
					TxInfo: &gw.UplinkTXInfo{
						Frequency: 868100000,
					},
					Dr:    3,
					FCnt:  10,
					FPort: 199,
					Data:  []byte{1, 2, 3},
				},
				dasResponse: &das.UplinkResponse{
					Result: das.UplinkDeviceMapResponse{
						helpers.EUI64(ts.device.DevEUI): das.UplinkResponseItem{
							Result: das.UplinkResponseResult{
								Downlink: &das.LoRaDownlink{
									Port:    10,
									Payload: helpers.HEXBytes{4, 5, 6},
								},
							},
						},
					},
				},
				expectedDASRequest: &das.UplinkRequest{
					helpers.EUI64(ts.device.DevEUI): &das.UplinkMsgModem{
						MsgType:   "modem",
						Payload:   helpers.HEXBytes{1, 2, 3},
						FCnt:      10,
						Timestamp: float64(now.UnixNano()) / float64(time.Second),
						DR:        3,
						Freq:      868100000,
					},
				},
				expectedDownlinkFPort:   10,
				expectedDownlinkPayload: []byte{4, 5, 6},
				expectedIntegrationEvent: &pb.IntegrationEvent{
					DevEui:          ts.device.DevEUI[:],
					IntegrationName: "loracloud",
					EventType:       "DAS_UplinkResponse",
					ObjectJson:      string(integrationDownlinkB),
				},
			},
			{
				name: "uplink meta-data",
				config: Config{
					DAS:          true,
					DASModemPort: 199,
				},
				uplinkEvent: pb.UplinkEvent{
					DevEui: ts.device.DevEUI[:],
					RxInfo: []*gw.UplinkRXInfo{
						{
							Time: nowPB,
						},
					},
					TxInfo: &gw.UplinkTXInfo{
						Frequency: 868100000,
					},
					Dr:    3,
					FCnt:  10,
					FPort: 190,
					Data:  []byte{1, 2, 3},
				},
				dasResponse: &das.UplinkResponse{
					Result: das.UplinkDeviceMapResponse{
						helpers.EUI64(ts.device.DevEUI): das.UplinkResponseItem{
							Result: das.UplinkResponseResult{},
						},
					},
				},
				expectedDASRequest: &das.UplinkRequest{
					helpers.EUI64(ts.device.DevEUI): &das.UplinkMsg{
						MsgType:   "updf",
						FCnt:      10,
						Timestamp: float64(now.UnixNano()) / float64(time.Second),
						DR:        3,
						Freq:      868100000,
						Port:      190,
					},
				},
				expectedIntegrationEvent: &pb.IntegrationEvent{
					DevEui:          ts.device.DevEUI[:],
					IntegrationName: "loracloud",
					EventType:       "DAS_UplinkResponse",
					ObjectJson:      string(integrationB),
				},
			},
			{
				name: "uplink gnss",
				config: Config{
					DAS:         true,
					DASGNSSPort: 198,
				},
				uplinkEvent: pb.UplinkEvent{
					DevEui: ts.device.DevEUI[:],
					RxInfo: []*gw.UplinkRXInfo{
						{
							Time: nowPB,
						},
					},
					TxInfo: &gw.UplinkTXInfo{
						Frequency: 868100000,
					},
					Dr:    3,
					FCnt:  10,
					FPort: 198,
					Data:  []byte{1, 2, 3},
				},
				dasResponse: &das.UplinkResponse{
					Result: das.UplinkDeviceMapResponse{
						helpers.EUI64(ts.device.DevEUI): das.UplinkResponseItem{
							Result: das.UplinkResponseResult{
								PositionSolution: &das.PositionSolution{
									LLH:      []float64{1.1, 2.2, 3.3},
									Accuracy: 10,
								},
							},
						},
					},
				},
				expectedDASRequest: &das.UplinkRequest{
					helpers.EUI64(ts.device.DevEUI): &das.UplinkMsgGNSS{
						MsgType:   "gnss",
						Payload:   helpers.HEXBytes{1, 2, 3},
						Timestamp: float64(now.UnixNano()) / float64(time.Second),
					},
				},
				expectedIntegrationEvent: &pb.IntegrationEvent{
					DevEui:          ts.device.DevEUI[:],
					IntegrationName: "loracloud",
					EventType:       "DAS_UplinkResponse",
					ObjectJson:      string(integrationLocB),
				},
				expectedLocationEvent: &pb.LocationEvent{
					DevEui: ts.device.DevEUI[:],
					Location: &common.Location{
						Latitude:  1.1,
						Longitude: 2.2,
						Altitude:  3.3,
						Source:    common.LocationSource_GEO_RESOLVER_GNSS,
						Accuracy:  10,
					},
					FCnt: 10,
				},
			},
		}

		for _, tst := range tests {
			t.Run(tst.name, func(t *testing.T) {
				assert := require.New(t)

				// set integration config
				ts.loraCloud.config = tst.config
				ts.apiRequest = ""

				// set api response
				if tst.dasResponse != nil {
					b, err := json.Marshal(tst.dasResponse)
					assert.NoError(err)
					ts.apiResponse = string(b)
				} else {
					ts.apiResponse = ""
				}

				// call LoRa Cloud method
				assert.NoError(ts.loraCloud.HandleUplinkEvent(context.Background(), ts.integration, nil, tst.uplinkEvent))

				// assert request
				if tst.expectedDASRequest != nil {
					b, err := json.Marshal(tst.expectedDASRequest)
					assert.NoError(err)
					assert.Equal(string(b), ts.apiRequest)
				} else {
					assert.Equal("", ts.apiRequest)
				}

				// assert downlink
				if len(tst.expectedDownlinkPayload) != 0 {
					downPL := <-ts.nsClient.CreateDeviceQueueItemChan
					assert.EqualValues(0, downPL.Item.FCnt)
					assert.EqualValues(tst.expectedDownlinkFPort, downPL.Item.FPort)

					b, err := lorawan.EncryptFRMPayload(ts.device.AppSKey, false, ts.device.DevAddr, downPL.Item.FCnt, downPL.Item.FrmPayload)
					assert.NoError(err)
					assert.Equal(tst.expectedDownlinkPayload, b)
				}

				// assert integration event
				if tst.expectedIntegrationEvent != nil {
					pl := <-ts.integration.SendIntegrationNotificationChan
					assert.NotNil(pl.PublishedAt)
					pl.PublishedAt = nil
					assert.Equal(*tst.expectedIntegrationEvent, pl)
				} else {
					assert.Len(ts.integration.SendIntegrationNotificationChan, 0)
				}

				// assert location event
				if tst.expectedLocationEvent != nil {
					pl := <-ts.integration.SendLocationNotificationChan
					assert.NotNil(pl.PublishedAt)
					pl.PublishedAt = nil
					assert.Equal(*tst.expectedLocationEvent, pl)
				} else {
					assert.Len(ts.integration.SendLocationNotificationChan, 0)
				}
			})
		}
	})

	ts.T().Run("DAS join", func(t *testing.T) {
		assert := require.New(t)
		nowPB := ptypes.TimestampNow()
		now, err := ptypes.Timestamp(nowPB)
		assert.NoError(err)

		// set integration config
		ts.loraCloud.config = Config{
			DAS: true,
		}
		ts.apiRequest = ""

		// set response
		b, err := json.Marshal(das.UplinkResponse{
			Result: das.UplinkDeviceMapResponse{
				helpers.EUI64(ts.device.DevEUI): das.UplinkResponseItem{},
			},
		})
		assert.NoError(err)
		ts.apiResponse = string(b)

		// handle join
		assert.NoError(ts.loraCloud.HandleJoinEvent(context.Background(), ts.integration, nil, pb.JoinEvent{
			DevAddr: []byte{1, 2, 3, 4},
			DevEui:  ts.device.DevEUI[:],
			Dr:      3,
			TxInfo: &gw.UplinkTXInfo{
				Frequency: 868100000,
			},
			RxInfo: []*gw.UplinkRXInfo{
				{
					Time: nowPB,
				},
			}}))

		// assert request
		expected := das.UplinkRequest{
			helpers.EUI64(ts.device.DevEUI): &das.UplinkMsgJoining{
				MsgType:   "joining",
				Timestamp: float64(now.UnixNano()) / float64(time.Second),
				DR:        3,
				Freq:      868100000,
			},
		}
		b, err = json.Marshal(expected)
		assert.NoError(err)
		assert.Equal(string(b), ts.apiRequest)
	})
}

func (ts *LoRaCloudTestSuite) apiHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	ts.apiRequest = string(b)
	w.Write([]byte(ts.apiResponse))
}

func TestLoRaCloud(t *testing.T) {
	suite.Run(t, new(LoRaCloudTestSuite))
}
