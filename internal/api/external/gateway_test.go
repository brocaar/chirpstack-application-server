package external

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/common"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

func (ts *APITestSuite) TestGateway() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	ctx := context.Background()
	validator := &TestValidator{}
	api := NewGatewayAPI(validator)

	n := storage.NetworkServer{
		Name:   "test-ns-gw",
		Server: "test:12345",
	}
	assert.NoError(storage.CreateNetworkServer(storage.DB(), &n))

	org := storage.Organization{
		Name: "test-org-gw",
	}
	assert.NoError(storage.CreateOrganization(storage.DB(), &org))

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
			},
		}
		_, err := api.Create(ctx, &createReq)
		assert.NoError(err)

		nsReq := <-nsClient.CreateGatewayChan
		nsClient.GetGatewayResponse = ns.GetGatewayResponse{
			Gateway: nsReq.Gateway,
		}

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
			assert.NoError(storage.CreateOrganization(storage.DB(), &org2))

			gw2 := storage.Gateway{
				Name:            "test-gw-2",
				MAC:             lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 2},
				OrganizationID:  org2.ID,
				NetworkServerID: n.ID,
			}
			assert.NoError(storage.CreateGateway(storage.DB(), &gw2))

			t.Run("List all", func(t *testing.T) {
				assert := require.New(t)

				validator.returnIsAdmin = true
				gws, err := api.List(ctx, &pb.ListGatewayRequest{
					Limit: 10,
				})
				assert.NoError(err)
				assert.EqualValues(2, gws.TotalCount)
				assert.Len(gws.Result, 2)
			})

			t.Run("List as org user", func(t *testing.T) {
				user := storage.User{
					Username: "testuser",
					Email:    "foo@bar.com",
				}
				_, err := storage.CreateUser(storage.DB(), &user, "password123")
				assert.NoError(err)
				validator.returnIsAdmin = false
				validator.returnUsername = user.Username

				gws, err := api.List(ctx, &pb.ListGatewayRequest{
					Limit: 10,
				})
				assert.NoError(err)
				assert.EqualValues(0, gws.TotalCount)
				assert.Len(gws.Result, 0)

				assert.NoError(storage.CreateOrganizationUser(storage.DB(), org.ID, user.ID, false))

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
			now := time.Now().UTC()
			nowPB, _ := ptypes.TimestampProto(now)
			now24hPB, _ := ptypes.TimestampProto(now.Add(24 * time.Hour))

			nsClient.GetGatewayStatsResponse = ns.GetGatewayStatsResponse{
				Result: []*ns.GatewayStats{
					{
						Timestamp:           nowPB,
						RxPacketsReceived:   10,
						RxPacketsReceivedOk: 9,
						TxPacketsReceived:   8,
						TxPacketsEmitted:    7,
					},
				},
			}

			statsResp, err := api.GetStats(ctx, &pb.GetGatewayStatsRequest{
				GatewayId:      createReq.Gateway.Id,
				Interval:       "DAY",
				StartTimestamp: nowPB,
				EndTimestamp:   now24hPB,
			})
			assert.NoError(err)

			assert.Equal(&pb.GetGatewayStatsResponse{
				Result: []*pb.GatewayStats{
					{
						Timestamp:           nowPB,
						RxPacketsReceived:   10,
						RxPacketsReceivedOk: 9,
						TxPacketsReceived:   8,
						TxPacketsEmitted:    7,
					},
				},
			}, statsResp)

			nsReq := <-nsClient.GetGatewayStatsChan
			assert.Equal(ns.GetGatewayStatsRequest{
				GatewayId:      []byte{8, 7, 6, 5, 4, 3, 2, 1},
				Interval:       ns.AggregationInterval_DAY,
				StartTimestamp: nowPB,
				EndTimestamp:   now24hPB,
			}, nsReq)
		})

		t.Run("GetLastPing", func(t *testing.T) {
			assert := require.New(t)

			gw, err := storage.GetGateway(storage.DB(), lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, false)
			assert.NoError(err)

			gw2 := storage.Gateway{
				OrganizationID:  org.ID,
				MAC:             lorawan.EUI64{2, 2, 3, 4, 5, 6, 7, 8},
				Name:            "test-gw-2",
				Description:     "test gw 2",
				NetworkServerID: n.ID,
			}
			assert.NoError(storage.CreateGateway(storage.DB(), &gw2))

			gw3 := storage.Gateway{
				OrganizationID:  org.ID,
				MAC:             lorawan.EUI64{3, 2, 3, 4, 5, 6, 7, 8},
				Name:            "test-gw-3",
				Description:     "test gw 3",
				NetworkServerID: n.ID,
			}
			assert.NoError(storage.CreateGateway(storage.DB(), &gw3))

			ping := storage.GatewayPing{
				GatewayMAC: lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
				Frequency:  868100000,
				DR:         5,
			}
			assert.NoError(storage.CreateGatewayPing(storage.DB(), &ping))
			ping.CreatedAt = ping.CreatedAt.Truncate(time.Millisecond)

			gw.LastPingID = &ping.ID
			assert.NoError(storage.UpdateGateway(storage.DB(), &gw))

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
				assert.NoError(storage.CreateGatewayPingRX(storage.DB(), &pingRX[i]))
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
