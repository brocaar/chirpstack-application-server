package api

import (
	"net"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/eventlog"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

func TestNodeAPI(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}

	p := storage.NewRedisPool(conf.RedisURL)

	config.C.PostgreSQL.DB = db
	config.C.Redis.Pool = p

	Convey("Given a clean database with an organization, application and api instance", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)
		test.MustFlushRedis(p)

		nsClient := test.NewNetworkServerClient()
		nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
			DeviceProfile: &ns.DeviceProfile{},
		}
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		ctx := context.Background()
		validator := &TestValidator{}

		grpcServer := grpc.NewServer()
		apiServer := NewDeviceAPI(validator)
		pb.RegisterDeviceServiceServer(grpcServer, apiServer)

		ln, err := net.Listen("tcp", "localhost:0")
		So(err, ShouldBeNil)
		go grpcServer.Serve(ln)
		defer func() {
			grpcServer.Stop()
			ln.Close()
		}()

		apiClient, err := grpc.Dial(ln.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
		So(err, ShouldBeNil)
		defer apiClient.Close()

		api := pb.NewDeviceServiceClient(apiClient)

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
		}
		So(storage.CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)
		spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
		So(err, ShouldBeNil)

		app := storage.Application{
			OrganizationID:   org.ID,
			Name:             "test-app",
			ServiceProfileID: spID,
		}
		So(storage.CreateApplication(config.C.PostgreSQL.DB, &app), ShouldBeNil)

		dp := storage.DeviceProfile{
			Name:            "test-dp",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
		}
		So(storage.CreateDeviceProfile(config.C.PostgreSQL.DB, &dp), ShouldBeNil)
		dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
		So(err, ShouldBeNil)

		Convey("When creating a device without a name set", func() {
			_, err := api.Create(ctx, &pb.CreateDeviceRequest{
				Device: &pb.Device{
					DevEui:          "0807060504030201",
					ApplicationId:   app.ID,
					Description:     "test device description",
					DeviceProfileId: dpID.String(),
				},
			})
			So(err, ShouldBeNil)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the DevEUI is used as name", func() {
				d, err := api.Get(ctx, &pb.GetDeviceRequest{
					DevEui: "0807060504030201",
				})
				So(err, ShouldBeNil)
				So(d.Device.Name, ShouldEqual, "0807060504030201")
			})
		})

		Convey("When creating a device", func() {
			createReq := pb.CreateDeviceRequest{
				Device: &pb.Device{
					ApplicationId:   app.ID,
					Name:            "test-device",
					Description:     "test device description",
					DevEui:          "0807060504030201",
					DeviceProfileId: dpID.String(),
					SkipFCntCheck:   true,
				},
			}

			_, err := api.Create(ctx, &createReq)
			So(err, ShouldBeNil)
			So(validator.validatorFuncs, ShouldHaveLength, 1)
			createReq.Device.XXX_sizecache = 0

			nsReq := <-nsClient.CreateDeviceChan
			nsClient.GetDeviceResponse = ns.GetDeviceResponse{
				Device: nsReq.Device,
			}

			Convey("The device has been created", func() {
				d, err := api.Get(ctx, &pb.GetDeviceRequest{
					DevEui: "0807060504030201",
				})
				So(err, ShouldBeNil)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				So(d.Device, ShouldResemble, createReq.Device)
				So(d.LastSeenAt, ShouldBeNil)
				So(d.DeviceStatusBattery, ShouldEqual, 256)
				So(d.DeviceStatusMargin, ShouldEqual, 256)

				Convey("When setting the device-status battery and margin", func() {
					ten := 10
					eleven := 11

					d, err := storage.GetDevice(config.C.PostgreSQL.DB, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, false, true)
					So(err, ShouldBeNil)
					d.DeviceStatusBattery = &ten
					d.DeviceStatusMargin = &eleven
					So(storage.UpdateDevice(config.C.PostgreSQL.DB, &d, true), ShouldBeNil)

					Convey("Then Get returns the battery and margin status", func() {
						d, err := api.Get(ctx, &pb.GetDeviceRequest{
							DevEui: "0807060504030201",
						})
						So(err, ShouldBeNil)
						So(d.DeviceStatusBattery, ShouldEqual, 10)
						So(d.DeviceStatusMargin, ShouldEqual, 11)
					})
				})

				Convey("When setting the LastSeenAt timestamp", func() {
					now := time.Now().Truncate(time.Millisecond)

					d, err := storage.GetDevice(config.C.PostgreSQL.DB, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, false, true)
					So(err, ShouldBeNil)
					d.LastSeenAt = &now
					So(storage.UpdateDevice(config.C.PostgreSQL.DB, &d, true), ShouldBeNil)

					Convey("Then Get returns the last-seen timestamp", func() {
						d, err := api.Get(ctx, &pb.GetDeviceRequest{
							DevEui: "0807060504030201",
						})
						So(err, ShouldBeNil)
						So(d.LastSeenAt, ShouldNotBeNil)
					})
				})
			})

			Convey("Testing the List method", func() {
				user := storage.User{
					Username: "testuser",
					Email:    "test@test.com",
					IsActive: true,
				}
				_, err := storage.CreateUser(db, &user, "testpassword")
				So(err, ShouldBeNil)

				Convey("Then a global admin user can list all devices", func() {
					validator.returnIsAdmin = true
					devices, err := api.List(ctx, &pb.ListDeviceRequest{
						Limit:  10,
						Offset: 0,
					})
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)
					So(devices.TotalCount, ShouldEqual, 1)
					So(devices.Result, ShouldHaveLength, 1)

					devices, err = api.List(ctx, &pb.ListDeviceRequest{
						Limit:         10,
						Offset:        0,
						ApplicationId: app.ID,
					})
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)
					So(devices.TotalCount, ShouldEqual, 1)
					So(devices.Result, ShouldHaveLength, 1)
				})

				Convey("Then a non-admin can not list the devices", func() {
					validator.returnIsAdmin = false
					validator.returnUsername = user.Username

					devices, err := api.List(ctx, &pb.ListDeviceRequest{
						Limit:  10,
						Offset: 0,
					})
					So(err, ShouldBeNil)
					So(devices.TotalCount, ShouldEqual, 0)
				})

				Convey("When assigning the user to the organization", func() {
					So(storage.CreateOrganizationUser(db, org.ID, user.ID, false), ShouldBeNil)

					Convey("Then it can list the devices", func() {
						validator.returnIsAdmin = false
						validator.returnUsername = user.Username

						devices, err := api.List(ctx, &pb.ListDeviceRequest{
							Limit:  10,
							Offset: 0,
						})
						So(err, ShouldBeNil)
						So(devices.TotalCount, ShouldEqual, 1)
						So(devices.Result, ShouldHaveLength, 1)
					})
				})
			})

			Convey("When updating the device", func() {
				updateReq := pb.UpdateDeviceRequest{
					Device: &pb.Device{
						ApplicationId:   app.ID,
						DevEui:          "0807060504030201",
						Name:            "test-device-updated",
						Description:     "test device description updated",
						DeviceProfileId: dpID.String(),
						SkipFCntCheck:   true,
					},
				}

				_, err := api.Update(ctx, &updateReq)
				So(err, ShouldBeNil)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				updateReq.Device.XXX_sizecache = 0

				Convey("Then the device has been updated", func() {
					d, err := api.Get(ctx, &pb.GetDeviceRequest{
						DevEui: "0807060504030201",
					})
					So(err, ShouldBeNil)
					So(d.Device, ShouldResemble, updateReq.Device)
				})
			})

			Convey("After deleting the device", func() {
				_, err := api.Delete(ctx, &pb.DeleteDeviceRequest{
					DevEui: "0807060504030201",
				})
				So(err, ShouldBeNil)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then listing the devices returns zero devices", func() {
					devices, err := api.List(ctx, &pb.ListDeviceRequest{
						ApplicationId: app.ID,
						Limit:         10,
					})
					So(err, ShouldBeNil)
					So(devices.TotalCount, ShouldEqual, 0)
					So(devices.Result, ShouldHaveLength, 0)
				})
			})

			Convey("Then CreateKeys creates device-keys", func() {
				createReq := pb.CreateDeviceKeysRequest{
					DeviceKeys: &pb.DeviceKeys{
						DevEui: "0807060504030201",
						NwkKey: "01020304050607080807060504030201",
					},
				}

				_, err := api.CreateKeys(ctx, &createReq)
				So(err, ShouldBeNil)

				Convey("Then GetKeys returns the device-keys", func() {
					dk, err := api.GetKeys(ctx, &pb.GetDeviceKeysRequest{
						DevEui: "0807060504030201",
					})
					So(err, ShouldBeNil)
					So(dk.DeviceKeys, ShouldResemble, &pb.DeviceKeys{
						DevEui: "0807060504030201",
						NwkKey: "01020304050607080807060504030201",
						AppKey: "00000000000000000000000000000000",
					})
				})

				Convey("Then UpdateKeys updates the device-keys", func() {
					updateReq := pb.UpdateDeviceKeysRequest{
						DeviceKeys: &pb.DeviceKeys{
							DevEui: "0807060504030201",
							NwkKey: "08070605040302010102030405060708",
						},
					}

					_, err := api.UpdateKeys(ctx, &updateReq)
					So(err, ShouldBeNil)

					dk, err := api.GetKeys(ctx, &pb.GetDeviceKeysRequest{
						DevEui: "0807060504030201",
					})
					So(err, ShouldBeNil)
					So(dk.DeviceKeys, ShouldResemble, &pb.DeviceKeys{
						DevEui: "0807060504030201",
						NwkKey: "08070605040302010102030405060708",
						AppKey: "00000000000000000000000000000000",
					})
				})

				Convey("Then DeleteKeys deletes the device-keys", func() {
					_, err := api.DeleteKeys(ctx, &pb.DeleteDeviceKeysRequest{
						DevEui: "0807060504030201",
					})
					So(err, ShouldBeNil)

					_, err = api.DeleteKeys(ctx, &pb.DeleteDeviceKeysRequest{
						DevEui: "0807060504030201",
					})
					So(err, ShouldNotBeNil)
					So(grpc.Code(err), ShouldEqual, codes.NotFound)
				})
			})

			Convey("When activating the device (ABP)", func() {
				activateReq := pb.ActivateDeviceRequest{
					DeviceActivation: &pb.DeviceActivation{
						DevEui:      "0807060504030201",
						DevAddr:     "01020304",
						AppSKey:     "01020304050607080102030405060708",
						NwkSEncKey:  "08070605040302010807060504030201",
						SNwkSIntKey: "08070605040302010807060504030202",
						FNwkSIntKey: "08070605040302010807060504030203",
						FCntUp:      10,
						NFCntDown:   11,
						AFCntDown:   12,
					},
				}

				_, err := api.Activate(ctx, &activateReq)
				So(err, ShouldBeNil)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then an attempt was made to deactivate the device-session", func() {
					So(nsClient.DeactivateDeviceChan, ShouldHaveLength, 1)
					So(<-nsClient.DeactivateDeviceChan, ShouldResemble, ns.DeactivateDeviceRequest{
						DevEui: []byte{8, 7, 6, 5, 4, 3, 2, 1},
					})
				})

				Convey("Then a device-session was created", func() {
					So(nsClient.ActivateDeviceChan, ShouldHaveLength, 1)
					So(<-nsClient.ActivateDeviceChan, ShouldResemble, ns.ActivateDeviceRequest{
						DeviceActivation: &ns.DeviceActivation{
							DevEui:      []uint8{8, 7, 6, 5, 4, 3, 2, 1},
							DevAddr:     []uint8{1, 2, 3, 4},
							NwkSEncKey:  []uint8{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1},
							SNwkSIntKey: []uint8{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 2},
							FNwkSIntKey: []uint8{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 3},
							FCntUp:      10,
							NFCntDown:   11,
							AFCntDown:   12,
						},
					})
				})

				Convey("Then the activation was stored", func() {
					da, err := storage.GetLastDeviceActivationForDevEUI(config.C.PostgreSQL.DB, [8]byte{8, 7, 6, 5, 4, 3, 2, 1})
					So(err, ShouldBeNil)
					So(da.AppSKey, ShouldEqual, lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8})
					So(da.DevAddr, ShouldEqual, lorawan.DevAddr{1, 2, 3, 4})
				})
			})

			Convey("When calling StreamEventLogs", func() {
				respChan := make(chan *pb.StreamDeviceEventLogsResponse)

				client, err := api.StreamEventLogs(ctx, &pb.StreamDeviceEventLogsRequest{
					DevEui: "0807060504030201",
				})
				So(err, ShouldBeNil)

				// some time for subscribing
				time.Sleep(100 * time.Millisecond)

				go func() {
					for {
						resp, err := client.Recv()
						if err != nil {
							break
						}
						respChan <- resp
					}
				}()

				Convey("When logging an event", func() {
					So(eventlog.LogEventForDevice(lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, eventlog.EventLog{
						Type: eventlog.Join,
					}), ShouldBeNil)

					Convey("Then the event was received by the client", func() {
						resp := <-respChan
						So(resp.Type, ShouldEqual, eventlog.Join)
					})
				})
			})
		})
	})
}
