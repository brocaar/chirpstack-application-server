package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/lib/pq/hstore"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/band"
)

func (ts *StorageTestSuite) TestDevice() {
	assert := require.New(ts.T())

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

	dp := DeviceProfile{
		NetworkServerID: n.ID,
		OrganizationID:  org.ID,
		Name:            "device-profile",
		UplinkInterval:  time.Second * 10,
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
	dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
	assert.NoError(err)

	app := Application{
		OrganizationID:   org.ID,
		Name:             "test-app",
		ServiceProfileID: spID,
	}
	assert.NoError(CreateApplication(context.Background(), ts.Tx(), &app))

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		ten := float32(10.5)
		eleven := 11

		d := Device{
			DevEUI:              lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			ApplicationID:       app.ID,
			DeviceProfileID:     dpID,
			Name:                "test-device",
			Description:         "test device",
			DeviceStatusBattery: &ten,
			DeviceStatusMargin:  &eleven,
			SkipFCntCheck:       true,
			ReferenceAltitude:   5.6,
			Variables: hstore.Hstore{
				Map: map[string]sql.NullString{
					"var_1": sql.NullString{String: "test value", Valid: true},
				},
			},
			Tags: hstore.Hstore{
				Map: map[string]sql.NullString{
					"foo": sql.NullString{String: "bar", Valid: true},
				},
			},
			IsDisabled: true,
		}
		assert.NoError(CreateDevice(context.Background(), ts.Tx(), &d))
		d.CreatedAt = d.CreatedAt.UTC().Truncate(time.Millisecond)
		d.UpdatedAt = d.UpdatedAt.UTC().Truncate(time.Millisecond)

		createReq := <-nsClient.CreateDeviceChan
		assert.Equal(ns.CreateDeviceRequest{
			Device: &ns.Device{
				DevEui:            []byte{1, 2, 3, 4, 5, 6, 7, 8},
				DeviceProfileId:   dp.DeviceProfile.Id,
				ServiceProfileId:  sp.ServiceProfile.Id,
				RoutingProfileId:  applicationServerID.Bytes(),
				SkipFCntCheck:     true,
				ReferenceAltitude: 5.6,
				IsDisabled:        true,
			},
		}, createReq)

		t.Run("List", func(t *testing.T) {
			assert := require.New(t)

			devices, err := GetDevices(context.Background(), ts.Tx(), DeviceFilters{Limit: 10})
			assert.NoError(err)
			assert.Len(devices, 1)

			count, err := GetDeviceCount(context.Background(), ts.Tx(), DeviceFilters{})
			assert.NoError(err)
			assert.Equal(1, count)
		})

		t.Run("List by OrganizationID", func(t *testing.T) {
			assert := require.New(t)

			devices, err := GetDevices(context.Background(), ts.Tx(), DeviceFilters{Limit: 10, OrganizationID: org.ID})
			assert.NoError(err)
			assert.Len(devices, 1)

			count, err := GetDeviceCount(context.Background(), ts.Tx(), DeviceFilters{OrganizationID: org.ID})
			assert.NoError(err)
			assert.Equal(1, count)
		})

		t.Run("List by ApplicationID", func(t *testing.T) {
			assert := require.New(t)

			devices, err := GetDevices(context.Background(), ts.Tx(), DeviceFilters{Limit: 10, ApplicationID: app.ID})
			assert.NoError(err)
			assert.Len(devices, 1)

			count, err := GetDeviceCount(context.Background(), ts.Tx(), DeviceFilters{ApplicationID: app.ID})
			assert.NoError(err)
			assert.Equal(1, count)
		})

		t.Run("List by Tags", func(t *testing.T) {
			assert := require.New(t)

			devices, err := GetDevices(context.Background(), ts.Tx(), DeviceFilters{Limit: 10, Tags: hstore.Hstore{
				Map: map[string]sql.NullString{
					"foo": sql.NullString{String: "bar", Valid: true},
				},
			}})
			assert.NoError(err)
			assert.Len(devices, 1)

			devices, err = GetDevices(context.Background(), ts.Tx(), DeviceFilters{Limit: 10, Tags: hstore.Hstore{
				Map: map[string]sql.NullString{
					"foo": sql.NullString{String: "bas", Valid: true},
				},
			}})
			assert.NoError(err)
			assert.Len(devices, 0)
		})

		t.Run("Get", func(t *testing.T) {
			nsClient.GetDeviceResponse = ns.GetDeviceResponse{
				Device: createReq.Device,
			}

			getDevice, err := GetDevice(context.Background(), ts.Tx(), d.DevEUI, false, false)
			assert.NoError(err)

			getDevice.CreatedAt = getDevice.CreatedAt.UTC().Truncate(time.Millisecond)
			getDevice.UpdatedAt = getDevice.UpdatedAt.UTC().Truncate(time.Millisecond)
			assert.Equal(d, getDevice)
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			dp2 := DeviceProfile{
				NetworkServerID: n.ID,
				OrganizationID:  org.ID,
				Name:            "device-profile-2",
			}
			assert.NoError(CreateDeviceProfile(context.Background(), ts.Tx(), &dp2))
			dp2ID, err := uuid.FromBytes(dp2.DeviceProfile.Id)
			assert.NoError(err)

			lat := float64(1.123)
			long := float64(2.123)
			alt := float64(3.123)
			dr := 3

			d.Name = "updated-test-device"
			d.DeviceProfileID = dp2ID
			d.Latitude = &lat
			d.Longitude = &long
			d.Altitude = &alt
			d.DR = &dr
			d.Variables.Map["var_2"] = sql.NullString{String: "test var 2", Valid: true}
			d.Tags.Map["bar"] = sql.NullString{String: "foo", Valid: true}
			d.IsDisabled = true

			assert.NoError(UpdateDevice(context.Background(), ts.Tx(), &d, false))
			d.UpdatedAt = d.UpdatedAt.UTC().Truncate(time.Millisecond)

			updateReq := <-nsClient.UpdateDeviceChan
			assert.Equal(ns.UpdateDeviceRequest{
				Device: &ns.Device{
					DevEui:            []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceProfileId:   dp2.DeviceProfile.Id,
					ServiceProfileId:  sp.ServiceProfile.Id,
					RoutingProfileId:  applicationServerID.Bytes(),
					SkipFCntCheck:     true,
					ReferenceAltitude: 5.6,
					IsDisabled:        true,
				},
			}, updateReq)

			nsClient.GetDeviceResponse = ns.GetDeviceResponse{
				Device: updateReq.Device,
			}

			deviceGet, err := GetDevice(context.Background(), ts.Tx(), d.DevEUI, false, false)
			assert.NoError(err)
			deviceGet.CreatedAt = deviceGet.CreatedAt.UTC().Truncate(time.Millisecond)
			deviceGet.UpdatedAt = deviceGet.UpdatedAt.UTC().Truncate(time.Millisecond)
			assert.Equal(d, deviceGet)
		})

		t.Run("GetDevicesActiveInactive", func(t *testing.T) {
			assert := require.New(t)
			ls := time.Now()

			// device is never seen
			d.LastSeenAt = nil
			assert.NoError(UpdateDevice(context.Background(), ts.Tx(), &d, true))

			ai, err := GetDevicesActiveInactive(context.Background(), ts.Tx(), org.ID)
			assert.NoError(err)
			assert.Equal(DevicesActiveInactive{
				NeverSeenCount: 1,
				ActiveCount:    0,
				InactiveCount:  0,
			}, ai)

			// device is active
			d.LastSeenAt = &ls
			assert.NoError(UpdateDevice(context.Background(), ts.Tx(), &d, true))

			ai, err = GetDevicesActiveInactive(context.Background(), ts.Tx(), org.ID)
			assert.NoError(err)
			assert.Equal(DevicesActiveInactive{
				NeverSeenCount: 0,
				ActiveCount:    1,
				InactiveCount:  0,
			}, ai)

			// device is inactive
			ls = ls.Add(time.Second * -11)
			assert.NoError(UpdateDevice(context.Background(), ts.Tx(), &d, true))

			ai, err = GetDevicesActiveInactive(context.Background(), ts.Tx(), org.ID)
			assert.NoError(err)
			assert.Equal(DevicesActiveInactive{
				NeverSeenCount: 0,
				ActiveCount:    0,
				InactiveCount:  1,
			}, ai)
		})

		t.Run("GetDevicesDataRates", func(t *testing.T) {
			assert := require.New(t)

			// no datarate set
			d.DR = nil
			assert.NoError(UpdateDevice(context.Background(), ts.Tx(), &d, true))

			ddr, err := GetDevicesDataRates(context.Background(), ts.Tx(), org.ID)
			assert.NoError(err)
			assert.Equal(DevicesDataRates{}, ddr)

			// dr 3
			three := 3
			d.DR = &three
			assert.NoError(UpdateDevice(context.Background(), ts.Tx(), &d, true))

			ddr, err = GetDevicesDataRates(context.Background(), ts.Tx(), org.ID)
			assert.NoError(err)
			assert.Equal(DevicesDataRates{
				3: 1,
			}, ddr)
		})

		t.Run("CreateDeviceKeys", func(t *testing.T) {
			assert := require.New(t)

			dk := DeviceKeys{
				DevEUI:    d.DevEUI,
				NwkKey:    lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1},
				AppKey:    lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
				JoinNonce: 1234,
			}
			assert.NoError(CreateDeviceKeys(context.Background(), ts.Tx(), &dk))
			dk.CreatedAt = dk.CreatedAt.UTC().Truncate(time.Millisecond)
			dk.UpdatedAt = dk.UpdatedAt.UTC().Truncate(time.Millisecond)

			t.Run("GetDeviceKeys", func(t *testing.T) {
				assert := require.New(t)

				dkGet, err := GetDeviceKeys(context.Background(), ts.Tx(), d.DevEUI)
				assert.NoError(err)
				dkGet.CreatedAt = dkGet.CreatedAt.UTC().Truncate(time.Millisecond)
				dkGet.UpdatedAt = dkGet.UpdatedAt.UTC().Truncate(time.Millisecond)

				assert.Equal(dk, dkGet)
			})

			t.Run("UpdateDeviceKeys", func(t *testing.T) {
				assert := require.New(t)

				dk.NwkKey = lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}
				dk.AppKey = lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1}
				dk.JoinNonce = 1235
				assert.NoError(UpdateDeviceKeys(context.Background(), ts.Tx(), &dk))
				dk.UpdatedAt = dk.UpdatedAt.UTC().Truncate(time.Millisecond)

				dkGet, err := GetDeviceKeys(context.Background(), ts.Tx(), d.DevEUI)
				assert.NoError(err)
				dkGet.CreatedAt = dkGet.CreatedAt.UTC().Truncate(time.Millisecond)
				dkGet.UpdatedAt = dkGet.UpdatedAt.UTC().Truncate(time.Millisecond)

				assert.Equal(dk, dkGet)
			})

			t.Run("DeleteDeviceKeys", func(t *testing.T) {
				assert := require.New(t)

				assert.NoError(DeleteDeviceKeys(context.Background(), ts.Tx(), d.DevEUI))
				_, err := GetDeviceKeys(context.Background(), ts.Tx(), d.DevEUI)
				assert.Equal(ErrDoesNotExist, err)
			})
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteDevice(context.Background(), ts.Tx(), d.DevEUI))
			assert.Equal(ns.DeleteDeviceRequest{
				DevEui: d.DevEUI[:],
			}, <-nsClient.DeleteDeviceChan)

			_, err := GetDevice(context.Background(), ts.Tx(), d.DevEUI, false, true)
			assert.Equal(ErrDoesNotExist, err)
		})
	})
}
