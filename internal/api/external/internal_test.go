package external

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

func (ts *APITestSuite) TestAPIKey() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	validator := &TestValidator{returnSubject: "user"}
	api := NewInternalAPI(validator)

	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))

	n := storage.NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:1234",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &n))

	sp := storage.ServiceProfile{
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp))
	spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
	assert.NoError(err)

	app := storage.Application{
		Name:             "test-app",
		Description:      "test-app",
		OrganizationID:   org.ID,
		ServiceProfileID: spID,
	}
	assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &app))

	ts.T().Run("Create", func(t *testing.T) {
		t.Run("Admin key", func(t *testing.T) {
			assert := require.New(t)

			resp, err := api.CreateAPIKey(context.Background(), &pb.CreateAPIKeyRequest{
				ApiKey: &pb.APIKey{
					Name:    "admin",
					IsAdmin: true,
				},
			})
			assert.NoError(err)
			assert.NotEqual("", resp.Id)
			assert.NotEqual("", resp.JwtToken)
		})

		t.Run("Organization key", func(t *testing.T) {
			assert := require.New(t)

			resp, err := api.CreateAPIKey(context.Background(), &pb.CreateAPIKeyRequest{
				ApiKey: &pb.APIKey{
					Name:           "organization",
					OrganizationId: org.ID,
				},
			})
			assert.NoError(err)
			assert.NotEqual("", resp.Id)
			assert.NotEqual("", resp.JwtToken)
		})

		t.Run("Application key", func(t *testing.T) {
			assert := require.New(t)

			resp, err := api.CreateAPIKey(context.Background(), &pb.CreateAPIKeyRequest{
				ApiKey: &pb.APIKey{
					Name:          "application",
					ApplicationId: app.ID,
				},
			})
			assert.NoError(err)
			assert.NotEqual("", resp.Id)
			assert.NotEqual("", resp.JwtToken)
		})

	})

	ts.T().Run("List", func(t *testing.T) {
		t.Run("Admin key", func(t *testing.T) {
			assert := require.New(t)

			resp, err := api.ListAPIKeys(context.Background(), &pb.ListAPIKeysRequest{
				Limit:   10,
				IsAdmin: true,
			})
			assert.NoError(err)
			assert.EqualValues(1, resp.TotalCount)
			assert.Equal(1, len(resp.Result))
			assert.Equal("admin", resp.Result[0].Name)
		})

		t.Run("Organization key", func(t *testing.T) {
			assert := require.New(t)

			resp, err := api.ListAPIKeys(context.Background(), &pb.ListAPIKeysRequest{
				Limit:          10,
				OrganizationId: org.ID,
			})
			assert.NoError(err)
			assert.EqualValues(1, resp.TotalCount)
			assert.Equal(1, len(resp.Result))
			assert.Equal("organization", resp.Result[0].Name)
		})

		t.Run("Application key", func(t *testing.T) {
			assert := require.New(t)

			resp, err := api.ListAPIKeys(context.Background(), &pb.ListAPIKeysRequest{
				Limit:         10,
				ApplicationId: app.ID,
			})
			assert.NoError(err)
			assert.EqualValues(1, resp.TotalCount)
			assert.Equal(1, len(resp.Result))
			assert.Equal("application", resp.Result[0].Name)
		})
	})

	ts.T().Run("Delete", func(t *testing.T) {
		assert := require.New(t)

		resp, err := api.ListAPIKeys(context.Background(), &pb.ListAPIKeysRequest{
			Limit:         10,
			ApplicationId: app.ID,
		})
		assert.NoError(err)
		assert.EqualValues(1, resp.TotalCount)

		_, err = api.DeleteAPIKey(context.Background(), &pb.DeleteAPIKeyRequest{
			Id: resp.Result[0].Id,
		})
		assert.NoError(err)

		_, err = api.DeleteAPIKey(context.Background(), &pb.DeleteAPIKeyRequest{
			Id: resp.Result[0].Id,
		})
		assert.Equal(codes.NotFound, grpc.Code(err))
	})
}
