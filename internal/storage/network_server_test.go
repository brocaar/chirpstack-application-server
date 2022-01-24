package storage

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/config"
)

func (ts *StorageTestSuite) TestNetworkServer() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	org := Organization{
		Name: "test-org",
	}
	assert.NoError(CreateOrganization(context.Background(), ts.Tx(), &org))

	rpID, err := uuid.FromString(config.C.ApplicationServer.ID)
	assert.NoError(err)

	ts.T().Run("Validate", func(t *testing.T) {
		testTable := []struct {
			NetworkServer NetworkServer
			ExpectedError error
		}{
			{
				NetworkServer: NetworkServer{
					Name:                     "test-ns",
					GatewayDiscoveryEnabled:  false,
					GatewayDiscoveryInterval: 5,
				},
				ExpectedError: nil,
			},
			{
				NetworkServer: NetworkServer{
					Name:                     "test-ns",
					GatewayDiscoveryEnabled:  true,
					GatewayDiscoveryInterval: 0,
				},
				ExpectedError: ErrInvalidGatewayDiscoveryInterval,
			},
		}

		for _, tst := range testTable {
			assert.Equal(tst.ExpectedError, tst.NetworkServer.Validate())
		}
	})

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

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
		assert.NoError(CreateNetworkServer(context.Background(), ts.Tx(), &n))
		n.CreatedAt = n.CreatedAt.UTC().Truncate(time.Millisecond)
		n.UpdatedAt = n.UpdatedAt.UTC().Truncate(time.Millisecond)

		assert.Equal(ns.CreateRoutingProfileRequest{
			RoutingProfile: &ns.RoutingProfile{
				Id:      rpID.Bytes(),
				AsId:    config.C.ApplicationServer.API.PublicHost,
				CaCert:  "RPCACERT",
				TlsCert: "RPTLSCERT",
				TlsKey:  "RPTLSKEY",
			},
		}, <-nsClient.CreateRoutingProfileChan)

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			nsGet, err := GetNetworkServer(context.Background(), ts.Tx(), n.ID)
			assert.NoError(err)
			nsGet.CreatedAt = nsGet.CreatedAt.UTC().Truncate(time.Millisecond)
			nsGet.UpdatedAt = nsGet.UpdatedAt.UTC().Truncate(time.Millisecond)

			assert.Equal(n, nsGet)
		})

		t.Run("List", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetNetworkServerCount(context.Background(), ts.Tx(), NetworkServerFilters{})
			assert.NoError(err)
			assert.Equal(1, count)

			items, err := GetNetworkServers(context.Background(), ts.Tx(), NetworkServerFilters{
				Limit: 10,
			})
			assert.NoError(err)
			assert.Len(items, 1)
		})

		t.Run("List by OrganizationID", func(t *testing.T) {
			assert := require.New(t)

			// not associated to org
			count, err := GetNetworkServerCount(context.Background(), ts.Tx(), NetworkServerFilters{
				OrganizationID: org.ID,
			})
			assert.NoError(err)
			assert.Equal(0, count)

			items, err := GetNetworkServers(context.Background(), ts.Tx(), NetworkServerFilters{
				OrganizationID: org.ID,
			})
			assert.NoError(err)
			assert.Len(items, 0)

			// associate ns to org using service-profile
			sp := ServiceProfile{
				Name:            "test-sp",
				OrganizationID:  org.ID,
				NetworkServerID: n.ID,
			}
			assert.NoError(CreateServiceProfile(context.Background(), ts.Tx(), &sp))

			// valid org id
			count, err = GetNetworkServerCount(context.Background(), ts.Tx(), NetworkServerFilters{
				OrganizationID: org.ID,
			})
			assert.NoError(err)
			assert.Equal(1, count)

			items, err = GetNetworkServers(context.Background(), ts.Tx(), NetworkServerFilters{
				OrganizationID: org.ID,
			})
			assert.NoError(err)
			assert.Len(items, 0)

			// delete sp
			var spID uuid.UUID
			copy(spID[:], sp.ServiceProfile.Id)
			assert.NoError(DeleteServiceProfile(context.Background(), ts.Tx(), spID))
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

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
			assert.NoError(UpdateNetworkServer(context.Background(), ts.Tx(), &n))
			n.UpdatedAt = n.UpdatedAt.UTC().Truncate(time.Millisecond)

			assert.Equal(ns.UpdateRoutingProfileRequest{
				RoutingProfile: &ns.RoutingProfile{
					Id:      rpID.Bytes(),
					AsId:    config.C.ApplicationServer.API.PublicHost,
					CaCert:  "RPCACERT2",
					TlsCert: "RPTLSCERT2",
					TlsKey:  "RPTLSKEY2",
				},
			}, <-nsClient.UpdateRoutingProfileChan)

			nsGet, err := GetNetworkServer(context.Background(), ts.Tx(), n.ID)
			assert.NoError(err)
			nsGet.CreatedAt = nsGet.CreatedAt.UTC().Truncate(time.Millisecond)
			nsGet.UpdatedAt = nsGet.UpdatedAt.UTC().Truncate(time.Millisecond)

			assert.Equal(n, nsGet)
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteNetworkServer(context.Background(), ts.Tx(), n.ID))
			assert.Equal(ns.DeleteRoutingProfileRequest{
				Id: rpID.Bytes(),
			}, <-nsClient.DeleteRoutingProfileChan)

			_, err := GetNetworkServer(context.Background(), db, n.ID)
			assert.Equal(ErrDoesNotExist, err)
		})
	})
}
