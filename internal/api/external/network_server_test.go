package external

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

func (ts *APITestSuite) TestNetworkServer() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	validator := &TestValidator{}
	api := NewNetworkServerAPI(validator)

	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))

	adminUser := storage.User{
		Email:    "admin@user.com",
		IsActive: true,
		IsAdmin:  true,
	}
	assert.NoError(storage.CreateUser(context.Background(), storage.DB(), &adminUser))

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

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

		resp, err := api.Create(context.Background(), &createReq)
		assert.NoError(err)
		assert.True(resp.Id > 0)

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			getResp, err := api.Get(context.Background(), &pb.GetNetworkServerRequest{
				Id: resp.Id,
			})
			assert.NoError(err)

			createReq.NetworkServer.Id = resp.Id
			createReq.NetworkServer.TlsKey = "" // key is not returned on get
			createReq.NetworkServer.RoutingProfileTlsKey = ""

			assert.Equal(createReq.NetworkServer, getResp.NetworkServer)
		})

		t.Run("CA and TLS fields are populated", func(t *testing.T) {
			assert := require.New(t)

			n, err := storage.GetNetworkServer(context.Background(), storage.DB(), resp.Id)
			assert.NoError(err)

			assert.Equal("CACERT", n.CACert)
			assert.Equal("TLSCERT", n.TLSCert)
			assert.Equal("TLSKEY", n.TLSKey)
			assert.Equal("RPCACERT", n.RoutingProfileCACert)
			assert.Equal("RPTLSCERT", n.RoutingProfileTLSCert)
			assert.Equal("RPTLSKEY", n.RoutingProfileTLSKey)
		})

		t.Run("List", func(t *testing.T) {
			assert := require.New(t)
			validator.returnUser = adminUser

			listResp, err := api.List(context.Background(), &pb.ListNetworkServerRequest{
				Limit:  10,
				Offset: 0,
			})
			assert.NoError(err)

			assert.EqualValues(1, listResp.TotalCount)
			assert.Len(listResp.Result, 1)
			assert.Equal("test ns", listResp.Result[0].Name)
			assert.Equal("test-ns:1234", listResp.Result[0].Server)
		})

		t.Run("Second organization and service-profile assigned to the first organization", func(t *testing.T) {
			assert := require.New(t)

			org2 := storage.Organization{
				Name: "test-org-2",
			}
			assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org2))

			sp := storage.ServiceProfile{
				NetworkServerID: resp.Id,
				OrganizationID:  org.ID,
				Name:            "test-sp",
			}
			assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp))

			t.Run("List with organization id filter", func(t *testing.T) {
				assert := require.New(t)

				listResp, err := api.List(context.Background(), &pb.ListNetworkServerRequest{
					Limit:          10,
					OrganizationId: org.ID,
				})
				assert.NoError(err)

				assert.EqualValues(1, listResp.TotalCount)
				assert.Len(listResp.Result, 1)

				listResp, err = api.List(context.Background(), &pb.ListNetworkServerRequest{
					Limit:          10,
					OrganizationId: org2.ID,
				})
				assert.NoError(err)

				assert.EqualValues(0, listResp.TotalCount)
				assert.Len(listResp.Result, 0)
			})

			assert.NoError(storage.DeleteOrganization(context.Background(), storage.DB(), org2.ID))
			var spID uuid.UUID
			copy(spID[:], sp.ServiceProfile.Id)
			assert.NoError(storage.DeleteServiceProfile(context.Background(), storage.DB(), spID))
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

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

			_, err := api.Update(context.Background(), &updateReq)
			assert.NoError(err)

			getResp, err := api.Get(context.Background(), &pb.GetNetworkServerRequest{
				Id: resp.Id,
			})
			assert.NoError(err)

			updateReq.NetworkServer.TlsKey = "" // is not returned on get
			updateReq.NetworkServer.RoutingProfileTlsKey = ""
			assert.Equal(updateReq.NetworkServer, getResp.NetworkServer)

			n, err := storage.GetNetworkServer(context.Background(), storage.DB(), resp.Id)
			assert.NoError(err)

			assert.Equal("CACERT2", n.CACert)
			assert.Equal("TLSCERT2", n.TLSCert)
			assert.Equal("TLSKEY2", n.TLSKey)
			assert.Equal("RPCACERT2", n.RoutingProfileCACert)
			assert.Equal("RPTLSCERT2", n.RoutingProfileTLSCert)
			assert.Equal("RPTLSKEY2", n.RoutingProfileTLSKey)
			assert.False(n.GatewayDiscoveryEnabled)
			assert.Equal(1, n.GatewayDiscoveryInterval)
			assert.Equal(868300000, n.GatewayDiscoveryTXFrequency)
			assert.Equal(4, n.GatewayDiscoveryDR)
		})

		t.Run("GetADRAlgorithms", func(t *testing.T) {
			assert := require.New(t)

			nsClient.GetADRAlgorithmsResponse = ns.GetADRAlgorithmsResponse{
				AdrAlgorithms: []*ns.ADRAlgorithm{
					{
						Id:   "default",
						Name: "Default ADR algorithm",
					},
				},
			}

			resp, err := api.GetADRAlgorithms(context.Background(), &pb.GetADRAlgorithmsRequest{
				NetworkServerId: resp.Id,
			})
			assert.NoError(err)
			assert.Equal(&pb.GetADRAlgorithmsResponse{
				AdrAlgorithms: []*pb.ADRAlgorithm{
					{
						Id:   "default",
						Name: "Default ADR algorithm",
					},
				},
			}, resp)
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			_, err := api.Delete(context.Background(), &pb.DeleteNetworkServerRequest{
				Id: resp.Id,
			})
			assert.NoError(err)

			_, err = api.Get(context.Background(), &pb.GetNetworkServerRequest{
				Id: resp.Id,
			})
			assert.Equal(codes.NotFound, grpc.Code(err))
		})
	})

}
