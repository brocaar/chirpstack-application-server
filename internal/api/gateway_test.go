package api

import (
	"context"
	"testing"
	"time"

	"github.com/brocaar/loraserver/api/gw"

	"github.com/golang/protobuf/ptypes"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

func TestGatewayAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given clean database with an organization and a mocked network-server api", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		config.C.PostgreSQL.DB = db
		test.MustResetDB(config.C.PostgreSQL.DB)

		nsClient := test.NewNetworkServerClient()
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewGatewayAPI(validator)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

		org := storage.Organization{
			Name: "test-organization",
		}
		So(storage.CreateOrganization(config.C.PostgreSQL.DB, &org), ShouldBeNil)

		org2 := storage.Organization{
			Name: "test-organization-2",
		}
		So(storage.CreateOrganization(config.C.PostgreSQL.DB, &org2), ShouldBeNil)

		now := time.Now().UTC()
		getGatewayResponseNS := ns.GetGatewayResponse{
			Gateway: &ns.Gateway{
				Id: []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Location: &gw.Location{
					Latitude:  1.1234,
					Longitude: 1.1235,
					Altitude:  5.5,
				},
			},
		}

		getGatewayResponseNS.CreatedAt, _ = ptypes.TimestampProto(now)
		getGatewayResponseNS.UpdatedAt, _ = ptypes.TimestampProto(now.Add(time.Second))
		getGatewayResponseNS.FirstSeenAt, _ = ptypes.TimestampProto(now.Add(2 * time.Second))
		getGatewayResponseNS.LastSeenAt, _ = ptypes.TimestampProto(now.Add(3 * time.Second))

		Convey("When calling create", func() {
			createReq := pb.CreateGatewayRequest{
				Gateway: &pb.Gateway{
					Id:          "0102030405060708",
					Name:        "test-gateway",
					Description: "test gateway",
					Location: &gw.Location{
						Latitude:  1.1234,
						Longitude: 1.1235,
						Altitude:  5.5,
					},
					OrganizationId:   org.ID,
					DiscoveryEnabled: true,
					NetworkServerId:  n.ID,
				},
			}

			_, err := api.Create(ctx, &createReq)
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the correct request was forwarded to the network-server api", func() {
				So(nsClient.CreateGatewayChan, ShouldHaveLength, 1)
				So(<-nsClient.CreateGatewayChan, ShouldResemble, ns.CreateGatewayRequest{
					Gateway: &ns.Gateway{
						Id: []byte{1, 2, 3, 4, 5, 6, 7, 8},
						Location: &gw.Location{
							Latitude:  1.1234,
							Longitude: 1.1235,
							Altitude:  5.5,
						},
					},
				})
			})

			Convey("When calling Get", func() {
				nsClient.GetGatewayResponse = getGatewayResponseNS

				resp, err := api.Get(ctx, &pb.GetGatewayRequest{
					Id: "0102030405060708",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the expected response was returned", func() {
					So(resp.CreatedAt, ShouldNotBeNil)
					So(resp.UpdatedAt, ShouldNotBeNil)
					So(resp.Gateway, ShouldResemble, createReq.Gateway)
				})

				Convey("Then the correct network-server request was made", func() {
					So(nsClient.GetGatewayChan, ShouldHaveLength, 1)
					So(<-nsClient.GetGatewayChan, ShouldResemble, ns.GetGatewayRequest{
						Id: []byte{1, 2, 3, 4, 5, 6, 7, 8},
					})
				})
			})

			Convey("Given an extra gateway beloning to a different organization", func() {
				org2 := storage.Organization{
					Name: "test-org-2",
				}
				So(storage.CreateOrganization(config.C.PostgreSQL.DB, &org2), ShouldBeNil)

				gw2 := storage.Gateway{
					Name:            "test-gw-2",
					MAC:             lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
					OrganizationID:  org2.ID,
					NetworkServerID: n.ID,
				}
				So(storage.CreateGateway(config.C.PostgreSQL.DB, &gw2), ShouldBeNil)

				Convey("When listing all gateways", func() {
					Convey("Then all gateways are visible to an admin user", func() {
						validator.returnIsAdmin = true
						gws, err := api.List(ctx, &pb.ListGatewayRequest{
							Limit:  10,
							Offset: 0,
						})
						So(err, ShouldBeNil)
						So(gws.TotalCount, ShouldEqual, 2)
						So(gws.Result, ShouldHaveLength, 2)
					})

					Convey("Then gateways are only visible to users assigned to an organization", func() {
						user := storage.User{
							Username: "testuser",
							Email:    "foo@bar.com",
						}
						_, err := storage.CreateUser(config.C.PostgreSQL.DB, &user, "password123")
						So(err, ShouldBeNil)
						validator.returnIsAdmin = false
						validator.returnUsername = user.Username

						gws, err := api.List(ctx, &pb.ListGatewayRequest{
							Limit:  10,
							Offset: 0,
						})
						So(err, ShouldBeNil)
						So(gws.TotalCount, ShouldEqual, 0)
						So(gws.Result, ShouldHaveLength, 0)

						So(storage.CreateOrganizationUser(config.C.PostgreSQL.DB, org.ID, user.ID, false), ShouldBeNil)
						gws, err = api.List(ctx, &pb.ListGatewayRequest{
							Limit:  10,
							Offset: 0,
						})
						So(err, ShouldBeNil)
						So(gws.TotalCount, ShouldEqual, 1)
						So(gws.Result, ShouldHaveLength, 1)
						So(gws.Result[0].Id, ShouldEqual, "0102030405060708")
					})

				})

				Convey("Whe listing gateways for an organization", func() {
					Convey("Then only the gateways for that organization are returned", func() {
						gws, err := api.List(ctx, &pb.ListGatewayRequest{
							Limit:          10,
							Offset:         0,
							OrganizationId: org2.ID,
						})
						So(err, ShouldBeNil)
						So(gws.TotalCount, ShouldEqual, 1)
						So(gws.Result, ShouldHaveLength, 1)
						So(gws.Result[0].Id, ShouldEqual, "0807060504030201")
					})
				})
			})

			Convey("When calling Update as non-admin", func() {
				validator.returnIsAdmin = false
				_, err := api.Update(ctx, &pb.UpdateGatewayRequest{
					Gateway: &pb.Gateway{
						Id:          "0102030405060708",
						Name:        "test-gateway-updated",
						Description: "updated test gateway",
						Location: &gw.Location{
							Latitude:  1.1235,
							Longitude: 1.1236,
							Altitude:  5.7,
						},
						OrganizationId: org2.ID,
					},
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the gateway has been updated and the OrganizationID has been ignored", func() {
					gw, err := storage.GetGateway(config.C.PostgreSQL.DB, lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}, false)
					So(err, ShouldBeNil)
					So(gw.Name, ShouldEqual, "test-gateway-updated")
					So(gw.Description, ShouldEqual, "updated test gateway")
					So(gw.OrganizationID, ShouldEqual, org.ID)
					So(gw.Ping, ShouldBeFalse)
				})

				Convey("Then the expected request was sent to the network-server", func() {
					So(nsClient.UpdateGatewayChan, ShouldHaveLength, 1)
					So(<-nsClient.UpdateGatewayChan, ShouldResemble, ns.UpdateGatewayRequest{
						Gateway: &ns.Gateway{
							Id: []byte{1, 2, 3, 4, 5, 6, 7, 8},
							Location: &gw.Location{
								Latitude:  1.1235,
								Longitude: 1.1236,
								Altitude:  5.7,
							},
						},
					})
				})
			})

			Convey("When calling Update as an admin", func() {
				validator.returnIsAdmin = true
				_, err := api.Update(ctx, &pb.UpdateGatewayRequest{
					Gateway: &pb.Gateway{
						Id:          "0102030405060708",
						Name:        "test-gateway-updated",
						Description: "updated test gateway",
						Location: &gw.Location{
							Latitude:  1.1235,
							Longitude: 1.1236,
							Altitude:  5.7,
						},
						OrganizationId: org2.ID,
					},
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the gateway has been updated", func() {
					gw, err := storage.GetGateway(config.C.PostgreSQL.DB, lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}, false)
					So(err, ShouldBeNil)
					So(gw.Name, ShouldEqual, "test-gateway-updated")
					So(gw.Description, ShouldEqual, "updated test gateway")
					So(gw.OrganizationID, ShouldEqual, org2.ID)
				})

				Convey("Then the expected request was sent to the network-server", func() {
					So(nsClient.UpdateGatewayChan, ShouldHaveLength, 1)
					So(<-nsClient.UpdateGatewayChan, ShouldResemble, ns.UpdateGatewayRequest{
						Gateway: &ns.Gateway{
							Id: []byte{1, 2, 3, 4, 5, 6, 7, 8},
							Location: &gw.Location{
								Latitude:  1.1235,
								Longitude: 1.1236,
								Altitude:  5.7,
							},
						},
					})
				})
			})

			Convey("When calling Delete", func() {
				_, err := api.Delete(ctx, &pb.DeleteGatewayRequest{
					Id: "0102030405060708",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the gateway has been deleted", func() {
					_, err := api.Get(ctx, &pb.GetGatewayRequest{
						Id: "0102030405060708",
					})
					So(err, ShouldNotBeNil)
					So(grpc.Code(err), ShouldEqual, codes.NotFound)
				})

				Convey("Then the expected request was sent to the network-server", func() {
					So(nsClient.DeleteGatewayChan, ShouldHaveLength, 1)
					So(<-nsClient.DeleteGatewayChan, ShouldResemble, ns.DeleteGatewayRequest{
						Id: []byte{1, 2, 3, 4, 5, 6, 7, 8},
					})
				})
			})

			Convey("When calling GetStats", func() {
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

				resp, err := api.GetStats(ctx, &pb.GetGatewayStatsRequest{
					GatewayId:      "0102030405060708",
					Interval:       "DAY",
					StartTimestamp: nowPB,
					EndTimestamp:   now24hPB,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the correct request / response was forwarded to / from the network-server api", func() {
					So(nsClient.GetGatewayStatsChan, ShouldHaveLength, 1)
					So(<-nsClient.GetGatewayStatsChan, ShouldResemble, ns.GetGatewayStatsRequest{
						GatewayId:      []byte{1, 2, 3, 4, 5, 6, 7, 8},
						Interval:       ns.AggregationInterval_DAY,
						StartTimestamp: nowPB,
						EndTimestamp:   now24hPB,
					})

					So(resp, ShouldResemble, &pb.GetGatewayStatsResponse{
						Result: []*pb.GatewayStats{
							{
								Timestamp:           nowPB,
								RxPacketsReceived:   10,
								RxPacketsReceivedOk: 9,
								TxPacketsReceived:   8,
								TxPacketsEmitted:    7,
							},
						},
					})
				})
			})

			Convey("Given gateway ping data", func() {
				gw, err := storage.GetGateway(config.C.PostgreSQL.DB, lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}, false)
				So(err, ShouldBeNil)

				gw2 := storage.Gateway{
					OrganizationID:  org.ID,
					MAC:             lorawan.EUI64{2, 2, 3, 4, 5, 6, 7, 8},
					Name:            "test-gw-2",
					Description:     "test gw 2",
					NetworkServerID: n.ID,
				}
				So(storage.CreateGateway(config.C.PostgreSQL.DB, &gw2), ShouldBeNil)

				gw3 := storage.Gateway{
					OrganizationID:  org.ID,
					MAC:             lorawan.EUI64{3, 2, 3, 4, 5, 6, 7, 8},
					Name:            "test-gw-3",
					Description:     "test gw 3",
					NetworkServerID: n.ID,
				}
				So(storage.CreateGateway(config.C.PostgreSQL.DB, &gw3), ShouldBeNil)

				ping := storage.GatewayPing{
					GatewayMAC: lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
					Frequency:  868100000,
					DR:         5,
				}
				So(storage.CreateGatewayPing(config.C.PostgreSQL.DB, &ping), ShouldBeNil)
				ping.CreatedAt = ping.CreatedAt.Truncate(time.Millisecond)

				gw.LastPingID = &ping.ID
				So(storage.UpdateGateway(config.C.PostgreSQL.DB, &gw), ShouldBeNil)

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
					So(storage.CreateGatewayPingRX(config.C.PostgreSQL.DB, &pingRX[i]), ShouldBeNil)
				}

				Convey("When calling GetLastPing", func() {
					resp, err := api.GetLastPing(ctx, &pb.GetLastPingRequest{
						GatewayId: "0102030405060708",
					})
					So(err, ShouldBeNil)

					Convey("Then the expected result is returned", func() {
						So(resp.CreatedAt, ShouldNotBeNil)
						So(resp.Frequency, ShouldEqual, 868100000)
						So(resp.Dr, ShouldEqual, 5)
						So(resp.PingRx, ShouldResemble, []*pb.PingRX{
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
						})
					})
				})
			})
		})
	})
}
