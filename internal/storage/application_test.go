package storage

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/test"
)

func (ts *StorageTestSuite) TestApplication() {
	assert := require.New(ts.T())

	conf := test.GetConfig()
	assert.NoError(Setup(conf))

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	org := Organization{
		Name: "test-org-123",
	}
	assert.NoError(CreateOrganization(context.Background(), ts.Tx(), &org))

	n := NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:1234",
	}
	assert.NoError(CreateNetworkServer(context.Background(), ts.Tx(), &n))

	sp := ServiceProfile{
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
		Name:            "test-service-profile",
		ServiceProfile: ns.ServiceProfile{
			UlRate:                 100,
			UlBucketSize:           10,
			UlRatePolicy:           ns.RatePolicy_MARK,
			DlRate:                 200,
			DlBucketSize:           20,
			DlRatePolicy:           ns.RatePolicy_DROP,
			AddGwMetadata:          true,
			DevStatusReqFreq:       4,
			ReportDevStatusBattery: true,
			ReportDevStatusMargin:  true,
			DrMin:                  3,
			DrMax:                  5,
			PrAllowed:              true,
			HrAllowed:              true,
			RaAllowed:              true,
			NwkGeoLoc:              true,
			TargetPer:              10,
			MinGwDiversity:         3,
		},
	}
	assert.NoError(CreateServiceProfile(context.Background(), ts.Tx(), &sp))
	spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
	assert.NoError(err)

	ts.T().Run("Create with invalid name", func(t *testing.T) {
		assert := require.New(t)

		app := Application{
			OrganizationID:   org.ID,
			ServiceProfileID: spID,
			Name:             "i contain spaces",
		}
		err := CreateApplication(context.Background(), ts.Tx(), &app)
		assert.Equal(ErrApplicationInvalidName, errors.Cause(err))
	})

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		app := Application{
			OrganizationID:       org.ID,
			ServiceProfileID:     spID,
			Name:                 "test-application",
			Description:          "A test application",
			PayloadCodec:         "CUSTOM_JS",
			PayloadEncoderScript: "Encode() {}",
			PayloadDecoderScript: "Decode() {}",
			MQTTTLSCert:          []byte{},
		}
		assert.NoError(CreateApplication(context.Background(), ts.Tx(), &app))

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			app2, err := GetApplication(context.Background(), ts.Tx(), app.ID)
			assert.NoError(err)
			assert.Equal(app, app2)
		})

		t.Run("Get applications", func(t *testing.T) {
			assert := require.New(t)

			apps, err := GetApplications(context.Background(), ts.Tx(), ApplicationFilters{
				Limit: 10,
			})
			assert.NoError(err)
			assert.Len(apps, 1)
			assert.Equal(app.ID, apps[0].ID)
			assert.Equal(sp.Name, apps[0].ServiceProfileName)

			apps, err = GetApplications(context.Background(), ts.Tx(), ApplicationFilters{OrganizationID: org.ID, Limit: 10})
			assert.NoError(err)
			assert.Len(apps, 1)
			assert.Equal(app.ID, apps[0].ID)
			assert.Equal(sp.Name, apps[0].ServiceProfileName)
		})

		t.Run("Get count", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetApplicationCount(context.Background(), ts.Tx(), ApplicationFilters{})
			assert.NoError(err)
			assert.Equal(1, count)

			count, err = GetApplicationCount(context.Background(), ts.Tx(), ApplicationFilters{
				OrganizationID: org.ID,
			})
			assert.NoError(err)
			assert.Equal(1, count)
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			app.Description = "some new description"
			app.MQTTTLSCert = []byte{1, 2, 3}

			assert.NoError(UpdateApplication(context.Background(), ts.Tx(), app))

			app2, err := GetApplication(context.Background(), ts.Tx(), app.ID)
			assert.NoError(err)
			assert.Equal(app, app2)
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteApplication(context.Background(), ts.Tx(), app.ID))

			count, err := GetApplicationCount(context.Background(), ts.Tx(), ApplicationFilters{})
			assert.NoError(err)
			assert.Equal(0, count)
		})
	})
}
