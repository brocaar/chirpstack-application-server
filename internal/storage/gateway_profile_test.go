package storage

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/brocaar/loraserver/api/common"
	"github.com/brocaar/loraserver/api/ns"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/test"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGatewayProfile(t *testing.T) {
	conf := test.GetConfig()
	db, err := OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}

	config.C.PostgreSQL.DB = db
	nsClient := test.NewNetworkServerClient()
	config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

	Convey("Given a clean database with network-server", t, func() {
		test.MustResetDB(db)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(db, &n), ShouldBeNil)

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
			So(CreateGatewayProfile(db, &gp), ShouldBeNil)
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

				gpGet, err := GetGatewayProfile(db, gpID)
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

				So(UpdateGatewayProfile(db, &gp), ShouldBeNil)
				gp.UpdatedAt = gp.UpdatedAt.UTC().Truncate(time.Millisecond)

				So(nsClient.UpdateGatewayProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.UpdateGatewayProfileChan, ShouldResemble, ns.UpdateGatewayProfileRequest{
					GatewayProfile: &gp.GatewayProfile,
				})

				gpGet, err := GetGatewayProfile(db, gpID)
				So(err, ShouldBeNil)
				So(gpGet.Name, ShouldEqual, "updated-gateway-profile")
			})

			Convey("Then DeleteGatewayProfile deletes the gateway-profile", func() {
				So(DeleteGatewayProfile(db, gpID), ShouldBeNil)
				So(nsClient.DeleteGatewayProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.DeleteGatewayProfileChan, ShouldResemble, ns.DeleteGatewayProfileRequest{
					Id: gp.GatewayProfile.Id,
				})

				_, err := GetGatewayProfile(db, gpID)
				So(err, ShouldEqual, ErrDoesNotExist)
			})

			Convey("Then GetGatewayProfileCount returns the number of gateway profiles", func() {
				count, err := GetGatewayProfileCount(db)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then GetGatewayProfiles returns the gateway profiles", func() {
				gps, err := GetGatewayProfiles(db, 10, 0)
				So(err, ShouldBeNil)
				So(gps, ShouldHaveLength, 1)
				So(gps[0].GatewayProfileID, ShouldEqual, gpID)
				So(gps[0].NetworkServerID, ShouldEqual, gp.NetworkServerID)
				So(gps[0].NetworkServerName, ShouldEqual, n.Name)
				So(gps[0].Name, ShouldEqual, gp.Name)
			})

			Convey("Then GetGatewayProfileCountForNetworkServerID returns the number of gateway profiles", func() {
				count, err := GetGatewayProfileCountForNetworkServerID(db, n.ID)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then GetGatewayProfilesForNetworkServerID returns the gateway profiles", func() {
				gps, err := GetGatewayProfilesForNetworkServerID(db, n.ID, 10, 0)
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
