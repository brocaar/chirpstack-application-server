package storage

import (
	"context"
	"testing"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
)

func (ts *StorageTestSuite) TestAPIKey() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	n := NetworkServer{
		Name:   "test",
		Server: "test:1234",
	}
	assert.NoError(CreateNetworkServer(context.Background(), ts.tx, &n))

	org := Organization{
		Name: "test-org",
	}
	assert.NoError(CreateOrganization(context.Background(), ts.tx, &org))

	sp := ServiceProfile{
		Name:            "test-sp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(CreateServiceProfile(context.Background(), ts.tx, &sp))
	var spID uuid.UUID
	copy(spID[:], sp.ServiceProfile.Id)

	app := Application{
		Name:             "test-app",
		OrganizationID:   org.ID,
		ServiceProfileID: spID,
	}
	assert.NoError(CreateApplication(context.Background(), ts.tx, &app))

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		apiKey := APIKey{
			Name:    "admin token",
			IsAdmin: true,
		}

		str, err := CreateAPIKey(context.Background(), ts.tx, &apiKey)
		assert.NoError(err)
		apiKey.CreatedAt = apiKey.CreatedAt.UTC().Truncate(time.Millisecond)

		var claims jwt.MapClaims

		token, err := jwt.ParseWithClaims(str, &claims, func(token *jwt.Token) (interface{}, error) {
			assert.Equal("HS256", token.Header["alg"])
			return jwtsecret, nil
		})

		assert.True(token.Valid)

		assert.Equal("api_key", claims["sub"])
		assert.Equal(apiKey.ID.String(), claims["api_key_id"])

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			res, err := GetAPIKey(context.Background(), ts.tx, apiKey.ID)
			assert.NoError(err)
			res.CreatedAt = res.CreatedAt.UTC().Truncate(time.Millisecond)

			assert.Equal(apiKey, res)
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteAPIKey(context.Background(), ts.tx, apiKey.ID))
			_, err := GetAPIKey(context.Background(), ts.tx, apiKey.ID)
			assert.Equal(ErrDoesNotExist, err)
		})

		t.Run("Listing", func(t *testing.T) {
			assert := require.New(t)

			keyAdmin := APIKey{
				Name:    "admin token",
				IsAdmin: true,
			}
			_, err := CreateAPIKey(context.Background(), ts.tx, &keyAdmin)
			assert.NoError(err)

			keyApp := APIKey{
				Name:          "app key",
				ApplicationID: &app.ID,
			}
			_, err = CreateAPIKey(context.Background(), ts.tx, &keyApp)
			assert.NoError(err)

			keyOrg := APIKey{
				Name:           "org key",
				OrganizationID: &org.ID,
			}
			_, err = CreateAPIKey(context.Background(), ts.tx, &keyOrg)
			assert.NoError(err)

			tests := []struct {
				Name          string
				Filters       APIKeyFilters
				Expected      []APIKey
				ExpectedCount int
			}{
				{
					Name:          "no admin",
					Filters:       APIKeyFilters{Limit: 10},
					ExpectedCount: 2,
					Expected:      []APIKey{keyApp, keyOrg},
				},
				{
					Name:          "admin",
					Filters:       APIKeyFilters{IsAdmin: true, Limit: 10},
					ExpectedCount: 1,
					Expected:      []APIKey{keyAdmin},
				},
				{
					Name:          "org",
					Filters:       APIKeyFilters{OrganizationID: &org.ID, Limit: 10},
					ExpectedCount: 1,
					Expected:      []APIKey{keyOrg},
				},
				{
					Name:          "app",
					Filters:       APIKeyFilters{ApplicationID: &app.ID, Limit: 10},
					ExpectedCount: 1,
					Expected:      []APIKey{keyApp},
				},
				{
					Name:          "test pagination",
					Filters:       APIKeyFilters{Limit: 1, Offset: 1},
					ExpectedCount: 2,
					Expected:      []APIKey{keyOrg},
				},
			}

			for _, tst := range tests {
				t.Run(tst.Name, func(t *testing.T) {
					assert := require.New(t)

					count, err := GetAPIKeyCount(context.Background(), ts.tx, tst.Filters)
					assert.NoError(err)
					assert.Equal(tst.ExpectedCount, count)

					items, err := GetAPIKeys(context.Background(), ts.tx, tst.Filters)
					assert.NoError(err)

					assert.Equal(len(tst.Expected), len(items))

					for i := range items {
						assert.Equal(tst.Expected[i].Name, items[i].Name)
					}
				})
			}
		})
	})
}
