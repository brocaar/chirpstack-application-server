package external

import (
	"testing"
	"time"

	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/ns"
	uuid "github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

func (ts *APITestSuite) TestGatewayProfile() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	validator := &TestValidator{}
	api := NewGatewayProfileAPI(validator)

	n := storage.NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:1234",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &n))

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		createReq := pb.CreateGatewayProfileRequest{
			GatewayProfile: &pb.GatewayProfile{
				Name:            "test-gp",
				NetworkServerId: n.ID,
				Channels:        []uint32{0, 1, 2},
				StatsInterval:   ptypes.DurationProto(time.Second * 30),
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

		createResp, err := api.Create(context.Background(), &createReq)
		assert.NoError(err)

		assert.NotEqual("", createResp.Id)
		assert.NotEqual(uuid.Nil.String(), createResp.Id)

		// set mock
		nsCreate := <-nsClient.CreateGatewayProfileChan
		nsClient.GetGatewayProfileResponse = ns.GetGatewayProfileResponse{
			GatewayProfile: nsCreate.GatewayProfile,
		}

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			getResp, err := api.Get(context.Background(), &pb.GetGatewayProfileRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			createReq.GatewayProfile.Id = createResp.Id
			assert.Equal(createReq.GatewayProfile, getResp.GatewayProfile)
		})

		t.Run("List with network-server id", func(t *testing.T) {
			assert := require.New(t)

			listResp, err := api.List(context.Background(), &pb.ListGatewayProfilesRequest{
				NetworkServerId: n.ID,
				Limit:           10,
			})
			assert.NoError(err)

			assert.EqualValues(1, listResp.TotalCount)
			assert.Len(listResp.Result, 1)
			assert.Equal(createResp.Id, listResp.Result[0].Id)
			assert.Equal(createReq.GatewayProfile.Name, listResp.Result[0].Name)
			assert.Equal(n.ID, listResp.Result[0].NetworkServerId)
		})

		t.Run("List", func(t *testing.T) {
			assert := require.New(t)

			listResp, err := api.List(context.Background(), &pb.ListGatewayProfilesRequest{
				Limit: 10,
			})
			assert.NoError(err)

			assert.EqualValues(1, listResp.TotalCount)
			assert.Len(listResp.Result, 1)
			assert.Equal(createResp.Id, listResp.Result[0].Id)
			assert.Equal(createReq.GatewayProfile.Name, listResp.Result[0].Name)
			assert.Equal(n.ID, listResp.Result[0].NetworkServerId)
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			updateReq := pb.UpdateGatewayProfileRequest{
				GatewayProfile: &pb.GatewayProfile{
					Id:              createResp.Id,
					NetworkServerId: n.ID,
					Name:            "updated-gp",
					Channels:        []uint32{1, 2},
					StatsInterval:   ptypes.DurationProto(time.Minute * 30),
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

			_, err := api.Update(context.Background(), &updateReq)
			assert.NoError(err)

			// set mock
			nsUpdate := <-nsClient.UpdateGatewayProfileChan
			nsClient.GetGatewayProfileResponse = ns.GetGatewayProfileResponse{
				GatewayProfile: nsUpdate.GatewayProfile,
			}

			getResp, err := api.Get(context.Background(), &pb.GetGatewayProfileRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			assert.Equal(updateReq.GatewayProfile, getResp.GatewayProfile)
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			_, err := api.Delete(context.Background(), &pb.DeleteGatewayProfileRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			_, err = api.Get(context.Background(), &pb.GetGatewayProfileRequest{
				Id: createResp.Id,
			})
			assert.Equal(codes.NotFound, grpc.Code(err))
		})
	})
}
