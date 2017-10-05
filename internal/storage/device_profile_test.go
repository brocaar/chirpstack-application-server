package storage

import (
	"testing"
	"time"

	"github.com/brocaar/loraserver/api/ns"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lorawan/backend"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDeviceProfile(t *testing.T) {
	conf := test.GetConfig()
	db, err := OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	common.DB = db
	nsClient := test.NewNetworkServerClient()
	common.NetworkServer = nsClient

	Convey("Given a clean database with organization and network-server", t, func() {
		test.MustResetDB(common.DB)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(common.DB, &org), ShouldBeNil)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(common.DB, &n), ShouldBeNil)

		Convey("Then CreateDeviceProfile creates the device-profile", func() {
			dp := DeviceProfile{
				NetworkServerID: n.ID,
				OrganizationID:  org.ID,
				Name:            "device-profile",
				DeviceProfile: backend.DeviceProfile{
					SupportsClassB:     true,
					ClassBTimeout:      10,
					PingSlotPeriod:     20,
					PingSlotDR:         5,
					PingSlotFreq:       868100000,
					SupportsClassC:     true,
					ClassCTimeout:      30,
					MACVersion:         "1.0.2",
					RegParamsRevision:  "B",
					RXDelay1:           1,
					RXDROffset1:        1,
					RXDataRate2:        6,
					RXFreq2:            868300000,
					FactoryPresetFreqs: []backend.Frequency{868100000, 868300000, 868500000},
					MaxEIRP:            14,
					MaxDutyCycle:       10,
					SupportsJoin:       true,
					RFRegion:           backend.EU868,
					Supports32bitFCnt:  true,
				},
			}
			So(CreateDeviceProfile(common.DB, &dp), ShouldBeNil)
			dp.CreatedAt = dp.CreatedAt.UTC().Truncate(time.Millisecond)
			dp.UpdatedAt = dp.UpdatedAt.UTC().Truncate(time.Millisecond)

			So(nsClient.CreateDeviceProfileChan, ShouldHaveLength, 1)
			So(<-nsClient.CreateDeviceProfileChan, ShouldResemble, ns.CreateDeviceProfileRequest{
				DeviceProfile: &ns.DeviceProfile{
					DeviceProfileID:    dp.DeviceProfile.DeviceProfileID,
					SupportsClassB:     true,
					ClassBTimeout:      10,
					PingSlotPeriod:     20,
					PingSlotDR:         5,
					PingSlotFreq:       868100000,
					SupportsClassC:     true,
					ClassCTimeout:      30,
					MacVersion:         "1.0.2",
					RegParamsRevision:  "B",
					RxDelay1:           1,
					RxDROffset1:        1,
					RxDataRate2:        6,
					RxFreq2:            868300000,
					FactoryPresetFreqs: []uint32{868100000, 868300000, 868500000},
					MaxEIRP:            14,
					MaxDutyCycle:       10,
					SupportsJoin:       true,
					RfRegion:           "EU868",
					Supports32BitFCnt:  true,
				},
			})

			Convey("Then GetDeviceProfile returns the device-profile", func() {
				nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
					DeviceProfile: &ns.DeviceProfile{
						DeviceProfileID:    dp.DeviceProfile.DeviceProfileID,
						SupportsClassB:     true,
						ClassBTimeout:      10,
						PingSlotPeriod:     20,
						PingSlotDR:         5,
						PingSlotFreq:       868100000,
						SupportsClassC:     true,
						ClassCTimeout:      30,
						MacVersion:         "1.0.2",
						RegParamsRevision:  "B",
						RxDelay1:           1,
						RxDROffset1:        1,
						RxDataRate2:        6,
						RxFreq2:            868300000,
						FactoryPresetFreqs: []uint32{868100000, 868300000, 868500000},
						MaxEIRP:            14,
						MaxDutyCycle:       10,
						SupportsJoin:       true,
						RfRegion:           "EU868",
						Supports32BitFCnt:  true,
					},
				}

				dpGet, err := GetDeviceProfile(common.DB, dp.DeviceProfile.DeviceProfileID)
				So(err, ShouldBeNil)
				dpGet.CreatedAt = dpGet.CreatedAt.UTC().Truncate(time.Millisecond)
				dpGet.UpdatedAt = dpGet.UpdatedAt.UTC().Truncate(time.Millisecond)

				So(dpGet, ShouldResemble, dp)
			})

			Convey("Then UpdateDeviceProfile updates the device-profile", func() {
				dp.Name = "updated-device-profile"
				dp.DeviceProfile = backend.DeviceProfile{
					DeviceProfileID:    dp.DeviceProfile.DeviceProfileID,
					SupportsClassB:     true,
					ClassBTimeout:      11,
					PingSlotPeriod:     21,
					PingSlotDR:         6,
					PingSlotFreq:       868300000,
					SupportsClassC:     true,
					ClassCTimeout:      31,
					MACVersion:         "1.1.0",
					RegParamsRevision:  "B",
					RXDelay1:           2,
					RXDROffset1:        2,
					RXDataRate2:        5,
					RXFreq2:            868500000,
					FactoryPresetFreqs: []backend.Frequency{868100000, 868300000, 868500000, 868700000},
					MaxEIRP:            17,
					MaxDutyCycle:       1,
					SupportsJoin:       true,
					RFRegion:           backend.EU868,
					Supports32bitFCnt:  true,
				}
				So(UpdateDeviceProfile(common.DB, &dp), ShouldBeNil)
				dp.UpdatedAt = dp.UpdatedAt.UTC().Truncate(time.Millisecond)
				So(nsClient.UpdateDeviceProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.UpdateDeviceProfileChan, ShouldResemble, ns.UpdateDeviceProfileRequest{
					DeviceProfile: &ns.DeviceProfile{
						DeviceProfileID:    dp.DeviceProfile.DeviceProfileID,
						SupportsClassB:     true,
						ClassBTimeout:      11,
						PingSlotPeriod:     21,
						PingSlotDR:         6,
						PingSlotFreq:       868300000,
						SupportsClassC:     true,
						ClassCTimeout:      31,
						MacVersion:         "1.1.0",
						RegParamsRevision:  "B",
						RxDelay1:           2,
						RxDROffset1:        2,
						RxDataRate2:        5,
						RxFreq2:            868500000,
						FactoryPresetFreqs: []uint32{868100000, 868300000, 868500000, 868700000},
						MaxEIRP:            17,
						MaxDutyCycle:       1,
						SupportsJoin:       true,
						RfRegion:           "EU868",
						Supports32BitFCnt:  true,
					},
				})

				dpGet, err := GetDeviceProfile(common.DB, dp.DeviceProfile.DeviceProfileID)
				So(err, ShouldBeNil)
				dpGet.UpdatedAt = dpGet.UpdatedAt.UTC().Truncate(time.Millisecond)
				So(dpGet.Name, ShouldEqual, "updated-device-profile")
				So(dpGet.UpdatedAt, ShouldResemble, dp.UpdatedAt)
			})

			Convey("Then DeleteDeviceProfile deletes the device-profile", func() {
				So(DeleteDeviceProfile(common.DB, dp.DeviceProfile.DeviceProfileID), ShouldBeNil)
				So(nsClient.DeleteDeviceProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.DeleteDeviceProfileChan, ShouldResemble, ns.DeleteDeviceProfileRequest{
					DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
				})

				_, err := GetDeviceProfile(common.DB, dp.DeviceProfile.DeviceProfileID)
				So(err, ShouldEqual, ErrDoesNotExist)
			})
		})
	})
}
