package api

import (
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"

	pb "github.com/gusseleet/lora-app-server/api"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/gusseleet/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNodeAPI(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}

	config.C.PostgreSQL.DB = db

	Convey("Given a clean database with an organization, application and api instance", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		nsClient := test.NewNetworkServerClient()
		nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
			DeviceProfile: &ns.DeviceProfile{},
		}
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewDeviceAPI(validator)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(config.C.PostgreSQL.DB, &org), ShouldBeNil)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

		sp := storage.ServiceProfile{
			Name:            "test-sp",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			ServiceProfile:  backend.ServiceProfile{},
		}
		So(storage.CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)

		app := storage.Application{
			OrganizationID:   org.ID,
			Name:             "test-app",
			ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
		}
		So(storage.CreateApplication(config.C.PostgreSQL.DB, &app), ShouldBeNil)

		dp := storage.DeviceProfile{
			Name:            "test-dp",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			DeviceProfile:   backend.DeviceProfile{},
		}
		So(storage.CreateDeviceProfile(config.C.PostgreSQL.DB, &dp), ShouldBeNil)

		Convey("When creating a device without a name set", func() {
			_, err := api.Create(ctx, &pb.CreateDeviceRequest{
				ApplicationID:   app.ID,
				Description:     "test device description",
				DevEUI:          "0807060504030201",
				DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the DevEUI is used as name", func() {
				d, err := api.Get(ctx, &pb.GetDeviceRequest{
					DevEUI: "0807060504030201",
				})
				So(err, ShouldBeNil)
				So(d.Name, ShouldEqual, "0807060504030201")
			})
		})

		Convey("When creating a device", func() {
			_, err := api.Create(ctx, &pb.CreateDeviceRequest{
				ApplicationID:   app.ID,
				Name:            "test-device",
				Description:     "test device description",
				DevEUI:          "0807060504030201",
				DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("The device has been created", func() {
				d, err := api.Get(ctx, &pb.GetDeviceRequest{
					DevEUI: "0807060504030201",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				So(d, ShouldResemble, &pb.GetDeviceResponse{
					Name:                "test-device",
					Description:         "test device description",
					DevEUI:              "0807060504030201",
					ApplicationID:       app.ID,
					DeviceProfileID:     dp.DeviceProfile.DeviceProfileID,
					DeviceStatusMargin:  256,
					DeviceStatusBattery: 256,
				})

				Convey("When setting the device-status battery and margin", func() {
					ten := 10
					eleven := 11

					d, err := storage.GetDevice(config.C.PostgreSQL.DB, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1})
					So(err, ShouldBeNil)
					d.DeviceStatusBattery = &ten
					d.DeviceStatusMargin = &eleven
					So(storage.UpdateDevice(config.C.PostgreSQL.DB, &d), ShouldBeNil)

					Convey("Then Get returns the battery and margin status", func() {
						d, err := api.Get(ctx, &pb.GetDeviceRequest{
							DevEUI: "0807060504030201",
						})
						So(err, ShouldBeNil)
						So(d.DeviceStatusBattery, ShouldEqual, 10)
						So(d.DeviceStatusMargin, ShouldEqual, 11)
					})
				})

				Convey("When setting the LastSeenAt timestamp", func() {
					now := time.Now().Truncate(time.Millisecond)

					d, err := storage.GetDevice(config.C.PostgreSQL.DB, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1})
					So(err, ShouldBeNil)
					d.LastSeenAt = &now
					So(storage.UpdateDevice(config.C.PostgreSQL.DB, &d), ShouldBeNil)

					Convey("Then Get returns the last-seen timestamp", func() {
						d, err := api.Get(ctx, &pb.GetDeviceRequest{
							DevEUI: "0807060504030201",
						})
						So(err, ShouldBeNil)
						So(d.LastSeenAt, ShouldEqual, now.Format(time.RFC3339Nano))
					})
				})
			})

			Convey("Then listing the devices for the application returns a single items", func() {
				devices, err := api.ListByApplicationID(ctx, &pb.ListDeviceByApplicationIDRequest{
					ApplicationID: app.ID,
					Limit:         10,
					Search:        "test",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				So(devices.Result, ShouldHaveLength, 1)
				So(devices.TotalCount, ShouldEqual, 1)
				So(devices.Result[0], ShouldResemble, &pb.DeviceListItem{
					Name:                "test-device",
					Description:         "test device description",
					DevEUI:              "0807060504030201",
					ApplicationID:       app.ID,
					DeviceProfileID:     dp.DeviceProfile.DeviceProfileID,
					DeviceProfileName:   dp.Name,
					DeviceStatusBattery: 256,
					DeviceStatusMargin:  256,
				})
			})

			Convey("When updating the device", func() {
				_, err := api.Update(ctx, &pb.UpdateDeviceRequest{
					ApplicationID:   app.ID,
					DevEUI:          "0807060504030201",
					Name:            "test-device-updated",
					Description:     "test device description updated",
					DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the device has been updated", func() {
					d, err := api.Get(ctx, &pb.GetDeviceRequest{
						DevEUI: "0807060504030201",
					})
					So(err, ShouldBeNil)
					So(d, ShouldResemble, &pb.GetDeviceResponse{
						Name:                "test-device-updated",
						Description:         "test device description updated",
						DevEUI:              "0807060504030201",
						ApplicationID:       app.ID,
						DeviceProfileID:     dp.DeviceProfile.DeviceProfileID,
						DeviceStatusBattery: 256,
						DeviceStatusMargin:  256,
					})
				})
			})

			Convey("After deleting the device", func() {
				_, err := api.Delete(ctx, &pb.DeleteDeviceRequest{
					DevEUI: "0807060504030201",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then listing the devices returns zero devices", func() {
					devices, err := api.ListByApplicationID(ctx, &pb.ListDeviceByApplicationIDRequest{
						ApplicationID: app.ID,
						Limit:         10,
					})
					So(err, ShouldBeNil)
					So(devices.TotalCount, ShouldEqual, 0)
					So(devices.Result, ShouldHaveLength, 0)
				})
			})

			Convey("Then CreateKeys creates device-keys", func() {
				_, err := api.CreateKeys(ctx, &pb.CreateDeviceKeysRequest{
					DevEUI: "0807060504030201",
					DeviceKeys: &pb.DeviceKeys{
						AppKey: "01020304050607080807060504030201",
					},
				})
				So(err, ShouldBeNil)

				Convey("Then GetKeys returns the device-keys", func() {
					dk, err := api.GetKeys(ctx, &pb.GetDeviceKeysRequest{
						DevEUI: "0807060504030201",
					})
					So(err, ShouldBeNil)
					So(dk, ShouldResemble, &pb.GetDeviceKeysResponse{
						DeviceKeys: &pb.DeviceKeys{
							AppKey: "01020304050607080807060504030201",
						},
					})
				})

				Convey("Then UpdateKeys updates the device-keys", func() {
					_, err := api.UpdateKeys(ctx, &pb.UpdateDeviceKeysRequest{
						DevEUI: "0807060504030201",
						DeviceKeys: &pb.DeviceKeys{
							AppKey: "08070605040302010102030405060708",
						},
					})
					So(err, ShouldBeNil)

					dk, err := api.GetKeys(ctx, &pb.GetDeviceKeysRequest{
						DevEUI: "0807060504030201",
					})
					So(err, ShouldBeNil)
					So(dk, ShouldResemble, &pb.GetDeviceKeysResponse{
						DeviceKeys: &pb.DeviceKeys{
							AppKey: "08070605040302010102030405060708",
						},
					})
				})

				Convey("Then DeleteKeys deletes the device-keys", func() {
					_, err := api.DeleteKeys(ctx, &pb.DeleteDeviceKeysRequest{
						DevEUI: "0807060504030201",
					})
					So(err, ShouldBeNil)

					_, err = api.DeleteKeys(ctx, &pb.DeleteDeviceKeysRequest{
						DevEUI: "0807060504030201",
					})
					So(err, ShouldNotBeNil)
					So(grpc.Code(err), ShouldEqual, codes.NotFound)
				})
			})

			Convey("When activating the device (ABP)", func() {
				_, err := api.Activate(ctx, &pb.ActivateDeviceRequest{
					DevEUI:   "0807060504030201",
					DevAddr:  "01020304",
					AppSKey:  "01020304050607080102030405060708",
					NwkSKey:  "08070605040302010807060504030201",
					FCntUp:   10,
					FCntDown: 11,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then an attempt was made to deactivate the device-session", func() {
					So(nsClient.DeactivateDeviceChan, ShouldHaveLength, 1)
					So(<-nsClient.DeactivateDeviceChan, ShouldResemble, ns.DeactivateDeviceRequest{
						DevEUI: []byte{8, 7, 6, 5, 4, 3, 2, 1},
					})
				})

				Convey("Then a device-session was created", func() {
					So(nsClient.ActivateDeviceChan, ShouldHaveLength, 1)
					So(<-nsClient.ActivateDeviceChan, ShouldResemble, ns.ActivateDeviceRequest{
						DevAddr:  []uint8{1, 2, 3, 4},
						DevEUI:   []uint8{8, 7, 6, 5, 4, 3, 2, 1},
						NwkSKey:  []uint8{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1},
						FCntUp:   10,
						FCntDown: 11,
					})
				})

				Convey("Then the activation was stored", func() {
					da, err := storage.GetLastDeviceActivationForDevEUI(config.C.PostgreSQL.DB, [8]byte{8, 7, 6, 5, 4, 3, 2, 1})
					So(err, ShouldBeNil)
					So(da.AppSKey, ShouldEqual, lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8})
					So(da.NwkSKey, ShouldEqual, lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1})
					So(da.DevAddr, ShouldEqual, lorawan.DevAddr{1, 2, 3, 4})
				})
			})
		})
	})
}
