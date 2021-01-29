package storage

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
)

func TestServiceProfileValidate(t *testing.T) {
	tests := []struct {
		ServiceProfile ServiceProfile
		Error          error
	}{
		{
			ServiceProfile: ServiceProfile{
				Name: "valid-name",
			},
		},
		{
			ServiceProfile: ServiceProfile{
				Name: "",
			},
			Error: ErrServiceProfileInvalidName,
		},
		{
			ServiceProfile: ServiceProfile{
				Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		},
		{
			ServiceProfile: ServiceProfile{
				Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			Error: ErrServiceProfileInvalidName,
		},
	}

	assert := require.New(t)

	for _, tst := range tests {
		assert.Equal(tst.Error, tst.ServiceProfile.Validate())
	}
}

func (ts *StorageTestSuite) TestServiceProfile() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	org := Organization{
		Name: "test-org",
	}
	assert.NoError(CreateOrganization(context.Background(), ts.Tx(), &org))

	u := User{
		IsAdmin:  false,
		IsActive: true,
		Email:    "foo@bar.com",
	}
	assert.NoError(CreateUser(context.Background(), ts.Tx(), &u))
	assert.NoError(CreateOrganizationUser(context.Background(), ts.Tx(), org.ID, u.ID, false, false, false))

	n := NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:1234",
	}
	assert.NoError(CreateNetworkServer(context.Background(), ts.Tx(), &n))

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

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

		sp.CreatedAt = sp.CreatedAt.UTC().Truncate(time.Millisecond)
		sp.UpdatedAt = sp.UpdatedAt.UTC().Truncate(time.Millisecond)
		spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
		assert.NoError(err)

		createReq := <-nsClient.CreateServiceProfileChan
		if !proto.Equal(createReq.ServiceProfile, &sp.ServiceProfile) {
			assert.Equal(sp.ServiceProfile, createReq.ServiceProfile)
		}
		nsClient.GetServiceProfileResponse.ServiceProfile = createReq.ServiceProfile

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			spGet, err := GetServiceProfile(context.Background(), ts.Tx(), spID, false)
			assert.NoError(err)
			spGet.CreatedAt = spGet.CreatedAt.UTC().Truncate(time.Millisecond)
			spGet.UpdatedAt = spGet.UpdatedAt.UTC().Truncate(time.Millisecond)

			assert.Equal(sp, spGet)
		})

		t.Run("Get without filters", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetServiceProfileCount(context.Background(), ts.Tx(), ServiceProfileFilters{})
			assert.NoError(err)
			assert.Equal(1, count)

			sps, err := GetServiceProfiles(context.Background(), ts.Tx(), ServiceProfileFilters{
				Limit: 10,
			})
			assert.NoError(err)
			assert.Len(sps, 1)
			assert.Equal("test-ns", sps[0].NetworkServerName)
		})

		t.Run("Get for OrganizationID", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetServiceProfileCount(context.Background(), ts.Tx(), ServiceProfileFilters{
				OrganizationID: org.ID,
			})
			assert.NoError(err)
			assert.Equal(1, count)

			count, err = GetServiceProfileCount(context.Background(), ts.Tx(), ServiceProfileFilters{
				OrganizationID: org.ID + 1,
			})
			assert.NoError(err)
			assert.Equal(0, count)

			sps, err := GetServiceProfiles(context.Background(), ts.Tx(), ServiceProfileFilters{
				OrganizationID: org.ID,
				Limit:          10,
			})
			assert.NoError(err)
			assert.Len(sps, 1)

			sps, err = GetServiceProfiles(context.Background(), ts.Tx(), ServiceProfileFilters{
				OrganizationID: org.ID + 1,
				Limit:          10,
			})
			assert.NoError(err)
			assert.Len(sps, 0)
		})

		t.Run("Get for UserID", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetServiceProfileCount(context.Background(), ts.Tx(), ServiceProfileFilters{
				UserID: u.ID,
			})
			assert.NoError(err)
			assert.Equal(1, count)

			count, err = GetServiceProfileCount(context.Background(), ts.Tx(), ServiceProfileFilters{
				UserID: u.ID + 1,
			})
			assert.NoError(err)
			assert.Equal(0, count)

			sps, err := GetServiceProfiles(context.Background(), ts.Tx(), ServiceProfileFilters{
				UserID: u.ID,
				Limit:  10,
			})
			assert.NoError(err)
			assert.Len(sps, 1)

			sps, err = GetServiceProfiles(context.Background(), ts.Tx(), ServiceProfileFilters{
				UserID: u.ID + 1,
				Limit:  10,
			})
			assert.NoError(err)
			assert.Len(sps, 0)
		})

		t.Run("Get for NetworkServerID", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetServiceProfileCount(context.Background(), ts.Tx(), ServiceProfileFilters{
				NetworkServerID: n.ID,
			})
			assert.NoError(err)
			assert.Equal(1, count)

			count, err = GetServiceProfileCount(context.Background(), ts.Tx(), ServiceProfileFilters{
				NetworkServerID: n.ID + 1,
			})
			assert.NoError(err)
			assert.Equal(0, count)

			sps, err := GetServiceProfiles(context.Background(), ts.Tx(), ServiceProfileFilters{
				NetworkServerID: n.ID,
				Limit:           10,
			})
			assert.NoError(err)
			assert.Len(sps, 1)

			sps, err = GetServiceProfiles(context.Background(), ts.Tx(), ServiceProfileFilters{
				NetworkServerID: n.ID + 1,
				Limit:           10,
			})
			assert.NoError(err)
			assert.Len(sps, 0)
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			sp.Name = "updated-service-profile"
			sp.ServiceProfile = ns.ServiceProfile{
				Id:                     sp.ServiceProfile.Id,
				UlRate:                 101,
				UlBucketSize:           11,
				UlRatePolicy:           ns.RatePolicy_DROP,
				DlRate:                 201,
				DlBucketSize:           21,
				DlRatePolicy:           ns.RatePolicy_MARK,
				AddGwMetadata:          true,
				DevStatusReqFreq:       5,
				ReportDevStatusBattery: true,
				ReportDevStatusMargin:  true,
				DrMin:                  4,
				DrMax:                  6,
				PrAllowed:              true,
				HrAllowed:              true,
				RaAllowed:              true,
				NwkGeoLoc:              true,
				TargetPer:              11,
				MinGwDiversity:         4,
			}

			assert.NoError(UpdateServiceProfile(context.Background(), ts.Tx(), &sp))
			sp.UpdatedAt = sp.UpdatedAt.UTC().Truncate(time.Millisecond)

			updateReq := <-nsClient.UpdateServiceProfileChan
			if !proto.Equal(&sp.ServiceProfile, updateReq.ServiceProfile) {
				assert.Equal(sp.ServiceProfile, updateReq.ServiceProfile)
			}

			nsClient.GetServiceProfileResponse.ServiceProfile = updateReq.ServiceProfile
			spGet, err := GetServiceProfile(context.Background(), ts.Tx(), spID, false)
			assert.NoError(err)
			spGet.CreatedAt = spGet.CreatedAt.UTC().Truncate(time.Millisecond)
			spGet.UpdatedAt = spGet.UpdatedAt.UTC().Truncate(time.Millisecond)

			assert.Equal(sp, spGet)
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteServiceProfile(context.Background(), ts.Tx(), spID))
			delReq := <-nsClient.DeleteServiceProfileChan
			assert.Equal(sp.ServiceProfile.Id, delReq.Id)
		})
	})
}
