package external

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/api/external/oidc"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

func (ts *APITestSuite) TestInternal() {
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
		Name:            "test-sp",
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

	ts.T().Run("APIKey", func(t *testing.T) {
		t.Run("Create", func(t *testing.T) {
			t.Run("Invalid", func(t *testing.T) {
				assert := require.New(t)

				_, err := api.CreateAPIKey(context.Background(), &pb.CreateAPIKeyRequest{
					ApiKey: &pb.APIKey{
						Name: "invalid",
					},
				})
				assert.Equal(codes.InvalidArgument, grpc.Code(err))
			})
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
	})

	ts.T().Run("Settings", func(t *testing.T) {
		assert := require.New(ts.T())

		brandingRegistration = "branding-reg"
		brandingFooter = "branding-foot"
		openIDConnectEnabled = true
		openIDLoginLabel = "login label"
		logoutURL = "http://logout.com"

		validator := &TestValidator{returnSubject: "user"}
		api := NewInternalAPI(validator)

		resp, err := api.Settings(context.Background(), nil)
		assert.NoError(err)
		assert.Equal(&pb.SettingsResponse{
			Branding: &pb.Branding{
				Registration: "branding-reg",
				Footer:       "branding-foot",
			},
			OpenidConnect: &pb.OpenIDConnect{
				Enabled:    true,
				LoginUrl:   "/auth/oidc/login",
				LoginLabel: "login label",
				LogoutUrl:  "http://logout.com",
			},
		}, resp)
	})

	ts.T().Run("OpenIDConnectLogin", func(t *testing.T) {
		validator := &TestValidator{returnSubject: "user"}
		api := NewInternalAPI(validator)

		t.Run("Token error", func(t *testing.T) {
			assert := require.New(t)

			oidc.MockGetUserUser = &oidc.User{}
			oidc.MockGetUserError = errors.New("token error")

			_, err := api.OpenIDConnectLogin(context.Background(), &pb.OpenIDConnectLoginRequest{
				Code:  "A",
				State: "B",
			})
			assert.Equal("rpc error: code = Unknown desc = token error", err.Error())
		})

		t.Run("Email not verified", func(t *testing.T) {
			assert := require.New(t)

			oidc.MockGetUserUser = &oidc.User{
				ExternalID:    "ext-test-id",
				Email:         "foo@bar.com",
				EmailVerified: false,
			}
			oidc.MockGetUserError = nil

			_, err := api.OpenIDConnectLogin(context.Background(), &pb.OpenIDConnectLoginRequest{
				Code:  "A",
				State: "B",
			})
			assert.Equal("rpc error: code = FailedPrecondition desc = email address must be verified before you can login", err.Error())
		})

		t.Run("User does not exist, registration disabled", func(t *testing.T) {
			assert := require.New(t)
			registrationEnabled = false

			oidc.MockGetUserUser = &oidc.User{
				ExternalID:    "ext-test-id",
				Email:         "foo@bar.com",
				EmailVerified: true,
			}
			oidc.MockGetUserError = nil

			_, err := api.OpenIDConnectLogin(context.Background(), &pb.OpenIDConnectLoginRequest{
				Code:  "A",
				State: "B",
			})
			assert.Equal("rpc error: code = NotFound desc = object does not exist", err.Error())
		})

		t.Run("User does not exist, registration enabled", func(t *testing.T) {
			registrationEnabled = true

			t.Run("With callback", func(t *testing.T) {
				oidc.MockGetUserUser = &oidc.User{
					ExternalID:    "ext-test-id-cb",
					Email:         "cb@bar.com",
					EmailVerified: true,
				}
				oidc.MockGetUserError = nil

				t.Run("With error", func(t *testing.T) {
					assert := require.New(t)
					reqChan := make(chan *http.Request, 1)

					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						reqChan <- r
						w.WriteHeader(http.StatusTeapot)
					}))
					defer server.Close()
					registrationCallbackURL = server.URL

					_, err := api.OpenIDConnectLogin(context.Background(), &pb.OpenIDConnectLoginRequest{
						Code:  "A",
						State: "B",
					})
					assert.NotNil(err)

					req := <-reqChan
					assert.NotEqual("0", req.URL.Query().Get("user_id"))
				})

				t.Run("Without error", func(t *testing.T) {
					assert := require.New(t)
					reqChan := make(chan *http.Request, 1)

					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						reqChan <- r
						w.WriteHeader(http.StatusOK)
					}))
					defer server.Close()
					registrationCallbackURL = server.URL

					resp, err := api.OpenIDConnectLogin(context.Background(), &pb.OpenIDConnectLoginRequest{
						Code:  "A",
						State: "B",
					})
					assert.Nil(err)
					assert.NotEqual("", resp.JwtToken)

					user, err := storage.GetUserByEmail(context.Background(), storage.DB(), "cb@bar.com")
					assert.NoError(err)

					req := <-reqChan
					assert.Equal(fmt.Sprintf("%d", user.ID), req.URL.Query().Get("user_id"))
				})
			})

			t.Run("Without callback", func(t *testing.T) {
				assert := require.New(t)
				registrationCallbackURL = ""

				oidc.MockGetUserUser = &oidc.User{
					ExternalID:    "ext-test-id-no-cb",
					Email:         "no-cb@bar.com",
					EmailVerified: true,
				}
				oidc.MockGetUserError = nil

				resp, err := api.OpenIDConnectLogin(context.Background(), &pb.OpenIDConnectLoginRequest{
					Code:  "A",
					State: "B",
				})
				assert.NoError(err)
				assert.NotEqual("", resp.JwtToken)

				user, err := storage.GetUserByEmail(context.Background(), storage.DB(), "no-cb@bar.com")
				assert.NoError(err)
				assert.Equal("ext-test-id-no-cb", *user.ExternalID)
			})
		})

		t.Run("Valid login by email", func(t *testing.T) {
			assert := require.New(t)

			oidc.MockGetUserUser = &oidc.User{
				ExternalID:    "ext-test-id",
				Email:         "foo@bar.com",
				EmailVerified: true,
			}
			oidc.MockGetUserError = nil

			user := storage.User{
				Email: "foo@bar.com",
			}
			assert.NoError(storage.CreateUser(context.Background(), storage.DB(), &user))

			resp, err := api.OpenIDConnectLogin(context.Background(), &pb.OpenIDConnectLoginRequest{
				Code:  "A",
				State: "B",
			})
			assert.NoError(err)
			assert.NotEqual("", resp.JwtToken)

			// validate ext. id has been updated
			user, err = storage.GetUserByEmail(context.Background(), storage.DB(), "foo@bar.com")
			assert.NoError(err)
			assert.Equal("ext-test-id", *user.ExternalID)
			assert.True(user.EmailVerified)
		})

		t.Run("Validate login by external id", func(t *testing.T) {
			assert := require.New(t)

			oidc.MockGetUserUser = &oidc.User{
				ExternalID:    "ext-test-id-2",
				Email:         "foo@bar.com",
				EmailVerified: true,
			}
			oidc.MockGetUserError = nil

			user := storage.User{
				Email: "bar@foo.com",
			}
			assert.NoError(storage.CreateUser(context.Background(), storage.DB(), &user))

			resp, err := api.OpenIDConnectLogin(context.Background(), &pb.OpenIDConnectLoginRequest{
				Code:  "A",
				State: "B",
			})
			assert.NoError(err)
			assert.NotEqual("", resp.JwtToken)

			// validate email has been updated
			user, err = storage.GetUserByExternalID(context.Background(), storage.DB(), "ext-test-id-2")
			assert.NoError(err)
			assert.Equal("foo@bar.com", user.Email)
		})
	})
}
