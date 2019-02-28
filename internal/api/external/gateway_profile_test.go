package external

import (
	"testing"

	"github.com/brocaar/loraserver/api/common"
	"github.com/brocaar/loraserver/api/ns"
	uuid "github.com/gofrs/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
)

func TestGatewayProfileTest(t *testing.T) {
	conf := test.GetConfig()
	if err := storage.Setup(conf); err != nil {
		t.Fatal(err)
	}

	Convey("Given a clean database and api instance", t, func() {
		test.MustResetDB(storage.DB().DB)

		nsClient := mock.NewClient()
		networkserver.SetPool(mock.NewPool(nsClient))

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewGatewayProfileAPI(validator)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(storage.DB(), &n), ShouldBeNil)

		Convey("Then Create creates the gateway-profile", func() {
			createReq := pb.CreateGatewayProfileRequest{
				GatewayProfile: &pb.GatewayProfile{
					Name:            "test-gp",
					NetworkServerId: n.ID,
					Channels:        []uint32{0, 1, 2},
					ExtraChannels: []*pb.GatewayProfileExtraChannel{
						{
							Modulation:       common.Modulation_LORA,
							Frequency:        867100000,
							Bandwidth:        125,
							SpreadingFactors: []uint32{10, 11, 12},
						},
						{
							Modulation: common.Modulation_FSK,
							Frequency:  867300000,
							Bitrate:    50000,
						},
					},
				},
			}

			createResp, err := api.Create(ctx, &createReq)
			So(err, ShouldBeNil)
			So(createResp.Id, ShouldNotEqual, "")
			So(createResp.Id, ShouldNotEqual, uuid.Nil.String())
			So(nsClient.CreateGatewayProfileChan, ShouldHaveLength, 1)

			// set mock
			nsCreate := <-nsClient.CreateGatewayProfileChan
			nsClient.GetGatewayProfileResponse = ns.GetGatewayProfileResponse{
				GatewayProfile: nsCreate.GatewayProfile,
			}

			Convey("Then Get returns the gateway-profile", func() {
				getResp, err := api.Get(ctx, &pb.GetGatewayProfileRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)

				createReq.GatewayProfile.Id = createResp.Id
				So(getResp.GatewayProfile, ShouldResemble, createReq.GatewayProfile)
			})

			Convey("Then Update updates the gateway-profile", func() {
				updateReq := pb.UpdateGatewayProfileRequest{
					GatewayProfile: &pb.GatewayProfile{
						Id:              createResp.Id,
						NetworkServerId: n.ID,
						Name:            "updated-gp",
						Channels:        []uint32{1, 2},
						ExtraChannels: []*pb.GatewayProfileExtraChannel{
							{
								Modulation: common.Modulation_FSK,
								Frequency:  867300000,
								Bitrate:    50000,
							},
							{
								Modulation:       common.Modulation_LORA,
								Frequency:        867100000,
								Bandwidth:        125,
								SpreadingFactors: []uint32{10, 11, 12},
							},
						},
					},
				}

				_, err := api.Update(ctx, &updateReq)
				So(err, ShouldBeNil)
				So(nsClient.UpdateGatewayProfileChan, ShouldHaveLength, 1)

				// set mock
				nsUpdate := <-nsClient.UpdateGatewayProfileChan
				nsClient.GetGatewayProfileResponse = ns.GetGatewayProfileResponse{
					GatewayProfile: nsUpdate.GatewayProfile,
				}

				getResp, err := api.Get(ctx, &pb.GetGatewayProfileRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)
				So(getResp.GatewayProfile, ShouldResemble, updateReq.GatewayProfile)
			})

			Convey("Then Delete deletes the gateway-profile", func() {
				_, err := api.Delete(ctx, &pb.DeleteGatewayProfileRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)

				_, err = api.Get(ctx, &pb.GetGatewayProfileRequest{
					Id: createResp.Id,
				})
				So(err, ShouldNotBeNil)
				So(grpc.Code(err), ShouldEqual, codes.NotFound)
			})

			Convey("Then List given a network-server ID lists the gateway-profiles", func() {
				listResp, err := api.List(ctx, &pb.ListGatewayProfilesRequest{
					NetworkServerId: n.ID,
					Limit:           10,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 1)
				So(listResp.Result, ShouldHaveLength, 1)
				So(listResp.Result[0].Id, ShouldEqual, createResp.Id)
				So(listResp.Result[0].Name, ShouldEqual, createReq.GatewayProfile.Name)
				So(listResp.Result[0].NetworkServerId, ShouldEqual, n.ID)
			})

			Convey("Then List given no network-server ID lists all the gateway-profiles", func() {
				listResp, err := api.List(ctx, &pb.ListGatewayProfilesRequest{
					Limit: 10,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 1)
				So(listResp.Result, ShouldHaveLength, 1)
				So(listResp.Result[0].Id, ShouldEqual, createResp.Id)
				So(listResp.Result[0].Name, ShouldEqual, createReq.GatewayProfile.Name)
				So(listResp.Result[0].NetworkServerId, ShouldEqual, n.ID)
			})
		})
	})
}
