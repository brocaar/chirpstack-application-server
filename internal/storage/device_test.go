package storage

import (
	"testing"
	"time"

	"github.com/brocaar/loraserver/api/ns"

	"github.com/brocaar/lorawan"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/test"
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
			ServiceProfile: backend.ServiceProfile{
				ULRate:                 100,
				ULBucketSize:           10,
				ULRatePolicy:           backend.Mark,
				DLRate:                 200,
				DLBucketSize:           20,
				DLRatePolicy:           backend.Drop,
				AddGWMetadata:          true,
				DevStatusReqFreq:       4,
				ReportDevStatusBattery: true,
				ReportDevStatusMargin:  true,
				DRMin:          3,
				DRMax:          5,
				PRAllowed:      true,
				HRAllowed:      true,
				RAAllowed:      true,
				NwkGeoLoc:      true,
				TargetPER:      10,
				MinGWDiversity: 3,
			},
		}
		So(CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)

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
		So(CreateDeviceProfile(config.C.PostgreSQL.DB, &dp), ShouldBeNil)

		app := Application{
			OrganizationID:   org.ID,
			Name:             "test-app",
			ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
		}
		So(CreateApplication(config.C.PostgreSQL.DB, &app), ShouldBeNil)

		Convey("Then CreateDevice creates the device", func() {
			ten := 10
			eleven := 11

			d := Device{
				DevEUI:              lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				ApplicationID:       app.ID,
				DeviceProfileID:     dp.DeviceProfile.DeviceProfileID,
				Name:                "test-device",
				Description:         "test device",
				DeviceStatusBattery: &ten,
				DeviceStatusMargin:  &eleven,
			}
			So(CreateDevice(config.C.PostgreSQL.DB, &d), ShouldBeNil)
			d.CreatedAt = d.CreatedAt.UTC().Truncate(time.Millisecond)
			d.UpdatedAt = d.UpdatedAt.UTC().Truncate(time.Millisecond)

			So(nsClient.CreateDeviceChan, ShouldHaveLength, 1)
			So(<-nsClient.CreateDeviceChan, ShouldResemble, ns.CreateDeviceRequest{
				Device: &ns.Device{
					DevEUI:           []byte{1, 2, 3, 4, 5, 6, 7, 8},
					DeviceProfileID:  dp.DeviceProfile.DeviceProfileID,
					ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
					RoutingProfileID: config.C.ApplicationServer.ID,
				},
			})

			Convey("Then GetDevice returns the device", func() {
				nsClient.GetDeviceResponse = ns.GetDeviceResponse{
					Device: &ns.Device{
						DevEUI:           []byte{1, 2, 3, 4, 5, 6, 7, 8},
						DeviceProfileID:  dp.DeviceProfile.DeviceProfileID,
						ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
						RoutingProfileID: config.C.ApplicationServer.ID,
					},
				}

				dGet, err := GetDevice(config.C.PostgreSQL.DB, d.DevEUI)
				So(err, ShouldBeNil)
				dGet.CreatedAt = dGet.CreatedAt.UTC().Truncate(time.Millisecond)
				dGet.UpdatedAt = dGet.UpdatedAt.UTC().Truncate(time.Millisecond)

				So(dGet, ShouldResemble, d)

				Convey("Then UpdateDevice updates the device", func() {
					dp2 := DeviceProfile{
						NetworkServerID: n.ID,
						OrganizationID:  org.ID,
						Name:            "device-profile-2",
						DeviceProfile:   backend.DeviceProfile{},
					}
					So(CreateDeviceProfile(config.C.PostgreSQL.DB, &dp2), ShouldBeNil)

					d.Name = "updated-test-device"
					d.DeviceProfileID = dp2.DeviceProfile.DeviceProfileID
					So(UpdateDevice(config.C.PostgreSQL.DB, &d), ShouldBeNil)
					d.UpdatedAt = d.UpdatedAt.UTC().Truncate(time.Millisecond)

					So(nsClient.UpdateDeviceChan, ShouldHaveLength, 1)
					So(<-nsClient.UpdateDeviceChan, ShouldResemble, ns.UpdateDeviceRequest{
						Device: &ns.Device{
							DevEUI:           []byte{1, 2, 3, 4, 5, 6, 7, 8},
							DeviceProfileID:  dp2.DeviceProfile.DeviceProfileID,
							ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
							RoutingProfileID: config.C.ApplicationServer.ID,
						},
					})

					dGet, err := GetDevice(config.C.PostgreSQL.DB, d.DevEUI)
					So(err, ShouldBeNil)
					dGet.CreatedAt = dGet.CreatedAt.UTC().Truncate(time.Millisecond)
					dGet.UpdatedAt = dGet.UpdatedAt.UTC().Truncate(time.Millisecond)
					So(dGet, ShouldResemble, d)
				})

				Convey("Then DeleteDevice deletes the device", func() {
					So(DeleteDevice(config.C.PostgreSQL.DB, d.DevEUI), ShouldBeNil)
					So(nsClient.DeleteDeviceChan, ShouldHaveLength, 1)
					So(<-nsClient.DeleteDeviceChan, ShouldResemble, ns.DeleteDeviceRequest{
						DevEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
					})

					_, err := GetDevice(config.C.PostgreSQL.DB, d.DevEUI)
					So(err, ShouldEqual, ErrDoesNotExist)
				})

				Convey("Then CreateDeviceKeys creates the device-keys", func() {
					dc := DeviceKeys{
						DevEUI:    d.DevEUI,
						AppKey:    lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1},
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
						dc.AppKey = lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}
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
						AppSKey: lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1},
					}
					So(CreateDeviceActivation(config.C.PostgreSQL.DB, &da), ShouldBeNil)
					da.CreatedAt = da.CreatedAt.UTC().Truncate(time.Millisecond)

					daGet, err := GetLastDeviceActivationForDevEUI(config.C.PostgreSQL.DB, d.DevEUI)
					So(err, ShouldBeNil)
					daGet.CreatedAt = daGet.CreatedAt.UTC().Truncate(time.Millisecond)
					So(daGet, ShouldResemble, da)

					Convey("Then GetLastDeviceActivationForDevEUI returns the last actication", func() {
						da2 := DeviceActivation{
							DevEUI:  d.DevEUI,
							DevAddr: lorawan.DevAddr{4, 3, 2, 1},
							NwkSKey: lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1},
							AppSKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
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

			Convey("Then GetDevicesForApplicationID returns the device", func() {
				devices, err := GetDevicesForApplicationID(config.C.PostgreSQL.DB, app.ID, 10, 0, "")
				So(err, ShouldBeNil)
				So(devices, ShouldHaveLength, 1)
				So(devices[0].DevEUI, ShouldEqual, d.DevEUI)
				So(devices[0].DeviceProfileName, ShouldEqual, dp.Name)
			})

			Convey("Then GetDeviceCountForApplicationID returns 1", func() {
				count, err := GetDeviceCountForApplicationID(config.C.PostgreSQL.DB, app.ID, "")
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})
		})
	})
}
