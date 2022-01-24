package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/proto"
	"github.com/lib/pq/hstore"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lorawan/band"
)

func TestDeviceProfileValidate(t *testing.T) {
	tests := []struct {
		DeviceProfile DeviceProfile
		Error         error
	}{
		{
			DeviceProfile: DeviceProfile{
				Name: "valid-name",
			},
		},
		{
			DeviceProfile: DeviceProfile{
				Name: "",
			},
			Error: ErrDeviceProfileInvalidName,
		},
		{
			DeviceProfile: DeviceProfile{
				Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		},
		{
			DeviceProfile: DeviceProfile{
				Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			Error: ErrDeviceProfileInvalidName,
		},
	}

	assert := require.New(t)

	for _, tst := range tests {
		assert.Equal(tst.Error, tst.DeviceProfile.Validate())
	}
}

func (ts *StorageTestSuite) TestDeviceProfile() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	org := Organization{
		Name: "test-org-123",
	}
	assert.NoError(CreateOrganization(context.Background(), ts.Tx(), &org))

	u := User{
		IsAdmin:  false,
		IsActive: true,
		Email:    "foo@bar.com",
	}
	err := CreateUser(context.Background(), ts.Tx(), &u)
	assert.NoError(err)
	assert.NoError(CreateOrganizationUser(context.Background(), ts.Tx(), org.ID, u.ID, false, false, false))

	n := NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:1234",
	}
	assert.NoError(CreateNetworkServer(context.Background(), DB(), &n))

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		dp := DeviceProfile{
			NetworkServerID:      n.ID,
			OrganizationID:       org.ID,
			Name:                 "device-profile",
			PayloadCodec:         "CUSTOM_JS",
			PayloadEncoderScript: "Encode() {}",
			PayloadDecoderScript: "Decode() {}",
			Tags: hstore.Hstore{
				Map: map[string]sql.NullString{
					"foo": sql.NullString{Valid: true, String: "bar"},
				},
			},
			UplinkInterval: 10 * time.Minute,
			DeviceProfile: ns.DeviceProfile{
				SupportsClassB:     true,
				ClassBTimeout:      10,
				PingSlotPeriod:     20,
				PingSlotDr:         5,
				PingSlotFreq:       868100000,
				SupportsClassC:     true,
				ClassCTimeout:      30,
				MacVersion:         "1.0.2",
				RegParamsRevision:  "B",
				RxDelay_1:          1,
				RxDrOffset_1:       1,
				RxDatarate_2:       6,
				RxFreq_2:           868300000,
				FactoryPresetFreqs: []uint32{868100000, 868300000, 868500000},
				MaxEirp:            14,
				MaxDutyCycle:       10,
				SupportsJoin:       true,
				RfRegion:           string(band.EU868),
				Supports_32BitFCnt: true,
			},
		}
		assert.NoError(CreateDeviceProfile(context.Background(), ts.Tx(), &dp))

		dp.CreatedAt = dp.CreatedAt.UTC().Truncate(time.Millisecond)
		dp.UpdatedAt = dp.UpdatedAt.UTC().Truncate(time.Millisecond)
		dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
		assert.NoError(err)

		createReq := <-nsClient.CreateDeviceProfileChan
		if !proto.Equal(createReq.DeviceProfile, &dp.DeviceProfile) {
			assert.Equal(dp.DeviceProfile, createReq.DeviceProfile)
		}
		nsClient.GetDeviceProfileResponse.DeviceProfile = createReq.DeviceProfile

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			dpGet, err := GetDeviceProfile(context.Background(), ts.Tx(), dpID, true, false)
			assert.NoError(err)
			dpGet.CreatedAt = dpGet.CreatedAt.UTC().Truncate(time.Millisecond)
			dpGet.UpdatedAt = dpGet.UpdatedAt.UTC().Truncate(time.Millisecond)

			assert.Equal(dp, dpGet)
		})

		t.Run("Get all", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetDeviceProfileCount(context.Background(), ts.Tx(), DeviceProfileFilters{})
			assert.NoError(err)
			assert.Equal(1, count)

			dps, err := GetDeviceProfiles(context.Background(), ts.Tx(), DeviceProfileFilters{
				Limit: 10,
			})
			assert.NoError(err)
			assert.Len(dps, 1)
			assert.Equal(dp.Name, dps[0].Name)
			assert.Equal(dp.OrganizationID, dps[0].OrganizationID)
			assert.Equal(dp.NetworkServerID, dps[0].NetworkServerID)
			assert.Equal(dpID, dps[0].DeviceProfileID)
		})

		t.Run("Get for organization ID", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetDeviceProfileCount(context.Background(), ts.Tx(), DeviceProfileFilters{
				OrganizationID: org.ID,
			})
			assert.NoError(err)
			assert.Equal(1, count)

			count, err = GetDeviceProfileCount(context.Background(), ts.Tx(), DeviceProfileFilters{
				OrganizationID: org.ID + 1,
			})
			assert.NoError(err)
			assert.Equal(0, count)

			dps, err := GetDeviceProfiles(context.Background(), ts.Tx(), DeviceProfileFilters{
				OrganizationID: org.ID,
				Limit:          10,
			})
			assert.NoError(err)
			assert.Len(dps, 1)

			dps, err = GetDeviceProfiles(context.Background(), ts.Tx(), DeviceProfileFilters{
				OrganizationID: org.ID + 1,
				Limit:          10,
			})
			assert.NoError(err)
			assert.Len(dps, 0)
		})

		t.Run("Get for username", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetDeviceProfileCount(context.Background(), ts.Tx(), DeviceProfileFilters{
				UserID: u.ID,
			})
			assert.NoError(err)
			assert.Equal(1, count)

			count, err = GetDeviceProfileCount(context.Background(), ts.Tx(), DeviceProfileFilters{
				UserID: u.ID + 999,
			})
			assert.NoError(err)
			assert.Equal(0, count)

			dps, err := GetDeviceProfiles(context.Background(), ts.Tx(), DeviceProfileFilters{
				UserID: u.ID,
				Limit:  10,
			})
			assert.NoError(err)
			assert.Len(dps, 1)

			dps, err = GetDeviceProfiles(context.Background(), ts.Tx(), DeviceProfileFilters{
				UserID: u.ID + 999,
				Limit:  10,
			})
			assert.NoError(err)
			assert.Len(dps, 0)
		})

		t.Run("Two service-profiles and applications", func(t *testing.T) {
			assert := require.New(t)

			n2 := NetworkServer{
				Name:   "ns-server-2",
				Server: "ns-server-2:1234",
			}
			assert.NoError(CreateNetworkServer(context.Background(), ts.Tx(), &n2))

			sp1 := ServiceProfile{
				Name:            "test-sp",
				NetworkServerID: n.ID,
				OrganizationID:  org.ID,
			}
			assert.NoError(CreateServiceProfile(context.Background(), ts.Tx(), &sp1))
			sp1ID, err := uuid.FromBytes(sp1.ServiceProfile.Id)
			assert.NoError(err)

			sp2 := ServiceProfile{
				Name:            "test-sp-2",
				NetworkServerID: n2.ID,
				OrganizationID:  org.ID,
			}
			assert.NoError(CreateServiceProfile(context.Background(), ts.Tx(), &sp2))
			sp2ID, err := uuid.FromBytes(sp2.ServiceProfile.Id)
			assert.NoError(err)

			app1 := Application{
				Name:             "test-app",
				Description:      "test app",
				OrganizationID:   org.ID,
				ServiceProfileID: sp1ID,
			}
			assert.NoError(CreateApplication(context.Background(), ts.Tx(), &app1))

			app2 := Application{
				Name:             "test-app-2",
				Description:      "test app 2",
				OrganizationID:   org.ID,
				ServiceProfileID: sp2ID,
			}
			assert.NoError(CreateApplication(context.Background(), ts.Tx(), &app2))

			t.Run("Get for application ID", func(t *testing.T) {
				assert := require.New(t)

				count, err := GetDeviceProfileCount(context.Background(), ts.Tx(), DeviceProfileFilters{
					ApplicationID: app1.ID,
				})
				assert.NoError(err)
				assert.Equal(1, count)

				count, err = GetDeviceProfileCount(context.Background(), ts.Tx(), DeviceProfileFilters{
					ApplicationID: app2.ID,
				})
				assert.NoError(err)
				assert.Equal(0, count)

				dps, err := GetDeviceProfiles(context.Background(), ts.Tx(), DeviceProfileFilters{
					ApplicationID: app1.ID,
					Limit:         10,
				})
				assert.NoError(err)
				assert.Len(dps, 1)
				assert.Equal(dpID, dps[0].DeviceProfileID)

				dps, err = GetDeviceProfiles(context.Background(), ts.Tx(), DeviceProfileFilters{
					ApplicationID: app2.ID,
					Limit:         10,
				})
				assert.NoError(err)
				assert.Len(dps, 0)
			})
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			dp.Name = "updated-device-profile"
			dp.UplinkInterval = 20 * time.Minute
			dp.DeviceProfile = ns.DeviceProfile{
				Id:                 dp.DeviceProfile.Id,
				SupportsClassB:     true,
				ClassBTimeout:      11,
				PingSlotPeriod:     21,
				PingSlotDr:         6,
				PingSlotFreq:       868300000,
				SupportsClassC:     true,
				ClassCTimeout:      31,
				MacVersion:         "1.1.0",
				RegParamsRevision:  "B",
				RxDelay_1:          2,
				RxDrOffset_1:       2,
				RxDatarate_2:       5,
				RxFreq_2:           868500000,
				FactoryPresetFreqs: []uint32{868100000, 868300000, 868500000, 868700000},
				MaxEirp:            17,
				MaxDutyCycle:       1,
				SupportsJoin:       true,
				RfRegion:           string(band.EU868),
				Supports_32BitFCnt: true,
			}
			assert.NoError(UpdateDeviceProfile(context.Background(), ts.Tx(), &dp))
			dp.UpdatedAt = dp.UpdatedAt.UTC().Truncate(time.Millisecond)

			updateReq := <-nsClient.UpdateDeviceProfileChan
			if !proto.Equal(&dp.DeviceProfile, updateReq.DeviceProfile) {
				assert.Equal(dp.DeviceProfile, updateReq.DeviceProfile)
			}

			nsClient.GetDeviceProfileResponse.DeviceProfile = updateReq.DeviceProfile
			dpGet, err := GetDeviceProfile(context.Background(), ts.Tx(), dpID, true, false)
			assert.NoError(err)
			dpGet.UpdatedAt = dpGet.UpdatedAt.UTC().Truncate(time.Millisecond)
			assert.Equal("updated-device-profile", dpGet.Name)
			assert.Equal(dp.UpdatedAt, dpGet.UpdatedAt)
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteDeviceProfile(context.Background(), ts.Tx(), dpID))
			delReq := <-nsClient.DeleteDeviceProfileChan
			assert.Equal(dp.DeviceProfile.Id, delReq.Id)
		})
	})
}
