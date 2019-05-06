package storage

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/common"
	"github.com/brocaar/loraserver/api/ns"
)

func TestGatewayProfile(t *testing.T) {
	conf := test.GetConfig()
	if err := Setup(conf); err != nil {
		t.Fatal(err)
	}
	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	Convey("Given a clean database with network-server", t, func() {
		test.MustResetDB(DB().DB)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(DB(), &n), ShouldBeNil)

		Convey("Then CreateGatewayProfile creates the gateway-profile", func() {
			gp := GatewayProfile{
				NetworkServerID: n.ID,
				Name:            "test-gateway-profile",
				GatewayProfile: ns.GatewayProfile{
					Channels: []uint32{0, 1, 2},
					ExtraChannels: []*ns.GatewayProfileExtraChannel{
						{
							Modulation:       common.Modulation_LORA,
							Frequency:        867100000,
							SpreadingFactors: []uint32{10, 11, 12},
							Bandwidth:        125,
						},
					},
				},
			}
			So(CreateGatewayProfile(DB(), &gp), ShouldBeNil)
			gp.CreatedAt = gp.CreatedAt.UTC().Truncate(time.Millisecond)
			gp.UpdatedAt = gp.UpdatedAt.UTC().Truncate(time.Millisecond)
			gpID, err := uuid.FromBytes(gp.GatewayProfile.Id)
			So(err, ShouldBeNil)

			So(nsClient.CreateGatewayProfileChan, ShouldHaveLength, 1)
			So(<-nsClient.CreateGatewayProfileChan, ShouldResemble, ns.CreateGatewayProfileRequest{
				GatewayProfile: &gp.GatewayProfile,
			})

			Convey("Then GetGatewayProfile reuturns the gateway-profile", func() {
				nsClient.GetGatewayProfileResponse = ns.GetGatewayProfileResponse{
					GatewayProfile: &gp.GatewayProfile,
				}

				gpGet, err := GetGatewayProfile(DB(), gpID)
				So(err, ShouldBeNil)
				gpGet.CreatedAt = gpGet.CreatedAt.UTC().Truncate(time.Millisecond)
				gpGet.UpdatedAt = gpGet.UpdatedAt.UTC().Truncate(time.Millisecond)

				So(gpGet, ShouldResemble, gp)
			})

			Convey("Then UpdateGatewayProfile updates the gateway-profile", func() {
				gp.Name = "updated-gateway-profile"
				gp.GatewayProfile = ns.GatewayProfile{
					Id:       gp.GatewayProfile.Id,
					Channels: []uint32{0, 1},
					ExtraChannels: []*ns.GatewayProfileExtraChannel{
						{
							Modulation:       common.Modulation_LORA,
							Frequency:        867300000,
							SpreadingFactors: []uint32{9, 10, 11, 12},
							Bandwidth:        250,
						},
					},
				}

				So(UpdateGatewayProfile(DB(), &gp), ShouldBeNil)
				gp.UpdatedAt = gp.UpdatedAt.UTC().Truncate(time.Millisecond)

				So(nsClient.UpdateGatewayProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.UpdateGatewayProfileChan, ShouldResemble, ns.UpdateGatewayProfileRequest{
					GatewayProfile: &gp.GatewayProfile,
				})

				gpGet, err := GetGatewayProfile(DB(), gpID)
				So(err, ShouldBeNil)
				So(gpGet.Name, ShouldEqual, "updated-gateway-profile")
			})

			Convey("Then DeleteGatewayProfile deletes the gateway-profile", func() {
				So(DeleteGatewayProfile(DB(), gpID), ShouldBeNil)
				So(nsClient.DeleteGatewayProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.DeleteGatewayProfileChan, ShouldResemble, ns.DeleteGatewayProfileRequest{
					Id: gp.GatewayProfile.Id,
				})

				_, err := GetGatewayProfile(DB(), gpID)
				So(err, ShouldEqual, ErrDoesNotExist)
			})

			Convey("Then GetGatewayProfileCount returns the number of gateway profiles", func() {
				count, err := GetGatewayProfileCount(db)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then GetGatewayProfiles returns the gateway profiles", func() {
				gps, err := GetGatewayProfiles(DB(), 10, 0)
				So(err, ShouldBeNil)
				So(gps, ShouldHaveLength, 1)
				So(gps[0].GatewayProfileID, ShouldEqual, gpID)
				So(gps[0].NetworkServerID, ShouldEqual, gp.NetworkServerID)
				So(gps[0].NetworkServerName, ShouldEqual, n.Name)
				So(gps[0].Name, ShouldEqual, gp.Name)
			})

			Convey("Then GetGatewayProfileCountForNetworkServerID returns the number of gateway profiles", func() {
				count, err := GetGatewayProfileCountForNetworkServerID(DB(), n.ID)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then GetGatewayProfilesForNetworkServerID returns the gateway profiles", func() {
				gps, err := GetGatewayProfilesForNetworkServerID(DB(), n.ID, 10, 0)
				So(err, ShouldBeNil)
				So(gps, ShouldHaveLength, 1)
				So(gps[0].GatewayProfileID, ShouldEqual, gpID)
				So(gps[0].NetworkServerID, ShouldEqual, gp.NetworkServerID)
				So(gps[0].NetworkServerName, ShouldEqual, n.Name)
				So(gps[0].Name, ShouldEqual, gp.Name)
			})
		})
	})
}
