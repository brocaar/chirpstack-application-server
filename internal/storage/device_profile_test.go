package storage

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/brocaar/loraserver/api/ns"

	"github.com/brocaar/lora-app-server/internal/config"
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
	config.C.PostgreSQL.DB = db
	nsClient := test.NewNetworkServerClient()
	config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

	Convey("Given a clean database with organization and network-server", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(config.C.PostgreSQL.DB, &org), ShouldBeNil)

		u := User{
			Username: "testuser",
			IsAdmin:  false,
			IsActive: true,
			Email:    "foo@bar.com",
		}
		uID, err := CreateUser(config.C.PostgreSQL.DB, &u, "testpassword")
		So(err, ShouldBeNil)
		So(CreateOrganizationUser(config.C.PostgreSQL.DB, org.ID, uID, false), ShouldBeNil)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

		Convey("Then CreateDeviceProfile creates the device-profile", func() {
			dp := DeviceProfile{
				NetworkServerID: n.ID,
				OrganizationID:  org.ID,
				Name:            "device-profile",
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
					RfRegion:           string(backend.EU868),
					Supports_32BitFCnt: true,
				},
			}
			So(CreateDeviceProfile(config.C.PostgreSQL.DB, &dp), ShouldBeNil)
			dp.CreatedAt = dp.CreatedAt.UTC().Truncate(time.Millisecond)
			dp.UpdatedAt = dp.UpdatedAt.UTC().Truncate(time.Millisecond)
			dp.DeviceProfile.XXX_sizecache = 0
			dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
			So(err, ShouldBeNil)

			So(nsClient.CreateDeviceProfileChan, ShouldHaveLength, 1)
			So(<-nsClient.CreateDeviceProfileChan, ShouldResemble, ns.CreateDeviceProfileRequest{
				DeviceProfile: &dp.DeviceProfile,
			})

			Convey("Then GetDeviceProfile returns the device-profile", func() {
				nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
					DeviceProfile: &dp.DeviceProfile,
				}
				dpGet, err := GetDeviceProfile(config.C.PostgreSQL.DB, dpID)
				So(err, ShouldBeNil)
				dpGet.CreatedAt = dpGet.CreatedAt.UTC().Truncate(time.Millisecond)
				dpGet.UpdatedAt = dpGet.UpdatedAt.UTC().Truncate(time.Millisecond)

				So(dpGet, ShouldResemble, dp)
			})

			Convey("Then UpdateDeviceProfile updates the device-profile", func() {
				dp.Name = "updated-device-profile"
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
					RfRegion:           string(backend.EU868),
					Supports_32BitFCnt: true,
				}
				So(UpdateDeviceProfile(config.C.PostgreSQL.DB, &dp), ShouldBeNil)
				dp.UpdatedAt = dp.UpdatedAt.UTC().Truncate(time.Millisecond)
				So(nsClient.UpdateDeviceProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.UpdateDeviceProfileChan, ShouldResemble, ns.UpdateDeviceProfileRequest{
					DeviceProfile: &dp.DeviceProfile,
				})

				dpGet, err := GetDeviceProfile(config.C.PostgreSQL.DB, dpID)
				So(err, ShouldBeNil)
				dpGet.UpdatedAt = dpGet.UpdatedAt.UTC().Truncate(time.Millisecond)
				So(dpGet.Name, ShouldEqual, "updated-device-profile")
				So(dpGet.UpdatedAt, ShouldResemble, dp.UpdatedAt)
			})

			Convey("Then DeleteDeviceProfile deletes the device-profile", func() {
				So(DeleteDeviceProfile(config.C.PostgreSQL.DB, dpID), ShouldBeNil)
				So(nsClient.DeleteDeviceProfileChan, ShouldHaveLength, 1)
				So(<-nsClient.DeleteDeviceProfileChan, ShouldResemble, ns.DeleteDeviceProfileRequest{
					Id: dp.DeviceProfile.Id,
				})

				_, err := GetDeviceProfile(config.C.PostgreSQL.DB, dpID)
				So(err, ShouldEqual, ErrDoesNotExist)
			})

			Convey("Then GetDeviceProfileCount returns 1", func() {
				count, err := GetDeviceProfileCount(config.C.PostgreSQL.DB)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then GetDeviceProfileCountForOrganizationID returns the number of device-profiles for the given organization", func() {
				count, err := GetDeviceProfileCountForOrganizationID(config.C.PostgreSQL.DB, org.ID)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)

				count, err = GetDeviceProfileCountForOrganizationID(config.C.PostgreSQL.DB, org.ID+1)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
			})

			Convey("Then GetDeviceProfileCountForUser returns the device-profile accessible by a given user", func() {
				count, err := GetDeviceProfileCountForUser(config.C.PostgreSQL.DB, u.Username)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)

				count, err = GetDeviceProfileCountForUser(config.C.PostgreSQL.DB, "fakeuser")
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
			})

			Convey("Then GetDeviceProfiles includes the device-profile", func() {
				dps, err := GetDeviceProfiles(config.C.PostgreSQL.DB, 10, 0)
				So(err, ShouldBeNil)
				So(dps, ShouldHaveLength, 1)
				So(dps[0].Name, ShouldEqual, dp.Name)
				So(dps[0].OrganizationID, ShouldEqual, dp.OrganizationID)
				So(dps[0].NetworkServerID, ShouldEqual, dp.NetworkServerID)
				So(dps[0].DeviceProfileID, ShouldEqual, dpID)
			})

			Convey("Then GetDeviceProfilesForOrganizationID returns the device-profiles for the given organization", func() {
				dps, err := GetDeviceProfilesForOrganizationID(config.C.PostgreSQL.DB, org.ID, 10, 0)
				So(err, ShouldBeNil)
				So(dps, ShouldHaveLength, 1)

				dps, err = GetDeviceProfilesForOrganizationID(config.C.PostgreSQL.DB, org.ID+1, 10, 0)
				So(err, ShouldBeNil)
				So(dps, ShouldHaveLength, 0)
			})

			Convey("Then GetDeviceProfilesForUser returns the device-profiles accessible by a given user", func() {
				dps, err := GetDeviceProfilesForUser(config.C.PostgreSQL.DB, u.Username, 10, 0)
				So(err, ShouldBeNil)
				So(dps, ShouldHaveLength, 1)

				dps, err = GetDeviceProfilesForUser(config.C.PostgreSQL.DB, "fakeuser", 10, 0)
				So(err, ShouldBeNil)
				So(dps, ShouldHaveLength, 0)
			})

			Convey("Given two service-profiles and applications", func() {
				n2 := NetworkServer{
					Name:   "ns-server-2",
					Server: "ns-server-2:1234",
				}
				So(CreateNetworkServer(config.C.PostgreSQL.DB, &n2), ShouldBeNil)

				sp1 := ServiceProfile{
					Name:            "test-sp",
					NetworkServerID: n.ID,
					OrganizationID:  org.ID,
				}
				So(CreateServiceProfile(config.C.PostgreSQL.DB, &sp1), ShouldBeNil)
				sp1ID, err := uuid.FromBytes(sp1.ServiceProfile.Id)
				So(err, ShouldBeNil)

				sp2 := ServiceProfile{
					Name:            "test-sp-2",
					NetworkServerID: n2.ID,
					OrganizationID:  org.ID,
				}
				So(CreateServiceProfile(config.C.PostgreSQL.DB, &sp2), ShouldBeNil)
				sp2ID, err := uuid.FromBytes(sp2.ServiceProfile.Id)
				So(err, ShouldBeNil)

				app1 := Application{
					Name:             "test-app",
					Description:      "test app",
					OrganizationID:   org.ID,
					ServiceProfileID: sp1ID,
				}
				So(CreateApplication(config.C.PostgreSQL.DB, &app1), ShouldBeNil)

				app2 := Application{
					Name:             "test-app-2",
					Description:      "test app 2",
					OrganizationID:   org.ID,
					ServiceProfileID: sp2ID,
				}
				So(CreateApplication(config.C.PostgreSQL.DB, &app2), ShouldBeNil)

				Convey("Then GetDeviceProfileCountForApplicationID returns the devices-profiles for the given application", func() {
					count, err := GetDeviceProfileCountForApplicationID(config.C.PostgreSQL.DB, app1.ID)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 1)

					count, err = GetDeviceProfileCountForApplicationID(config.C.PostgreSQL.DB, app2.ID)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)
				})

				Convey("Then GetDeviceProfilesForApplicationID returns the device-profile available for the given application id", func() {
					dps, err := GetDeviceProfilesForApplicationID(config.C.PostgreSQL.DB, app1.ID, 10, 0)
					So(err, ShouldBeNil)
					So(dps, ShouldHaveLength, 1)
					So(dps[0].DeviceProfileID, ShouldEqual, dpID)

					dps, err = GetDeviceProfilesForApplicationID(config.C.PostgreSQL.DB, app2.ID, 10, 0)
					So(err, ShouldBeNil)
					So(dps, ShouldHaveLength, 0)
				})
			})
		})
	})
}
