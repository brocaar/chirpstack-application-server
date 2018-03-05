package api

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	. "github.com/smartystreets/goconvey/convey"

	pb "github.com/gusseleet/lora-app-server/api"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/gusseleet/lora-app-server/internal/test"
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
			Mac:         []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Name:        "test-gateway",
			Description: "test gateway",
			Latitude:    1.1234,
			Longitude:   1.1235,
			Altitude:    5.5,
			CreatedAt:   now.UTC().Format(time.RFC3339Nano),
			UpdatedAt:   now.UTC().Add(1 * time.Second).Format(time.RFC3339Nano),
			FirstSeenAt: now.UTC().Add(2 * time.Second).Format(time.RFC3339Nano),
			LastSeenAt:  now.UTC().Add(3 * time.Second).Format(time.RFC3339Nano),
		}

		getGatewayResponseAS := pb.GetGatewayResponse{
			Mac:             "0102030405060708",
			Name:            "test-gateway",
			Description:     "test gateway",
			Latitude:        1.1234,
			Longitude:       1.1235,
			Altitude:        5.5,
			FirstSeenAt:     now.UTC().Add(2 * time.Second).Format(time.RFC3339Nano),
			LastSeenAt:      now.UTC().Add(3 * time.Second).Format(time.RFC3339Nano),
			OrganizationID:  org.ID,
			Ping:            true,
			NetworkServerID: n.ID,
		}

		Convey("When calling create", func() {
			_, err := api.Create(ctx, &pb.CreateGatewayRequest{
				Mac:             "0102030405060708",
				Name:            "test-gateway",
				Description:     "test gateway",
				Latitude:        1.1234,
				Longitude:       1.1235,
				Altitude:        5.5,
				OrganizationID:  org.ID,
				Ping:            true,
				NetworkServerID: n.ID,
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the correct request was forwarded to the network-server api", func() {
				So(nsClient.CreateGatewayChan, ShouldHaveLength, 1)
				So(<-nsClient.CreateGatewayChan, ShouldResemble, ns.CreateGatewayRequest{
					Mac:         []byte{1, 2, 3, 4, 5, 6, 7, 8},
					Name:        "test-gateway",
					Description: "test gateway",
					Latitude:    1.1234,
					Longitude:   1.1235,
					Altitude:    5.5,
				})
			})

			Convey("Then the gateway was created in the config.C.PostgreSQL.DB", func() {
				_, err := storage.GetGateway(config.C.PostgreSQL.DB, lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}, false)
				So(err, ShouldBeNil)
			})

			Convey("When calling Get", func() {
				nsClient.GetGatewayResponse = getGatewayResponseNS
				resp, err := api.Get(ctx, &pb.GetGatewayRequest{
					Mac: "0102030405060708",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the expected response was returned", func() {
					So(resp.CreatedAt, ShouldNotEqual, "")
					So(resp.UpdatedAt, ShouldNotEqual, "")
					resp.CreatedAt = ""
					resp.UpdatedAt = ""
					So(resp, ShouldResemble, &getGatewayResponseAS)
				})

				Convey("Then the correct network-server request was made", func() {
					So(nsClient.GetGatewayChan, ShouldHaveLength, 1)
					So(<-nsClient.GetGatewayChan, ShouldResemble, ns.GetGatewayRequest{
						Mac: []byte{1, 2, 3, 4, 5, 6, 7, 8},
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
						So(gws.Result[0].Mac, ShouldEqual, "0102030405060708")
					})

				})

				Convey("Whe listing gateways for an organization", func() {
					Convey("Then only the gateways for that organization are returned", func() {
						gws, err := api.List(ctx, &pb.ListGatewayRequest{
							Limit:          10,
							Offset:         0,
							OrganizationID: org2.ID,
						})
						So(err, ShouldBeNil)
						So(gws.TotalCount, ShouldEqual, 1)
						So(gws.Result, ShouldHaveLength, 1)
						So(gws.Result[0].Mac, ShouldEqual, "0807060504030201")
					})
				})
			})

			Convey("When calling Update as non-admin", func() {
				validator.returnIsAdmin = false
				_, err := api.Update(ctx, &pb.UpdateGatewayRequest{
					Mac:            "0102030405060708",
					Name:           "test-gateway-updated",
					Description:    "updated test gateway",
					Latitude:       1.1235,
					Longitude:      1.1236,
					Altitude:       5.7,
					OrganizationID: org2.ID,
					Ping:           false,
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
						Mac:         []byte{1, 2, 3, 4, 5, 6, 7, 8},
						Name:        "test-gateway-updated",
						Description: "updated test gateway",
						Latitude:    1.1235,
						Longitude:   1.1236,
						Altitude:    5.7,
					})
				})
			})

			Convey("When calling Update as an admin", func() {
				validator.returnIsAdmin = true
				_, err := api.Update(ctx, &pb.UpdateGatewayRequest{
					Mac:            "0102030405060708",
					Name:           "test-gateway-updated",
					Description:    "updated test gateway",
					Latitude:       1.1235,
					Longitude:      1.1236,
					Altitude:       5.7,
					OrganizationID: org2.ID,
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
						Mac:         []byte{1, 2, 3, 4, 5, 6, 7, 8},
						Name:        "test-gateway-updated",
						Description: "updated test gateway",
						Latitude:    1.1235,
						Longitude:   1.1236,
						Altitude:    5.7,
					})
				})
			})

			Convey("When calling Delete", func() {
				_, err := api.Delete(ctx, &pb.DeleteGatewayRequest{
					Mac: "0102030405060708",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the gateway has been deleted", func() {
					_, err := api.Get(ctx, &pb.GetGatewayRequest{
						Mac: "0102030405060708",
					})
					So(err, ShouldNotBeNil)
					So(grpc.Code(err), ShouldEqual, codes.NotFound)
				})

				Convey("Then the expected request was sent to the network-server", func() {
					So(nsClient.DeleteGatewayChan, ShouldHaveLength, 1)
					So(<-nsClient.DeleteGatewayChan, ShouldResemble, ns.DeleteGatewayRequest{
						Mac: []byte{1, 2, 3, 4, 5, 6, 7, 8},
					})
				})
			})

			Convey("When calling GenerateToken", func() {
				nsClient.GenerateGatewayTokenResponse = ns.GenerateGatewayTokenResponse{
					Token: "secrettoken",
				}

				tokenResp, err := api.GenerateToken(ctx, &pb.GenerateGatewayTokenRequest{
					Mac: "0102030405060708",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the exepcted token was returned", func() {
					So(tokenResp.Token, ShouldEqual, "secrettoken")
				})

				Convey("Then the expected request wat sent to the network-server", func() {
					So(nsClient.GenerateGatewayTokenChan, ShouldHaveLength, 1)
					So(<-nsClient.GenerateGatewayTokenChan, ShouldResemble, ns.GenerateGatewayTokenRequest{
						Mac: []byte{1, 2, 3, 4, 5, 6, 7, 8},
					})
				})
			})

			Convey("When calling GetStats", func() {
				now := time.Now().UTC()

				nsClient.GetGatewayStatsResponse = ns.GetGatewayStatsResponse{
					Result: []*ns.GatewayStats{
						{
							Timestamp:           now.Format(time.RFC3339Nano),
							RxPacketsReceived:   10,
							RxPacketsReceivedOK: 9,
							TxPacketsReceived:   8,
							TxPacketsEmitted:    7,
						},
					},
				}

				resp, err := api.GetStats(ctx, &pb.GetGatewayStatsRequest{
					Mac:            "0102030405060708",
					Interval:       "DAY",
					StartTimestamp: now.Format(time.RFC3339Nano),
					EndTimestamp:   now.Add(24 * time.Hour).Format(time.RFC3339Nano),
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the correct request / response was forwarded to / from the network-server api", func() {
					So(nsClient.GetGatewayStatsChan, ShouldHaveLength, 1)
					So(<-nsClient.GetGatewayStatsChan, ShouldResemble, ns.GetGatewayStatsRequest{
						Mac:            []byte{1, 2, 3, 4, 5, 6, 7, 8},
						Interval:       ns.AggregationInterval_DAY,
						StartTimestamp: now.Format(time.RFC3339Nano),
						EndTimestamp:   now.Add(24 * time.Hour).Format(time.RFC3339Nano),
					})

					So(resp, ShouldResemble, &pb.GetGatewayStatsResponse{
						Result: []*pb.GatewayStats{
							{
								Timestamp:           now.Format(time.RFC3339Nano),
								RxPacketsReceived:   10,
								RxPacketsReceivedOK: 9,
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
						Mac: "0102030405060708",
					})
					So(err, ShouldBeNil)

					Convey("Then the expected result is returned", func() {
						createdAt, err := time.Parse(time.RFC3339Nano, resp.CreatedAt)
						So(err, ShouldBeNil)
						So(createdAt.Truncate(time.Millisecond).Equal(ping.CreatedAt), ShouldBeTrue)
						So(resp.Frequency, ShouldEqual, 868100000)
						So(resp.Dr, ShouldEqual, 5)
						So(resp.PingRX, ShouldResemble, []*pb.PingRX{
							{
								Mac:       "0202030405060708",
								Rssi:      12,
								LoraSNR:   5.5,
								Latitude:  1.12345,
								Longitude: 2.12345,
								Altitude:  10,
							},
							{
								Mac:       "0302030405060708",
								Rssi:      15,
								LoraSNR:   7.5,
								Latitude:  2.12345,
								Longitude: 3.12345,
								Altitude:  11,
							},
						})
					})
				})
			})

			Convey("When calling CreateChannelConfiguration", func() {
				nsClient.CreateChannelConfigurationResponse = ns.CreateChannelConfigurationResponse{
					Id: 123,
				}

				resp, err := api.CreateChannelConfiguration(ctx, &pb.CreateChannelConfigurationRequest{
					Name:            "test-config",
					Channels:        []int32{0, 1, 2},
					NetworkServerID: n.ID,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the exepcted request was made to the network-server", func() {
					So(resp, ShouldResemble, &pb.CreateChannelConfigurationResponse{
						Id: 123,
					})
					So(nsClient.CreateChannelConfigurationChan, ShouldHaveLength, 1)
					So(<-nsClient.CreateChannelConfigurationChan, ShouldResemble, ns.CreateChannelConfigurationRequest{
						Name:     "test-config",
						Channels: []int32{0, 1, 2},
					})
				})
			})

			Convey("When calling GetChannelConfiguration", func() {
				now := time.Now()
				nowStr := now.Format(time.RFC3339Nano)

				nsClient.GetChannelConfigurationResponse = ns.GetChannelConfigurationResponse{
					Id:        123,
					Name:      "test-config",
					Channels:  []int32{0, 1, 2},
					CreatedAt: nowStr,
					UpdatedAt: nowStr,
				}

				resp, err := api.GetChannelConfiguration(ctx, &pb.GetChannelConfigurationRequest{
					Id:              123,
					NetworkServerID: n.ID,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the expected request was made to the network-server", func() {
					So(resp, ShouldResemble, &pb.GetChannelConfigurationResponse{
						Id:              123,
						Name:            "test-config",
						Channels:        []int32{0, 1, 2},
						CreatedAt:       nowStr,
						UpdatedAt:       nowStr,
						NetworkServerID: n.ID,
					})
					So(nsClient.GetChannelConfigurationChan, ShouldHaveLength, 1)
					So(<-nsClient.GetChannelConfigurationChan, ShouldResemble, ns.GetChannelConfigurationRequest{
						Id: 123,
					})
				})
			})

			Convey("When calling UpdateChannelConfiguration", func() {
				resp, err := api.UpdateChannelConfiguration(ctx, &pb.UpdateChannelConfigurationRequest{
					Id:              123,
					Name:            "updated-config",
					Channels:        []int32{0, 1},
					NetworkServerID: n.ID,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the expected request was made to the network-server", func() {
					So(resp, ShouldResemble, &pb.UpdateChannelConfigurationResponse{})
					So(nsClient.UpdateChannelConfigurationChan, ShouldHaveLength, 1)
					So(<-nsClient.UpdateChannelConfigurationChan, ShouldResemble, ns.UpdateChannelConfigurationRequest{
						Id:       123,
						Name:     "updated-config",
						Channels: []int32{0, 1},
					})
				})
			})

			Convey("When calling DeleteChannelConfiguration", func() {
				resp, err := api.DeleteChannelConfiguration(ctx, &pb.DeleteChannelConfigurationRequest{
					Id:              123,
					NetworkServerID: n.ID,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the expected request was made to the network-server", func() {
					So(resp, ShouldResemble, &pb.DeleteChannelConfigurationResponse{})
					So(nsClient.DeleteChannelConfigurationChan, ShouldHaveLength, 1)
					So(<-nsClient.DeleteChannelConfigurationChan, ShouldResemble, ns.DeleteChannelConfigurationRequest{
						Id: 123,
					})

				})
			})

			Convey("When calling ListChannelConfigurations", func() {
				now := time.Now()
				nowStr := now.Format(time.RFC3339Nano)

				nsClient.ListChannelConfigurationsResponse = ns.ListChannelConfigurationsResponse{
					Result: []*ns.GetChannelConfigurationResponse{
						{
							Id:        123,
							Name:      "test-config",
							Channels:  []int32{0, 1, 2},
							CreatedAt: nowStr,
							UpdatedAt: nowStr,
						},
					},
				}

				resp, err := api.ListChannelConfigurations(ctx, &pb.ListChannelConfigurationsRequest{
					NetworkServerID: n.ID,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the expected request was made to the network-server", func() {
					So(resp, ShouldResemble, &pb.ListChannelConfigurationsResponse{
						Result: []*pb.GetChannelConfigurationResponse{
							{
								Id:              123,
								Name:            "test-config",
								Channels:        []int32{0, 1, 2},
								CreatedAt:       nowStr,
								UpdatedAt:       nowStr,
								NetworkServerID: n.ID,
							},
						},
					})
					So(nsClient.ListChannelConfigurationsChan, ShouldHaveLength, 1)
					So(<-nsClient.ListChannelConfigurationsChan, ShouldResemble, ns.ListChannelConfigurationsRequest{})
				})
			})

			Convey("When calling CreateExtraChannel", func() {
				nsClient.CreateExtraChannelResponse = ns.CreateExtraChannelResponse{
					Id: 321,
				}

				resp, err := api.CreateExtraChannel(ctx, &pb.CreateExtraChannelRequest{
					ChannelConfigurationID: 123,
					Modulation:             pb.Modulation_LORA,
					Frequency:              867100000,
					BandWidth:              125,
					BitRate:                50000,
					SpreadFactors:          []int32{5},
					NetworkServerID:        n.ID,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the expected request was made to the network-server", func() {
					So(resp, ShouldResemble, &pb.CreateExtraChannelResponse{
						Id: 321,
					})
					So(nsClient.CreateExtraChannelChan, ShouldHaveLength, 1)
					So(<-nsClient.CreateExtraChannelChan, ShouldResemble, ns.CreateExtraChannelRequest{
						ChannelConfigurationID: 123,
						Modulation:             ns.Modulation_LORA,
						Frequency:              867100000,
						BandWidth:              125,
						BitRate:                50000,
						SpreadFactors:          []int32{5},
					})
				})
			})

			Convey("When calling UpdateExtraChannel", func() {
				resp, err := api.UpdateExtraChannel(ctx, &pb.UpdateExtraChannelRequest{
					Id: 321,
					ChannelConfigurationID: 123,
					Modulation:             pb.Modulation_LORA,
					Frequency:              867100000,
					BandWidth:              125,
					BitRate:                50000,
					SpreadFactors:          []int32{5},
					NetworkServerID:        n.ID,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the expected request was made to the network-server", func() {
					So(resp, ShouldResemble, &pb.UpdateExtraChannelResponse{})
					So(nsClient.UpdateExtraChannelChan, ShouldHaveLength, 1)
					So(<-nsClient.UpdateExtraChannelChan, ShouldResemble, ns.UpdateExtraChannelRequest{
						Id: 321,
						ChannelConfigurationID: 123,
						Modulation:             ns.Modulation_LORA,
						Frequency:              867100000,
						BandWidth:              125,
						BitRate:                50000,
						SpreadFactors:          []int32{5},
					})
				})
			})

			Convey("When calling DeleteExtraChannel", func() {
				resp, err := api.DeleteExtraChannel(ctx, &pb.DeleteExtraChannelRequest{
					Id:              321,
					NetworkServerID: n.ID,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the expected request was made to the network-server", func() {
					So(resp, ShouldResemble, &pb.DeleteExtraChannelResponse{})
					So(nsClient.DeleteExtraChannelChan, ShouldHaveLength, 1)
					So(<-nsClient.DeleteExtraChannelChan, ShouldResemble, ns.DeleteExtraChannelRequest{
						Id: 321,
					})
				})
			})

			Convey("When calling GetExtraChannelsForChannelConfigurationID", func() {
				now := time.Now()
				nowStr := now.Format(time.RFC3339Nano)

				nsClient.GetExtraChannelsForChannelConfigurationIDResponse = ns.GetExtraChannelsForChannelConfigurationIDResponse{
					Result: []*ns.GetExtraChannelResponse{
						{
							Id: 321,
							ChannelConfigurationID: 123,
							CreatedAt:              nowStr,
							UpdatedAt:              nowStr,
							Modulation:             ns.Modulation_LORA,
							Frequency:              867100000,
							Bandwidth:              125,
							BitRate:                50000,
							SpreadFactors:          []int32{5},
						},
					},
				}

				resp, err := api.GetExtraChannelsForChannelConfigurationID(ctx, &pb.GetExtraChannelsForChannelConfigurationIDRequest{
					Id:              123,
					NetworkServerID: n.ID,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the expected request was made to the network-server", func() {
					So(resp, ShouldResemble, &pb.GetExtraChannelsForChannelConfigurationIDResponse{
						Result: []*pb.GetExtraChannelResponse{
							{
								Id: 321,
								ChannelConfigurationID: 123,
								CreatedAt:              nowStr,
								UpdatedAt:              nowStr,
								Modulation:             pb.Modulation_LORA,
								Frequency:              867100000,
								Bandwidth:              125,
								BitRate:                50000,
								SpreadFactors:          []int32{5},
							},
						},
					})
					So(nsClient.GetExtraChannelsForChannelConfigurationIDChan, ShouldHaveLength, 1)
					So(<-nsClient.GetExtraChannelsForChannelConfigurationIDChan, ShouldResemble, ns.GetExtraChannelsForChannelConfigurationIDRequest{
						Id: 123,
					})
				})
			})
		})
	})
}
