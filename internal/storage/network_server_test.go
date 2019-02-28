package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNetworkServer(t *testing.T) {
	conf := test.GetConfig()
	if err := Setup(conf); err != nil {
		t.Fatal(err)
	}

	Convey("Testing the Validate function", t, func() {
		testTable := []struct {
			NetworkServer NetworkServer
			ExpectedError error
		}{
			{
				NetworkServer: NetworkServer{
					GatewayDiscoveryEnabled:  false,
					GatewayDiscoveryInterval: 0,
				},
				ExpectedError: nil,
			},
			{
				NetworkServer: NetworkServer{
					GatewayDiscoveryEnabled:  true,
					GatewayDiscoveryInterval: 0,
				},
				ExpectedError: ErrInvalidGatewayDiscoveryInterval,
			},
		}

		for i, test := range testTable {
			Convey(fmt.Sprintf("Test %d", i), func() {
				So(test.NetworkServer.Validate(), ShouldEqual, test.ExpectedError)
			})
		}
	})

	Convey("Given a clean database with an organization", t, func() {
		test.MustResetDB(DB().DB)

		nsClient := mock.NewClient()
		networkserver.SetPool(mock.NewPool(nsClient))

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(db, &org), ShouldBeNil)

		rpID, err := uuid.FromString(config.C.ApplicationServer.ID)
		So(err, ShouldBeNil)

		Convey("Then CreateNetworkServer creates a network-server", func() {
			n := NetworkServer{
				Name:                        "test-ns",
				Server:                      "test-ns:123",
				CACert:                      "CACERT",
				TLSCert:                     "TLSCERT",
				TLSKey:                      "TLSKey",
				RoutingProfileCACert:        "RPCACERT",
				RoutingProfileTLSCert:       "RPTLSCERT",
				RoutingProfileTLSKey:        "RPTLSKEY",
				GatewayDiscoveryEnabled:     true,
				GatewayDiscoveryInterval:    5,
				GatewayDiscoveryTXFrequency: 868100000,
				GatewayDiscoveryDR:          5,
			}
			So(CreateNetworkServer(db, &n), ShouldBeNil)
			n.CreatedAt = n.CreatedAt.UTC().Truncate(time.Millisecond)
			n.UpdatedAt = n.UpdatedAt.UTC().Truncate(time.Millisecond)
			So(nsClient.CreateRoutingProfileChan, ShouldHaveLength, 1)
			So(<-nsClient.CreateRoutingProfileChan, ShouldResemble, ns.CreateRoutingProfileRequest{
				RoutingProfile: &ns.RoutingProfile{
					Id:      rpID.Bytes(),
					AsId:    config.C.ApplicationServer.API.PublicHost,
					CaCert:  "RPCACERT",
					TlsCert: "RPTLSCERT",
					TlsKey:  "RPTLSKEY",
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
				So(CreateOrganization(DB(), &org2), ShouldBeNil)

				sp := ServiceProfile{
					NetworkServerID: n.ID,
					OrganizationID:  org.ID,
					Name:            "test-service-profile",
				}
				So(CreateServiceProfile(DB(), &sp), ShouldBeNil)

				Convey("Then GetNetworkServerCountForOrganizationID returns the number of network-servers for the given organization", func() {
					count, err := GetNetworkServerCountForOrganizationID(DB(), org.ID)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 1)

					count, err = GetNetworkServerCountForOrganizationID(DB(), org2.ID)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)
				})

				Convey("Then GetNetworkServersForOrganizationID returns the network-servers for the given organization", func() {
					items, err := GetNetworkServersForOrganizationID(DB(), org.ID, 10, 0)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 1)

					items, err = GetNetworkServersForOrganizationID(DB(), org2.ID, 10, 0)
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
				n.GatewayDiscoveryEnabled = false
				n.GatewayDiscoveryInterval = 1
				n.GatewayDiscoveryTXFrequency = 868300000
				n.GatewayDiscoveryDR = 4
				So(UpdateNetworkServer(db, &n), ShouldBeNil)
				So(nsClient.UpdateRoutingProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.UpdateRoutingProfileChan, ShouldResemble, ns.UpdateRoutingProfileRequest{
					RoutingProfile: &ns.RoutingProfile{
						Id:      rpID.Bytes(),
						AsId:    config.C.ApplicationServer.API.PublicHost,
						CaCert:  "RPCACERT2",
						TlsCert: "RPTLSCERT2",
						TlsKey:  "RPTLSKEY2",
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
					Id: rpID.Bytes(),
				})

				_, err := GetNetworkServer(db, n.ID)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, ErrDoesNotExist)
			})
		})
	})
}
