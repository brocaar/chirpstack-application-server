package storage

import (
	"context"
	"testing"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
)

func (ts *StorageTestSuite) TestGatewayProfile() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	n := NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:1234",
	}
	assert.NoError(CreateNetworkServer(context.Background(), ts.Tx(), &n))

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		gp := GatewayProfile{
			NetworkServerID: n.ID,
			Name:            "test-gateway-profile",
			GatewayProfile: ns.GatewayProfile{
				Channels:      []uint32{0, 1, 2},
				StatsInterval: ptypes.DurationProto(time.Second * 30),
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
		assert.NoError(CreateGatewayProfile(context.Background(), ts.Tx(), &gp))
		gp.CreatedAt = gp.CreatedAt.UTC().Truncate(time.Millisecond)
		gp.UpdatedAt = gp.UpdatedAt.UTC().Truncate(time.Millisecond)

		gpID, err := uuid.FromBytes(gp.GatewayProfile.Id)
		assert.NoError(err)

		assert.Equal(ns.CreateGatewayProfileRequest{
			GatewayProfile: &gp.GatewayProfile,
		}, <-nsClient.CreateGatewayProfileChan)

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			nsClient.GetGatewayProfileResponse = ns.GetGatewayProfileResponse{
				GatewayProfile: &gp.GatewayProfile,
			}

			gpGet, err := GetGatewayProfile(context.Background(), ts.Tx(), gpID)
			assert.NoError(err)
			gpGet.CreatedAt = gpGet.CreatedAt.UTC().Truncate(time.Millisecond)
			gpGet.UpdatedAt = gpGet.UpdatedAt.UTC().Truncate(time.Millisecond)

			assert.Equal(gp, gpGet)
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			gp.Name = "updated-gateway-profile"
			gp.GatewayProfile = ns.GatewayProfile{
				Id:            gp.GatewayProfile.Id,
				Channels:      []uint32{0, 1},
				StatsInterval: ptypes.DurationProto(time.Minute * 30),
				ExtraChannels: []*ns.GatewayProfileExtraChannel{
					{
						Modulation:       common.Modulation_LORA,
						Frequency:        867300000,
						SpreadingFactors: []uint32{9, 10, 11, 12},
						Bandwidth:        250,
					},
				},
			}

			assert.NoError(UpdateGatewayProfile(context.Background(), ts.Tx(), &gp))
			gp.UpdatedAt = gp.UpdatedAt.UTC().Truncate(time.Millisecond)

			assert.Equal(ns.UpdateGatewayProfileRequest{
				GatewayProfile: &gp.GatewayProfile,
			}, <-nsClient.UpdateGatewayProfileChan)

			gpGet, err := GetGatewayProfile(context.Background(), ts.Tx(), gpID)
			assert.NoError(err)
			assert.Equal("updated-gateway-profile", gpGet.Name)
		})

		t.Run("List", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetGatewayProfileCount(context.Background(), ts.Tx())
			assert.NoError(err)
			assert.Equal(1, count)

			gps, err := GetGatewayProfiles(context.Background(), ts.Tx(), 10, 0)
			assert.NoError(err)

			assert.Len(gps, 1)
			assert.Equal(gpID, gps[0].GatewayProfileID)
			assert.Equal(gp.NetworkServerID, gps[0].NetworkServerID)
			assert.Equal(n.Name, gps[0].NetworkServerName)
			assert.Equal(gp.Name, gps[0].Name)
		})

		t.Run("List by NetworkServerID", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetGatewayProfileCountForNetworkServerID(context.Background(), ts.Tx(), n.ID)
			assert.NoError(err)
			assert.Equal(1, count)

			gps, err := GetGatewayProfilesForNetworkServerID(context.Background(), ts.Tx(), n.ID, 10, 0)
			assert.Len(gps, 1)
			assert.Equal(gpID, gps[0].GatewayProfileID)
			assert.Equal(gp.NetworkServerID, gps[0].NetworkServerID)
			assert.Equal(n.Name, gps[0].NetworkServerName)
			assert.Equal(gp.Name, gps[0].Name)
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteGatewayProfile(context.Background(), ts.Tx(), gpID))
			assert.Equal(<-nsClient.DeleteGatewayProfileChan, ns.DeleteGatewayProfileRequest{
				Id: gp.GatewayProfile.Id,
			})

			_, err := GetGatewayProfile(context.Background(), ts.Tx(), gpID)
			assert.Equal(ErrDoesNotExist, err)
		})
	})
}
