package api

import (
	"testing"

	"github.com/brocaar/loraserver/api/ns"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
)

func TestGatewayProfileTest(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	config.C.PostgreSQL.DB = db

	Convey("Given a clean database and api instance", t, func() {
		test.MustResetDB(db)

		nsClient := test.NewNetworkServerClient()
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewGatewayProfileAPI(validator)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

		Convey("Then Create creates the gateway-profile", func() {
			createReq := pb.CreateGatewayProfileRequest{
				Name:            "test-gp",
				NetworkServerID: n.ID,
				GatewayProfile: &pb.GatewayProfile{
					Channels: []uint32{0, 1, 2},
					ExtraChannels: []*pb.GatewayProfileExtraChannel{
						{
							Modulation:       pb.Modulation_LORA,
							Frequency:        867100000,
							Bandwidth:        125,
							SpreadingFactors: []uint32{10, 11, 12},
						},
						{
							Modulation: pb.Modulation_FSK,
							Frequency:  867300000,
							Bitrate:    50000,
						},
					},
				},
			}

			createResp, err := api.Create(ctx, &createReq)
			So(err, ShouldBeNil)
			So(createResp.GatewayProfileID, ShouldNotEqual, "")
			So(nsClient.CreateGatewayProfileChan, ShouldHaveLength, 1)

			// set mock
			nsCreate := <-nsClient.CreateGatewayProfileChan
			nsClient.GetGatewayProfileResponse = ns.GetGatewayProfileResponse{
				GatewayProfile: nsCreate.GatewayProfile,
			}

			Convey("Then Get returns the gateway-profile", func() {
				getResp, err := api.Get(ctx, &pb.GetGatewayProfileRequest{
					GatewayProfileID: createResp.GatewayProfileID,
				})
				So(err, ShouldBeNil)
				So(getResp.Name, ShouldEqual, createReq.Name)
				So(getResp.NetworkServerID, ShouldEqual, createReq.NetworkServerID)
				So(getResp.GatewayProfile, ShouldResemble, &pb.GatewayProfile{
					GatewayProfileID: createResp.GatewayProfileID,
					Channels:         []uint32{0, 1, 2},
					ExtraChannels: []*pb.GatewayProfileExtraChannel{
						{
							Modulation:       pb.Modulation_LORA,
							Frequency:        867100000,
							Bandwidth:        125,
							SpreadingFactors: []uint32{10, 11, 12},
						},
						{
							Modulation: pb.Modulation_FSK,
							Frequency:  867300000,
							Bitrate:    50000,
						},
					},
				})
			})

			Convey("Then Update updates the gateway-profile", func() {
				_, err := api.Update(ctx, &pb.UpdateGatewayProfileRequest{
					Name: "updated-gp",
					GatewayProfile: &pb.GatewayProfile{
						GatewayProfileID: createResp.GatewayProfileID,
						Channels:         []uint32{1, 2},
						ExtraChannels: []*pb.GatewayProfileExtraChannel{
							{
								Modulation: pb.Modulation_FSK,
								Frequency:  867300000,
								Bitrate:    50000,
							},
							{
								Modulation:       pb.Modulation_LORA,
								Frequency:        867100000,
								Bandwidth:        125,
								SpreadingFactors: []uint32{10, 11, 12},
							},
						},
					},
				})
				So(err, ShouldBeNil)
				So(nsClient.UpdateGatewayProfileChan, ShouldHaveLength, 1)

				// set mock
				nsUpdate := <-nsClient.UpdateGatewayProfileChan
				nsClient.GetGatewayProfileResponse = ns.GetGatewayProfileResponse{
					GatewayProfile: nsUpdate.GatewayProfile,
				}

				getResp, err := api.Get(ctx, &pb.GetGatewayProfileRequest{
					GatewayProfileID: createResp.GatewayProfileID,
				})
				So(err, ShouldBeNil)
				So(getResp.Name, ShouldEqual, "updated-gp")
				So(getResp.NetworkServerID, ShouldEqual, createReq.NetworkServerID)
				So(getResp.GatewayProfile, ShouldResemble, &pb.GatewayProfile{
					GatewayProfileID: createResp.GatewayProfileID,
					Channels:         []uint32{1, 2},
					ExtraChannels: []*pb.GatewayProfileExtraChannel{
						{
							Modulation: pb.Modulation_FSK,
							Frequency:  867300000,
							Bitrate:    50000,
						},
						{
							Modulation:       pb.Modulation_LORA,
							Frequency:        867100000,
							Bandwidth:        125,
							SpreadingFactors: []uint32{10, 11, 12},
						},
					},
				})
			})

			Convey("Then Delete deletes the gateway-profile", func() {
				_, err := api.Delete(ctx, &pb.DeleteGatewayProfileRequest{
					GatewayProfileID: createResp.GatewayProfileID,
				})
				So(err, ShouldBeNil)

				_, err = api.Get(ctx, &pb.GetGatewayProfileRequest{
					GatewayProfileID: createResp.GatewayProfileID,
				})
				So(err, ShouldNotBeNil)
				So(grpc.Code(err), ShouldEqual, codes.NotFound)
			})

			Convey("Then List given a network-server ID lists the gateway-profiles", func() {
				listResp, err := api.List(ctx, &pb.ListGatewayProfilesRequest{
					NetworkServerID: n.ID,
					Limit:           10,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 1)
				So(listResp.Result, ShouldHaveLength, 1)
				So(listResp.Result[0].GatewayProfileID, ShouldEqual, createResp.GatewayProfileID)
				So(listResp.Result[0].Name, ShouldEqual, createReq.Name)
				So(listResp.Result[0].NetworkServerID, ShouldEqual, n.ID)
			})

			Convey("Then List given no network-server ID lists all the gateway-profiles", func() {
				listResp, err := api.List(ctx, &pb.ListGatewayProfilesRequest{
					Limit: 10,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 1)
				So(listResp.Result, ShouldHaveLength, 1)
				So(listResp.Result[0].GatewayProfileID, ShouldEqual, createResp.GatewayProfileID)
				So(listResp.Result[0].Name, ShouldEqual, createReq.Name)
				So(listResp.Result[0].NetworkServerID, ShouldEqual, n.ID)
			})
		})
	})
}
