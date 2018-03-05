package storage

import (
	"testing"
	"time"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/test"
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
	config.C.PostgreSQL.DB = db

	Convey("Given a clean database with an organization", t, func() {
		test.MustResetDB(db)

		nsClient := test.NewNetworkServerClient()
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(db, &org), ShouldBeNil)

		Convey("Then CreateNetworkServer creates a network-server", func() {
			n := NetworkServer{
				Name:                  "test-ns",
				Server:                "test-ns:123",
				CACert:                "CACERT",
				TLSCert:               "TLSCERT",
				TLSKey:                "TLSKey",
				RoutingProfileCACert:  "RPCACERT",
				RoutingProfileTLSCert: "RPTLSCERT",
				RoutingProfileTLSKey:  "RPTLSKEY",
			}
			So(CreateNetworkServer(db, &n), ShouldBeNil)
			n.CreatedAt = n.CreatedAt.UTC().Truncate(time.Millisecond)
			n.UpdatedAt = n.UpdatedAt.UTC().Truncate(time.Millisecond)
			So(nsClient.CreateRoutingProfileChan, ShouldHaveLength, 1)
			So(<-nsClient.CreateRoutingProfileChan, ShouldResemble, ns.CreateRoutingProfileRequest{
				RoutingProfile: &ns.RoutingProfile{
					RoutingProfileID: config.C.ApplicationServer.ID,
					AsID:             config.C.ApplicationServer.API.PublicHost,
				},
				CaCert:  "RPCACERT",
				TlsCert: "RPTLSCERT",
				TlsKey:  "RPTLSKEY",
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
				So(CreateOrganization(config.C.PostgreSQL.DB, &org2), ShouldBeNil)

				sp := ServiceProfile{
					NetworkServerID: n.ID,
					OrganizationID:  org.ID,
					Name:            "test-service-profile",
					ServiceProfile:  backend.ServiceProfile{},
				}
				So(CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)

				Convey("Then GetNetworkServerCountForOrganizationID returns the number of network-servers for the given organization", func() {
					count, err := GetNetworkServerCountForOrganizationID(config.C.PostgreSQL.DB, org.ID)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 1)

					count, err = GetNetworkServerCountForOrganizationID(config.C.PostgreSQL.DB, org2.ID)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)
				})

				Convey("Then GetNetworkServersForOrganizationID returns the network-servers for the given organization", func() {
					items, err := GetNetworkServersForOrganizationID(config.C.PostgreSQL.DB, org.ID, 10, 0)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 1)

					items, err = GetNetworkServersForOrganizationID(config.C.PostgreSQL.DB, org2.ID, 10, 0)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 0)
				})
			})

			Convey("Then UpdateNetworkServer updates the network-server", func() {
				n.Name = "new-nw-server"
				n.Server = "new-nw-server:123"
				n.CACert = "CACERT2"
				n.TLSCert = "TLSCERT2"
				n.TLSKey = "TLSKey2"
				n.RoutingProfileCACert = "RPCACERT2"
				n.RoutingProfileTLSCert = "RPTLSCERT2"
				n.RoutingProfileTLSKey = "RPTLSKEY2"
				So(UpdateNetworkServer(db, &n), ShouldBeNil)
				So(nsClient.UpdateRoutingProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.UpdateRoutingProfileChan, ShouldResemble, ns.UpdateRoutingProfileRequest{
					RoutingProfile: &ns.RoutingProfile{
						RoutingProfileID: config.C.ApplicationServer.ID,
						AsID:             config.C.ApplicationServer.API.PublicHost,
					},
					CaCert:  "RPCACERT2",
					TlsCert: "RPTLSCERT2",
					TlsKey:  "RPTLSKEY2",
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
					RoutingProfileID: config.C.ApplicationServer.ID,
				})

				_, err := GetNetworkServer(db, n.ID)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, ErrDoesNotExist)
			})
		})
	})
}
