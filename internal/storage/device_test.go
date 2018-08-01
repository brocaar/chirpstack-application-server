package storage

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

func TestDevice(t *testing.T) {
	conf := test.GetConfig()
	db, err := OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	config.C.PostgreSQL.DB = db
	nsClient := test.NewNetworkServerClient()
	config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

	Convey("Given a clean database with organization, network-server, service-profile, device-profile and application", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(config.C.PostgreSQL.DB, &org), ShouldBeNil)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

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
				DrMin:          3,
				DrMax:          5,
				PrAllowed:      true,
				HrAllowed:      true,
				RaAllowed:      true,
				NwkGeoLoc:      true,
				TargetPer:      10,
				MinGwDiversity: 3,
			},
		}
		So(CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)
		spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
		So(err, ShouldBeNil)

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
		dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
		So(err, ShouldBeNil)

		app := Application{
			OrganizationID:   org.ID,
			Name:             "test-app",
			ServiceProfileID: spID,
		}
		So(CreateApplication(config.C.PostgreSQL.DB, &app), ShouldBeNil)

		Convey("Then CreateDevice creates the device", func() {
			ten := 10
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
			}
			So(CreateDevice(config.C.PostgreSQL.DB, &d), ShouldBeNil)
			d.CreatedAt = d.CreatedAt.UTC().Truncate(time.Millisecond)
			d.UpdatedAt = d.UpdatedAt.UTC().Truncate(time.Millisecond)

			rpID, err := uuid.FromString(config.C.ApplicationServer.ID)
			So(err, ShouldBeNil)

			So(nsClient.CreateDeviceChan, ShouldHaveLength, 1)
			So(<-nsClient.CreateDeviceChan, ShouldResemble, ns.CreateDeviceRequest{
				Device: &ns.Device{
					DevEui:           []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceProfileId:  dp.DeviceProfile.Id,
					ServiceProfileId: sp.ServiceProfile.Id,
					RoutingProfileId: rpID.Bytes(),
					SkipFCntCheck:    true,
				},
			})

			Convey("Then the device can be listed and counted", func() {
				devices, err := GetDevices(db, 10, 0, "")
				So(err, ShouldBeNil)
				So(devices, ShouldHaveLength, 1)

				count, err := GetDeviceCount(db, "")
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then the device can be listed and counted by application id", func() {
				devices, err := GetDevicesForApplicationID(db, app.ID, 10, 0, "")
				So(err, ShouldBeNil)
				So(devices, ShouldHaveLength, 1)

				count, err := GetDeviceCountForApplicationID(db, app.ID, "")
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then GetDevice returns the device", func() {
				nsClient.GetDeviceResponse = ns.GetDeviceResponse{
					Device: &ns.Device{
						DevEui:           []byte{1, 2, 3, 4, 5, 6, 7, 8},
						DeviceProfileId:  dp.DeviceProfile.Id,
						ServiceProfileId: sp.ServiceProfile.Id,
						RoutingProfileId: rpID.Bytes(),
						SkipFCntCheck:    true,
					},
				}

				dGet, err := GetDevice(config.C.PostgreSQL.DB, d.DevEUI, false, false)
				So(err, ShouldBeNil)
				dGet.CreatedAt = dGet.CreatedAt.UTC().Truncate(time.Millisecond)
				dGet.UpdatedAt = dGet.UpdatedAt.UTC().Truncate(time.Millisecond)

				So(dGet, ShouldResemble, d)

				Convey("Then UpdateDevice updates the device", func() {
					dp2 := DeviceProfile{
						NetworkServerID: n.ID,
						OrganizationID:  org.ID,
						Name:            "device-profile-2",
					}
					So(CreateDeviceProfile(config.C.PostgreSQL.DB, &dp2), ShouldBeNil)
					dp2ID, err := uuid.FromBytes(dp2.DeviceProfile.Id)
					So(err, ShouldBeNil)

					d.Name = "updated-test-device"
					d.DeviceProfileID = dp2ID
					So(UpdateDevice(config.C.PostgreSQL.DB, &d, false), ShouldBeNil)
					d.UpdatedAt = d.UpdatedAt.UTC().Truncate(time.Millisecond)

					So(nsClient.UpdateDeviceChan, ShouldHaveLength, 1)
					So(<-nsClient.UpdateDeviceChan, ShouldResemble, ns.UpdateDeviceRequest{
						Device: &ns.Device{
							DevEui:           []byte{1, 2, 3, 4, 5, 6, 7, 8},
							DeviceProfileId:  dp2.DeviceProfile.Id,
							ServiceProfileId: sp.ServiceProfile.Id,
							RoutingProfileId: rpID.Bytes(),
							SkipFCntCheck:    true,
						},
					})

					dGet, err := GetDevice(config.C.PostgreSQL.DB, d.DevEUI, false, false)
					So(err, ShouldBeNil)
					dGet.CreatedAt = dGet.CreatedAt.UTC().Truncate(time.Millisecond)
					dGet.UpdatedAt = dGet.UpdatedAt.UTC().Truncate(time.Millisecond)
					So(dGet, ShouldResemble, d)
				})

				Convey("Then DeleteDevice deletes the device", func() {
					So(DeleteDevice(config.C.PostgreSQL.DB, d.DevEUI), ShouldBeNil)
					So(nsClient.DeleteDeviceChan, ShouldHaveLength, 1)
					So(<-nsClient.DeleteDeviceChan, ShouldResemble, ns.DeleteDeviceRequest{
						DevEui: []byte{1, 2, 3, 4, 5, 6, 7, 8},
					})

					_, err := GetDevice(config.C.PostgreSQL.DB, d.DevEUI, false, true)
					So(err, ShouldEqual, ErrDoesNotExist)
				})

				Convey("Then CreateDeviceKeys creates the device-keys", func() {
					dc := DeviceKeys{
						DevEUI:    d.DevEUI,
						NwkKey:    lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1},
						AppKey:    lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
						JoinNonce: 1234,
					}
					So(CreateDeviceKeys(config.C.PostgreSQL.DB, &dc), ShouldBeNil)
					dc.CreatedAt = dc.CreatedAt.UTC().Truncate(time.Millisecond)
					dc.UpdatedAt = dc.UpdatedAt.UTC().Truncate(time.Millisecond)

					Convey("Then GetDeviceKeys returns the device-keys", func() {
						dcGet, err := GetDeviceKeys(config.C.PostgreSQL.DB, dc.DevEUI)
						So(err, ShouldBeNil)
						dcGet.CreatedAt = dc.CreatedAt.UTC().Truncate(time.Millisecond)
						dcGet.UpdatedAt = dc.UpdatedAt.UTC().Truncate(time.Millisecond)
						So(dcGet, ShouldResemble, dc)
					})

					Convey("Then UpdateDeviceKeys updates the device-keys", func() {
						dc.NwkKey = lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}
						dc.AppKey = lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1}
						dc.JoinNonce = 1235
						So(UpdateDeviceKeys(config.C.PostgreSQL.DB, &dc), ShouldBeNil)
						dc.UpdatedAt = dc.UpdatedAt.UTC().Truncate(time.Millisecond)

						dcGet, err := GetDeviceKeys(config.C.PostgreSQL.DB, dc.DevEUI)
						So(err, ShouldBeNil)
						dcGet.CreatedAt = dc.CreatedAt.UTC().Truncate(time.Millisecond)
						dcGet.UpdatedAt = dc.UpdatedAt.UTC().Truncate(time.Millisecond)
						So(dcGet, ShouldResemble, dc)
					})

					Convey("Then DeleteDeviceKeys deletes the device-keys", func() {
						So(DeleteDeviceKeys(config.C.PostgreSQL.DB, dc.DevEUI), ShouldBeNil)
						_, err := GetDeviceKeys(config.C.PostgreSQL.DB, dc.DevEUI)
						So(err, ShouldEqual, ErrDoesNotExist)
					})
				})

				Convey("Then CreateDeviceActivation creates the device-activation", func() {
					da := DeviceActivation{
						DevEUI:  d.DevEUI,
						DevAddr: lorawan.DevAddr{1, 2, 3, 4},
						AppSKey: lorawan.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2},
					}
					So(CreateDeviceActivation(config.C.PostgreSQL.DB, &da), ShouldBeNil)
					da.CreatedAt = da.CreatedAt.UTC().Truncate(time.Millisecond)

					daGet, err := GetLastDeviceActivationForDevEUI(config.C.PostgreSQL.DB, d.DevEUI)
					So(err, ShouldBeNil)
					daGet.CreatedAt = daGet.CreatedAt.UTC().Truncate(time.Millisecond)
					So(daGet, ShouldResemble, da)

					Convey("Then GetLastDeviceActivationForDevEUI returns the last activation", func() {
						da2 := DeviceActivation{
							DevEUI:  d.DevEUI,
							DevAddr: lorawan.DevAddr{4, 3, 2, 1},
							AppSKey: lorawan.AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2},
						}
						So(CreateDeviceActivation(config.C.PostgreSQL.DB, &da2), ShouldBeNil)
						da2.CreatedAt = da2.CreatedAt.UTC().Truncate(time.Millisecond)

						daGet, err := GetLastDeviceActivationForDevEUI(config.C.PostgreSQL.DB, d.DevEUI)
						So(err, ShouldBeNil)
						daGet.CreatedAt = daGet.CreatedAt.UTC().Truncate(time.Millisecond)
						So(daGet, ShouldResemble, da2)
					})
				})
			})

			Convey("Given an user", func() {
				user := User{
					Username: "testuser",
					IsActive: true,
					Email:    "foo@bar.com",
				}

				_, err := CreateUser(db, &user, "password123")
				So(err, ShouldBeNil)

				Convey("Then no devices can be retrieved for this user", func() {
					// app id given
					devices, err := GetDevicesForUser(db, user.Username, app.ID, 10, 0, "")
					So(err, ShouldBeNil)
					So(devices, ShouldHaveLength, 0)
					count, err := GetDeviceCountForUser(db, user.Username, app.ID, "")
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)

					// no app given
					devices, err = GetDevicesForUser(db, user.Username, 0, 10, 0, "")
					So(err, ShouldBeNil)
					So(devices, ShouldHaveLength, 0)
					count, err = GetDeviceCountForUser(db, user.Username, 0, "")
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)
				})

				Convey("Given an organization user", func() {
					So(CreateOrganizationUser(db, org.ID, user.ID, false), ShouldBeNil)

					Convey("Then devices can be retrieved for this user", func() {
						// app id given
						devices, err := GetDevicesForUser(db, user.Username, app.ID, 10, 0, "")
						So(err, ShouldBeNil)
						So(devices, ShouldHaveLength, 1)
						count, err := GetDeviceCountForUser(db, user.Username, app.ID, "")
						So(err, ShouldBeNil)
						So(count, ShouldEqual, 1)

						// no app given
						devices, err = GetDevicesForUser(db, user.Username, 0, 10, 0, "")
						So(err, ShouldBeNil)
						So(devices, ShouldHaveLength, 1)
						count, err = GetDeviceCountForUser(db, user.Username, 0, "")
						So(err, ShouldBeNil)
						So(count, ShouldEqual, 1)
					})
				})
			})
		})
	})
}
