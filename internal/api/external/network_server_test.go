package external

import (
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

func TestNetworkServerAPI(t *testing.T) {
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
		api := NewNetworkServerAPI(validator)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(storage.DB(), &org), ShouldBeNil)

		Convey("Then Create creates a network-server", func() {
			createReq := pb.CreateNetworkServerRequest{
				NetworkServer: &pb.NetworkServer{
					Name:                        "test ns",
					Server:                      "test-ns:1234",
					CaCert:                      "CACERT",
					TlsCert:                     "TLSCERT",
					TlsKey:                      "TLSKEY",
					RoutingProfileCaCert:        "RPCACERT",
					RoutingProfileTlsCert:       "RPTLSCERT",
					RoutingProfileTlsKey:        "RPTLSKEY",
					GatewayDiscoveryEnabled:     true,
					GatewayDiscoveryInterval:    5,
					GatewayDiscoveryTxFrequency: 868100000,
					GatewayDiscoveryDr:          5,
				},
			}

			resp, err := api.Create(ctx, &createReq)
			So(err, ShouldBeNil)
			So(resp.Id, ShouldBeGreaterThan, 0)

			Convey("Then Get returns the network-server", func() {
				getResp, err := api.Get(ctx, &pb.GetNetworkServerRequest{
					Id: resp.Id,
				})
				So(err, ShouldBeNil)

				createReq.NetworkServer.Id = resp.Id
				createReq.NetworkServer.TlsKey = "" // key is not returned on get
				createReq.NetworkServer.RoutingProfileTlsKey = ""

				So(getResp.NetworkServer, ShouldResemble, createReq.NetworkServer)
			})

			Convey("Then the CA and TLS fields are populated", func() {
				n, err := storage.GetNetworkServer(storage.DB(), resp.Id)
				So(err, ShouldBeNil)
				So(n.CACert, ShouldEqual, "CACERT")
				So(n.TLSCert, ShouldEqual, "TLSCERT")
				So(n.TLSKey, ShouldEqual, "TLSKEY")
				So(n.RoutingProfileCACert, ShouldEqual, "RPCACERT")
				So(n.RoutingProfileTLSCert, ShouldEqual, "RPTLSCERT")
				So(n.RoutingProfileTLSKey, ShouldEqual, "RPTLSKEY")
			})

			Convey("Then Update updates the network-server", func() {
				updateReq := pb.UpdateNetworkServerRequest{
					NetworkServer: &pb.NetworkServer{
						Id:                          resp.Id,
						Name:                        "updated-test-ns",
						Server:                      "updated-test-ns:1234",
						CaCert:                      "CACERT2",
						TlsCert:                     "TLSCERT2",
						TlsKey:                      "TLSKEY2",
						RoutingProfileCaCert:        "RPCACERT2",
						RoutingProfileTlsCert:       "RPTLSCERT2",
						RoutingProfileTlsKey:        "RPTLSKEY2",
						GatewayDiscoveryEnabled:     false,
						GatewayDiscoveryInterval:    1,
						GatewayDiscoveryTxFrequency: 868300000,
						GatewayDiscoveryDr:          4,
					},
				}

				_, err := api.Update(ctx, &updateReq)
				So(err, ShouldBeNil)

				getResp, err := api.Get(ctx, &pb.GetNetworkServerRequest{
					Id: resp.Id,
				})
				So(err, ShouldBeNil)

				updateReq.NetworkServer.TlsKey = "" // is not returned on get
				updateReq.NetworkServer.RoutingProfileTlsKey = ""
				So(getResp.NetworkServer, ShouldResemble, updateReq.NetworkServer)

				Convey("Then the network-server is updated", func() {
					n, err := storage.GetNetworkServer(storage.DB(), resp.Id)
					So(err, ShouldBeNil)
					So(n.CACert, ShouldEqual, "CACERT2")
					So(n.TLSCert, ShouldEqual, "TLSCERT2")
					So(n.TLSKey, ShouldEqual, "TLSKEY2")
					So(n.RoutingProfileCACert, ShouldEqual, "RPCACERT2")
					So(n.RoutingProfileTLSCert, ShouldEqual, "RPTLSCERT2")
					So(n.RoutingProfileTLSKey, ShouldEqual, "RPTLSKEY2")
					So(n.GatewayDiscoveryEnabled, ShouldBeFalse)
					So(n.GatewayDiscoveryInterval, ShouldEqual, 1)
					So(n.GatewayDiscoveryTXFrequency, ShouldEqual, 868300000)
					So(n.GatewayDiscoveryDR, ShouldEqual, 4)
				})
			})

			Convey("Then Delete deletes the network-server", func() {
				_, err := api.Delete(ctx, &pb.DeleteNetworkServerRequest{
					Id: resp.Id,
				})
				So(err, ShouldBeNil)

				_, err = api.Get(ctx, &pb.GetNetworkServerRequest{
					Id: resp.Id,
				})
				So(err, ShouldNotBeNil)
				So(grpc.Code(err), ShouldEqual, codes.NotFound)
			})

			Convey("Then List lists the network-servers", func() {
				// non admin returns nothing
				listResp, err := api.List(ctx, &pb.ListNetworkServerRequest{
					Limit:  10,
					Offset: 0,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 0)
				So(listResp.Result, ShouldHaveLength, 0)

				validator.returnIsAdmin = true
				listResp, err = api.List(ctx, &pb.ListNetworkServerRequest{
					Limit:  10,
					Offset: 0,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 1)
				So(listResp.Result, ShouldHaveLength, 1)
				So(listResp.Result[0].Name, ShouldEqual, "test ns")
				So(listResp.Result[0].Server, ShouldEqual, "test-ns:1234")
			})

			Convey("Given a second organization and service-profile assigned to the first organization", func() {
				org2 := storage.Organization{
					Name: "test-org-2",
				}
				So(storage.CreateOrganization(storage.DB(), &org2), ShouldBeNil)

				sp := storage.ServiceProfile{
					NetworkServerID: resp.Id,
					OrganizationID:  org.ID,
					Name:            "test-sp",
				}
				So(storage.CreateServiceProfile(storage.DB(), &sp), ShouldBeNil)

				Convey("Then List with organization id lists the network-servers for the given organization id", func() {
					listResp, err := api.List(ctx, &pb.ListNetworkServerRequest{
						Limit:          10,
						OrganizationId: org.ID,
					})
					So(err, ShouldBeNil)
					So(listResp.TotalCount, ShouldEqual, 1)
					So(listResp.Result, ShouldHaveLength, 1)

					listResp, err = api.List(ctx, &pb.ListNetworkServerRequest{
						Limit:          10,
						OrganizationId: org2.ID,
					})
					So(err, ShouldBeNil)
					So(listResp.TotalCount, ShouldEqual, 0)
					So(listResp.Result, ShouldHaveLength, 0)
				})
			})
		})
	})
}
