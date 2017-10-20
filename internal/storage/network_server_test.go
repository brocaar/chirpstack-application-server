package storage

import (
	"testing"
	"time"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan/backend"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNetworkServer(t *testing.T) {
	conf := test.GetConfig()
	db, err := OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	common.DB = db

	Convey("Given a clean database with an organization", t, func() {
		test.MustResetDB(db)

		nsClient := test.NewNetworkServerClient()
		common.NetworkServerPool = test.NewNetworkServerPool(nsClient)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(db, &org), ShouldBeNil)

		Convey("Then CreateNetworkServer creates a network-server", func() {
			n := NetworkServer{
				Name:   "test-ns",
				Server: "test-ns:123",
			}
			So(CreateNetworkServer(db, &n), ShouldBeNil)
			n.CreatedAt = n.CreatedAt.UTC().Truncate(time.Millisecond)
			n.UpdatedAt = n.UpdatedAt.UTC().Truncate(time.Millisecond)
			So(nsClient.CreateRoutingProfileChan, ShouldHaveLength, 1)
			So(<-nsClient.CreateRoutingProfileChan, ShouldResemble, ns.CreateRoutingProfileRequest{
				RoutingProfile: &ns.RoutingProfile{
					RoutingProfileID: common.ApplicationServerID,
					AsID:             common.ApplicationServerServer,
				},
			})

			Convey("Then GetNetworkServer returns the network-server", func() {
				nsGet, err := GetNetworkServer(db, n.ID)
				So(err, ShouldBeNil)
				nsGet.CreatedAt = nsGet.CreatedAt.UTC().Truncate(time.Millisecond)
				nsGet.UpdatedAt = nsGet.UpdatedAt.UTC().Truncate(time.Millisecond)

				So(nsGet, ShouldResemble, n)
			})

			Convey("Then GetNetworkServerCount returns 1", func() {
				count, err := GetNetworkServerCount(db)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then GetNetworkServers returns a single item", func() {
				items, err := GetNetworkServers(db, 10, 0)
				So(err, ShouldBeNil)
				So(items, ShouldHaveLength, 1)
			})

			Convey("Given a second organization and a service-profile attached to the first organization", func() {
				org2 := Organization{
					Name: "test-org-2",
				}
				So(CreateOrganization(common.DB, &org2), ShouldBeNil)

				sp := ServiceProfile{
					NetworkServerID: n.ID,
					OrganizationID:  org.ID,
					Name:            "test-service-profile",
					ServiceProfile:  backend.ServiceProfile{},
				}
				So(CreateServiceProfile(common.DB, &sp), ShouldBeNil)

				Convey("Then GetNetworkServerCountForOrganizationID returns the number of network-servers for the given organization", func() {
					count, err := GetNetworkServerCountForOrganizationID(common.DB, org.ID)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 1)

					count, err = GetNetworkServerCountForOrganizationID(common.DB, org2.ID)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)
				})

				Convey("Then GetNetworkServersForOrganizationID returns the network-servers for the given organization", func() {
					items, err := GetNetworkServersForOrganizationID(common.DB, org.ID, 10, 0)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 1)

					items, err = GetNetworkServersForOrganizationID(common.DB, org2.ID, 10, 0)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 0)
				})
			})

			Convey("Then UpdateNetworkServer updates the network-server", func() {
				n.Name = "new-nw-server"
				n.Server = "new-nw-server:123"
				So(UpdateNetworkServer(db, &n), ShouldBeNil)
				So(nsClient.UpdateRoutingProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.UpdateRoutingProfileChan, ShouldResemble, ns.UpdateRoutingProfileRequest{
					RoutingProfile: &ns.RoutingProfile{
						RoutingProfileID: common.ApplicationServerID,
						AsID:             common.ApplicationServerServer,
					},
				})

				n.UpdatedAt = n.UpdatedAt.UTC().Truncate(time.Millisecond)

				nsGet, err := GetNetworkServer(db, n.ID)
				So(err, ShouldBeNil)
				nsGet.CreatedAt = nsGet.CreatedAt.UTC().Truncate(time.Millisecond)
				nsGet.UpdatedAt = nsGet.UpdatedAt.UTC().Truncate(time.Millisecond)

				So(nsGet, ShouldResemble, n)
			})

			Convey("Then DeleteNetworkServer deletes the network-server", func() {
				So(DeleteNetworkServer(db, n.ID), ShouldBeNil)
				So(nsClient.DeleteRoutingProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.DeleteRoutingProfileChan, ShouldResemble, ns.DeleteRoutingProfileRequest{
					RoutingProfileID: common.ApplicationServerID,
				})

				_, err := GetNetworkServer(db, n.ID)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, ErrDoesNotExist)
			})
		})
	})
}
