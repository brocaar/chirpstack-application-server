package api

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	. "github.com/smartystreets/goconvey/convey"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
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
		test.MustResetDB(db)

		nsClient := test.NewNetworkServerClient()
		ctx := context.Background()
		lsCtx := common.Context{NetworkServer: nsClient, DB: db}
		validator := &TestValidator{}
		api := NewGatewayAPI(lsCtx, validator)

		org := storage.Organization{
			Name: "test-organization",
		}
		So(storage.CreateOrganization(db, &org), ShouldBeNil)
		org2 := storage.Organization{
			Name: "test-organization-2",
		}
		So(storage.CreateOrganization(db, &org2), ShouldBeNil)

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
			Mac:            "0102030405060708",
			Name:           "test-gateway",
			Description:    "test gateway",
			Latitude:       1.1234,
			Longitude:      1.1235,
			Altitude:       5.5,
			FirstSeenAt:    now.UTC().Add(2 * time.Second).Format(time.RFC3339Nano),
			LastSeenAt:     now.UTC().Add(3 * time.Second).Format(time.RFC3339Nano),
			OrganizationID: org.ID,
		}

		Convey("When calling create", func() {
			_, err := api.Create(ctx, &pb.CreateGatewayRequest{
				Mac:            "0102030405060708",
				Name:           "test-gateway",
				Description:    "test gateway",
				Latitude:       1.1234,
				Longitude:      1.1235,
				Altitude:       5.5,
				OrganizationID: org.ID,
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

			Convey("Then the gateway was created in the db", func() {
				_, err := storage.GetGateway(db, lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8})
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
				So(storage.CreateOrganization(db, &org2), ShouldBeNil)
				gw2 := storage.Gateway{
					Name:           "test-gw-2",
					MAC:            lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
					OrganizationID: org2.ID,
				}
				So(storage.CreateGateway(db, &gw2), ShouldBeNil)

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
						user := storage.User{Username: "testuser"}
						_, err := storage.CreateUser(db, &user, "password123")
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

						So(storage.CreateOrganizationUser(db, org.ID, user.ID, false), ShouldBeNil)
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
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the gateway has been updated and the OrganizationID has been ignored", func() {
					gw, err := storage.GetGateway(db, lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8})
					So(err, ShouldBeNil)
					So(gw.Name, ShouldEqual, "test-gateway-updated")
					So(gw.Description, ShouldEqual, "updated test gateway")
					So(gw.OrganizationID, ShouldEqual, org.ID)
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
					gw, err := storage.GetGateway(db, lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8})
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
		})
	})
}
