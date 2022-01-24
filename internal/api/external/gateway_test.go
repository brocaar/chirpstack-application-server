package external

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

func (ts *APITestSuite) TestGateway() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	ctx := context.Background()
	validator := &TestValidator{
		returnSubject: "user",
	}
	api := NewGatewayAPI(validator)

	n := storage.NetworkServer{
		Name:   "test-ns-gw",
		Server: "test:12345",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &n))

	org := storage.Organization{
		Name:            "test-org-gw",
		MaxGatewayCount: 1,
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))

	org2 := storage.Organization{
		Name: "test-org-2",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org2))

	sp := storage.ServiceProfile{
		NetworkServerID: n.ID,
		OrganizationID:  org.ID,
		Name:            "test-sp",
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp))
	var spID uuid.UUID
	copy(spID[:], sp.ServiceProfile.Id)

	sp2 := storage.ServiceProfile{
		NetworkServerID: n.ID,
		OrganizationID:  org2.ID,
		Name:            "test-sp-2",
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp2))
	var sp2ID uuid.UUID
	copy(sp2ID[:], sp2.ServiceProfile.Id)

	adminUser := storage.User{
		Email:    "admin@user.com",
		IsActive: true,
		IsAdmin:  true,
	}
	assert.NoError(storage.CreateUser(context.Background(), storage.DB(), &adminUser))

	ts.T().Run("Create with service-profile under different org", func(t *testing.T) {
		assert := require.New(t)

		createReq := pb.CreateGatewayRequest{
			Gateway: &pb.Gateway{
				Id:          "0807060504030201",
				Name:        "test-gateway",
				Description: "test gateway",
				Location: &common.Location{
					Latitude:  1.1234,
					Longitude: 1.1235,
					Altitude:  5.5,
				},
				OrganizationId:   org.ID,
				DiscoveryEnabled: true,
				NetworkServerId:  n.ID,
				Boards: []*pb.GatewayBoard{
					{
						FpgaId: "0102030405060708",
					},
					{
						FineTimestampKey: "01020304050607080102030405060708",
					},
				},
				Tags: map[string]string{
					"foo": "bar",
				},
				Metadata:         make(map[string]string),
				ServiceProfileId: sp2ID.String(),
			},
		}
		_, err := api.Create(ctx, &createReq)
		assert.Equal(codes.InvalidArgument, grpc.Code(err))
	})

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		createReq := pb.CreateGatewayRequest{
			Gateway: &pb.Gateway{
				Id:          "0807060504030201",
				Name:        "test-gateway",
				Description: "test gateway",
				Location: &common.Location{
					Latitude:  1.1234,
					Longitude: 1.1235,
					Altitude:  5.5,
				},
				OrganizationId:   org.ID,
				DiscoveryEnabled: true,
				NetworkServerId:  n.ID,
				Boards: []*pb.GatewayBoard{
					{
						FpgaId: "0102030405060708",
					},
					{
						FineTimestampKey: "01020304050607080102030405060708",
					},
				},
				Tags: map[string]string{
					"foo": "bar",
				},
				Metadata:         make(map[string]string),
				ServiceProfileId: spID.String(),
			},
		}
		_, err := api.Create(ctx, &createReq)
		assert.NoError(err)

		nsReq := <-nsClient.CreateGatewayChan
		nsClient.GetGatewayResponse = ns.GetGatewayResponse{
			Gateway: nsReq.Gateway,
		}
		assert.Equal(applicationServerID.Bytes(), nsReq.Gateway.RoutingProfileId)

		t.Run("Create second exceeds max gateway count", func(t *testing.T) {
			assert := require.New(t)

			createReq := pb.CreateGatewayRequest{
				Gateway: &pb.Gateway{
					Id:          "0807060504030202",
					Name:        "test-gateway-2",
					Description: "test gateway",
					Location: &common.Location{
						Latitude:  1.1234,
						Longitude: 1.1235,
						Altitude:  5.5,
					},
					OrganizationId:  org.ID,
					NetworkServerId: n.ID,
				},
			}
			_, err := api.Create(ctx, &createReq)
			assert.Equal(codes.FailedPrecondition, grpc.Code(err))
			assert.Equal("rpc error: code = FailedPrecondition desc = organization reached max. gateway count", err.Error())
		})

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			getResp, err := api.Get(ctx, &pb.GetGatewayRequest{
				Id: createReq.Gateway.Id,
			})
			assert.NoError(err)
			assert.Equal(createReq.Gateway, getResp.Gateway)
			assert.NotEqual("", getResp.CreatedAt)
			assert.NotEqual("", getResp.UpdatedAt)
		})

		t.Run("List", func(t *testing.T) {
			assert := require.New(t)

			org2 := storage.Organization{
				Name: "test-org-gw-2",
			}
			assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org2))

			gw2 := storage.Gateway{
				Name:            "test-gw-2",
				MAC:             lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 2},
				OrganizationID:  org2.ID,
				NetworkServerID: n.ID,
			}
			assert.NoError(storage.CreateGateway(context.Background(), storage.DB(), &gw2))

			t.Run("List all", func(t *testing.T) {
				assert := require.New(t)

				validator.returnUser = adminUser
				gws, err := api.List(ctx, &pb.ListGatewayRequest{
					Limit: 10,
				})
				assert.NoError(err)
				assert.EqualValues(2, gws.TotalCount)
				assert.Len(gws.Result, 2)
			})

			t.Run("List as org user", func(t *testing.T) {
				user := storage.User{
					Email: "foo@bar.com",
				}
				err := storage.CreateUser(context.Background(), storage.DB(), &user)
				assert.NoError(err)
				validator.returnUser = user

				gws, err := api.List(ctx, &pb.ListGatewayRequest{
					Limit: 10,
				})
				assert.NoError(err)
				assert.EqualValues(0, gws.TotalCount)
				assert.Len(gws.Result, 0)

				assert.NoError(storage.CreateOrganizationUser(context.Background(), storage.DB(), org.ID, user.ID, false, false, false))

				gws, err = api.List(ctx, &pb.ListGatewayRequest{
					Limit: 10,
				})
				assert.NoError(err)
				assert.EqualValues(1, gws.TotalCount)
				assert.Len(gws.Result, 1)
				assert.Equal(createReq.Gateway.Id, gws.Result[0].Id)
			})

			t.Run("List for organization", func(t *testing.T) {
				assert := require.New(t)

				gws, err := api.List(ctx, &pb.ListGatewayRequest{
					Limit:          10,
					OrganizationId: org2.ID,
				})
				assert.NoError(err)
				assert.EqualValues(1, gws.TotalCount)
				assert.Len(gws.Result, 1)
				assert.Equal(gw2.MAC.String(), gws.Result[0].Id)
			})
		})

		t.Run("Update with service-profile under different org", func(t *testing.T) {
			assert := require.New(t)
			updateReq := pb.UpdateGatewayRequest{
				Gateway: &pb.Gateway{
					Id:   "0807060504030201",
					Name: "test-gateway-updated",
					Description: "test gateway updated	",
					Location: &common.Location{
						Latitude:  2.1234,
						Longitude: 2.1235,
						Altitude:  6.5,
					},
					OrganizationId:   org.ID,
					DiscoveryEnabled: true,
					NetworkServerId:  n.ID,
					Boards: []*pb.GatewayBoard{
						{
							FineTimestampKey: "02020304050607080102030405060708",
						},
						{
							FpgaId: "0202030405060708",
						},
					},
					Tags: map[string]string{
						"bar": "foo",
					},
					Metadata:         make(map[string]string),
					ServiceProfileId: sp2ID.String(),
				},
			}
			_, err := api.Update(ctx, &updateReq)
			assert.Equal(codes.InvalidArgument, grpc.Code(err))
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)
			updateReq := pb.UpdateGatewayRequest{
				Gateway: &pb.Gateway{
					Id:   "0807060504030201",
					Name: "test-gateway-updated",
					Description: "test gateway updated	",
					Location: &common.Location{
						Latitude:  2.1234,
						Longitude: 2.1235,
						Altitude:  6.5,
					},
					OrganizationId:   org.ID,
					DiscoveryEnabled: true,
					NetworkServerId:  n.ID,
					Boards: []*pb.GatewayBoard{
						{
							FineTimestampKey: "02020304050607080102030405060708",
						},
						{
							FpgaId: "0202030405060708",
						},
					},
					Tags: map[string]string{
						"bar": "foo",
					},
					Metadata: make(map[string]string),
				},
			}
			_, err := api.Update(ctx, &updateReq)
			assert.NoError(err)

			nsReq := <-nsClient.UpdateGatewayChan
			nsClient.GetGatewayResponse = ns.GetGatewayResponse{
				Gateway: nsReq.Gateway,
			}

			getResp, err := api.Get(ctx, &pb.GetGatewayRequest{
				Id: createReq.Gateway.Id,
			})
			assert.NoError(err)
			assert.Equal(updateReq.Gateway, getResp.Gateway)
			assert.NotEqual("", getResp.CreatedAt)
			assert.NotEqual("", getResp.UpdatedAt)
		})

		t.Run("GetStats", func(t *testing.T) {
			assert := require.New(t)
			assert.NoError(storage.SetAggregationIntervals([]storage.AggregationInterval{storage.AggregationMinute}))
			storage.SetMetricsTTL(time.Minute, time.Minute, time.Minute, time.Minute)

			now := time.Now().UTC()
			metrics := storage.MetricsRecord{
				Time: now,
				Metrics: map[string]float64{
					"rx_count":           10,
					"rx_ok_count":        5,
					"tx_count":           11,
					"tx_ok_count":        10,
					"tx_freq_868100000":  10,
					"rx_freq_868300000":  5,
					"tx_dr_3":            10,
					"rx_dr_2":            5,
					"tx_status_OK":       10,
					"tx_status_TOO_LATE": 1,
				},
			}
			assert.NoError(storage.SaveMetricsForInterval(context.Background(), storage.AggregationMinute, "gw:0102030405060708", metrics))

			start, _ := ptypes.TimestampProto(now.Truncate(time.Minute))
			end, _ := ptypes.TimestampProto(now)
			nowTrunc, _ := ptypes.TimestampProto(now.Truncate(time.Minute))

			resp, err := api.GetStats(ctx, &pb.GetGatewayStatsRequest{
				GatewayId:      "0102030405060708",
				Interval:       "MINUTE",
				StartTimestamp: start,
				EndTimestamp:   end,
			})
			assert.NoError(err)

			assert.Len(resp.Result, 1)
			assert.Equal(pb.GatewayStats{
				Timestamp:           nowTrunc,
				RxPacketsReceived:   10,
				RxPacketsReceivedOk: 5,
				TxPacketsReceived:   11,
				TxPacketsEmitted:    10,
				TxPacketsPerFrequency: map[uint32]uint32{
					868100000: 10,
				},
				RxPacketsPerFrequency: map[uint32]uint32{
					868300000: 5,
				},
				TxPacketsPerDr: map[uint32]uint32{
					3: 10,
				},
				RxPacketsPerDr: map[uint32]uint32{
					2: 5,
				},
				TxPacketsPerStatus: map[string]uint32{
					"OK":       10,
					"TOO_LATE": 1,
				},
			}, *resp.Result[0])
		})

		t.Run("GetLastPing", func(t *testing.T) {
			assert := require.New(t)

			gw, err := storage.GetGateway(context.Background(), storage.DB(), lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, false)
			assert.NoError(err)

			gw2 := storage.Gateway{
				OrganizationID:  org.ID,
				MAC:             lorawan.EUI64{2, 2, 3, 4, 5, 6, 7, 8},
				Name:            "test-gw-2",
				Description:     "test gw 2",
				NetworkServerID: n.ID,
			}
			assert.NoError(storage.CreateGateway(context.Background(), storage.DB(), &gw2))

			gw3 := storage.Gateway{
				OrganizationID:  org.ID,
				MAC:             lorawan.EUI64{3, 2, 3, 4, 5, 6, 7, 8},
				Name:            "test-gw-3",
				Description:     "test gw 3",
				NetworkServerID: n.ID,
			}
			assert.NoError(storage.CreateGateway(context.Background(), storage.DB(), &gw3))

			ping := storage.GatewayPing{
				GatewayMAC: lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
				Frequency:  868100000,
				DR:         5,
			}
			assert.NoError(storage.CreateGatewayPing(context.Background(), storage.DB(), &ping))
			ping.CreatedAt = ping.CreatedAt.Truncate(time.Millisecond)

			gw.LastPingID = &ping.ID
			assert.NoError(storage.UpdateGateway(context.Background(), storage.DB(), &gw))

			pingRX := []storage.GatewayPingRX{
				{
					PingID:     ping.ID,
					GatewayMAC: gw2.MAC,
					RSSI:       12,
					LoRaSNR:    5.5,
					Location: storage.GPSPoint{
						Latitude:  1.12345,
						Longitude: 2.12345,
					},
					Altitude: 10,
				},
				{
					PingID:     ping.ID,
					GatewayMAC: gw3.MAC,
					RSSI:       15,
					LoRaSNR:    7.5,
					Location: storage.GPSPoint{
						Latitude:  2.12345,
						Longitude: 3.12345,
					},
					Altitude: 11,
				},
			}
			for i := range pingRX {
				assert.NoError(storage.CreateGatewayPingRX(context.Background(), storage.DB(), &pingRX[i]))
			}

			pingResp, err := api.GetLastPing(ctx, &pb.GetLastPingRequest{
				GatewayId: createReq.Gateway.Id,
			})
			assert.NoError(err)

			assert.NotEqual("", pingResp.CreatedAt)
			assert.EqualValues(868100000, pingResp.Frequency)
			assert.EqualValues(5, pingResp.Dr)
			assert.Equal([]*pb.PingRX{
				{
					GatewayId: "0202030405060708",
					Rssi:      12,
					LoraSnr:   5.5,
					Latitude:  1.12345,
					Longitude: 2.12345,
					Altitude:  10,
				},
				{
					GatewayId: "0302030405060708",
					Rssi:      15,
					LoraSnr:   7.5,
					Latitude:  2.12345,
					Longitude: 3.12345,
					Altitude:  11,
				},
			}, pingResp.PingRx)
		})

		t.Run("GenerateGatewayClientCertificate", func(t *testing.T) {
			assert := require.New(t)

			nsClient.GenerateGatewayClientCertificateResponse = ns.GenerateGatewayClientCertificateResponse{
				TlsCert: []byte("foo"),
				TlsKey:  []byte("bar"),
				CaCert:  []byte("test"),
			}

			resp, err := api.GenerateGatewayClientCertificate(ctx, &pb.GenerateGatewayClientCertificateRequest{
				GatewayId: createReq.Gateway.Id,
			})
			assert.NoError(err)
			assert.Equal(&pb.GenerateGatewayClientCertificateResponse{
				TlsCert: "foo",
				TlsKey:  "bar",
				CaCert:  "test",
			}, resp)

		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			_, err := api.Delete(ctx, &pb.DeleteGatewayRequest{
				Id: createReq.Gateway.Id,
			})
			assert.NoError(err)

			_, err = api.Get(ctx, &pb.GetGatewayRequest{
				Id: createReq.Gateway.Id,
			})
			assert.Equal(codes.NotFound, grpc.Code(err))
		})
	})
}
