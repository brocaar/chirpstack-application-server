package api

import (
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/loraserver/api/ns"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/gusseleet/lora-app-server/api"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/gusseleet/lora-app-server/internal/test"
)

func TestDeviceProfileServiceAPI(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	config.C.PostgreSQL.DB = db

	Convey("Given a clean database and api instance", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		nsClient := test.NewNetworkServerClient()
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewDeviceProfileServiceAPI(validator)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(config.C.PostgreSQL.DB, &org), ShouldBeNil)

		Convey("Then Create creates a device-profile", func() {
			createReq := pb.CreateDeviceProfileRequest{
				Name:            "test-dp",
				OrganizationID:  org.ID,
				NetworkServerID: n.ID,
				DeviceProfile: &pb.DeviceProfile{
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

			createResp, err := api.Create(ctx, &createReq)
			So(err, ShouldBeNil)
			So(createResp.DeviceProfileID, ShouldNotEqual, "")
			So(nsClient.CreateDeviceProfileChan, ShouldHaveLength, 1)

			// set network-server mock
			nsCreate := <-nsClient.CreateDeviceProfileChan
			nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
				DeviceProfile: nsCreate.DeviceProfile,
			}

			Convey("Then Get returns the device-profile", func() {
				getResp, err := api.Get(ctx, &pb.GetDeviceProfileRequest{
					DeviceProfileID: createResp.DeviceProfileID,
				})
				So(err, ShouldBeNil)
				So(getResp.Name, ShouldEqual, createReq.Name)
				So(getResp.OrganizationID, ShouldEqual, createReq.OrganizationID)
				So(getResp.NetworkServerID, ShouldEqual, createReq.NetworkServerID)
				So(getResp.DeviceProfile, ShouldResemble, &pb.DeviceProfile{
					DeviceProfileID:    createResp.DeviceProfileID,
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
				})
			})

			Convey("Then Update updates the device-profile", func() {
				_, err := api.Update(ctx, &pb.UpdateDeviceProfileRequest{
					Name: "updated-dp",
					DeviceProfile: &pb.DeviceProfile{
						DeviceProfileID:    createResp.DeviceProfileID,
						SupportsClassB:     true,
						ClassBTimeout:      20,
						PingSlotPeriod:     30,
						PingSlotDR:         4,
						PingSlotFreq:       868300000,
						SupportsClassC:     true,
						ClassCTimeout:      20,
						MacVersion:         "1.1.0",
						RegParamsRevision:  "C",
						RxDelay1:           2,
						RxDROffset1:        3,
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
				So(err, ShouldBeNil)
				So(nsClient.UpdateDeviceProfileChan, ShouldHaveLength, 1)

				nsUpdate := <-nsClient.UpdateDeviceProfileChan
				nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
					DeviceProfile: nsUpdate.DeviceProfile,
				}

				getResp, err := api.Get(ctx, &pb.GetDeviceProfileRequest{
					DeviceProfileID: createResp.DeviceProfileID,
				})
				So(err, ShouldBeNil)
				So(getResp.Name, ShouldEqual, "updated-dp")
				So(getResp.OrganizationID, ShouldEqual, createReq.OrganizationID)
				So(getResp.NetworkServerID, ShouldEqual, createReq.NetworkServerID)
				So(getResp.DeviceProfile, ShouldResemble, &pb.DeviceProfile{
					DeviceProfileID:    createResp.DeviceProfileID,
					SupportsClassB:     true,
					ClassBTimeout:      20,
					PingSlotPeriod:     30,
					PingSlotDR:         4,
					PingSlotFreq:       868300000,
					SupportsClassC:     true,
					ClassCTimeout:      20,
					MacVersion:         "1.1.0",
					RegParamsRevision:  "C",
					RxDelay1:           2,
					RxDROffset1:        3,
					RxDataRate2:        5,
					RxFreq2:            868500000,
					FactoryPresetFreqs: []uint32{868100000, 868300000, 868500000, 868700000},
					MaxEIRP:            17,
					MaxDutyCycle:       1,
					SupportsJoin:       true,
					RfRegion:           "EU868",
					Supports32BitFCnt:  true,
				})
			})

			Convey("Then Delete deletes the device-profile", func() {
				_, err := api.Delete(ctx, &pb.DeleteDeviceProfileRequest{
					DeviceProfileID: createResp.DeviceProfileID,
				})
				So(err, ShouldBeNil)

				_, err = api.Get(ctx, &pb.GetDeviceProfileRequest{
					DeviceProfileID: createResp.DeviceProfileID,
				})
				So(err, ShouldNotBeNil)
				So(grpc.Code(err), ShouldEqual, codes.NotFound)
			})

			Convey("Given a global admin user", func() {
				validator.returnIsAdmin = true

				Convey("Then List without organization id returns all device-profiles", func() {
					listResp, err := api.List(ctx, &pb.ListDeviceProfileRequest{
						Limit: 10,
					})
					So(err, ShouldBeNil)
					So(listResp.TotalCount, ShouldEqual, 1)
					So(listResp.Result, ShouldHaveLength, 1)
				})
			})

			Convey("Given an organization user", func() {
				userID, err := storage.CreateUser(config.C.PostgreSQL.DB, &storage.User{
					Username: "testuser",
					IsActive: true,
					Email:    "foo@bar.com",
				}, "testpassword")
				So(err, ShouldBeNil)
				So(storage.CreateOrganizationUser(config.C.PostgreSQL.DB, org.ID, userID, false), ShouldBeNil)

				Convey("Then List without organization id returns all device-profiles related to the user", func() {
					validator.returnUsername = "testuser"
					listResp, err := api.List(ctx, &pb.ListDeviceProfileRequest{
						Limit: 10,
					})
					So(err, ShouldBeNil)
					So(listResp.TotalCount, ShouldEqual, 1)
					So(listResp.Result, ShouldHaveLength, 1)
				})

				Convey("Then calling List using a different username returns no items", func() {
					validator.returnUsername = "differentuser"
					listResp, err := api.List(ctx, &pb.ListDeviceProfileRequest{
						Limit: 10,
					})
					So(err, ShouldBeNil)
					So(listResp.TotalCount, ShouldEqual, 0)
					So(listResp.Result, ShouldHaveLength, 0)
				})
			})

			Convey("Then List returns the device-profiles for the given organization id", func() {
				listResp, err := api.List(ctx, &pb.ListDeviceProfileRequest{
					Limit:          10,
					OrganizationID: org.ID,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 1)
				So(listResp.Result, ShouldHaveLength, 1)

				listResp, err = api.List(ctx, &pb.ListDeviceProfileRequest{
					Limit:          10,
					OrganizationID: org.ID + 1,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 0)
				So(listResp.Result, ShouldHaveLength, 0)
			})

			Convey("Given two service-profiles and applications", func() {
				n2 := storage.NetworkServer{
					Name:   "ns-server-2",
					Server: "ns-server-2:1234",
				}
				So(storage.CreateNetworkServer(config.C.PostgreSQL.DB, &n2), ShouldBeNil)

				sp1 := storage.ServiceProfile{
					Name:            "test-sp",
					NetworkServerID: n.ID,
					OrganizationID:  org.ID,
				}
				So(storage.CreateServiceProfile(config.C.PostgreSQL.DB, &sp1), ShouldBeNil)

				sp2 := storage.ServiceProfile{
					Name:            "test-sp-2",
					NetworkServerID: n2.ID,
					OrganizationID:  org.ID,
				}
				So(storage.CreateServiceProfile(config.C.PostgreSQL.DB, &sp2), ShouldBeNil)

				app1 := storage.Application{
					Name:             "test-app",
					Description:      "test app",
					OrganizationID:   org.ID,
					ServiceProfileID: sp1.ServiceProfile.ServiceProfileID,
				}
				So(storage.CreateApplication(config.C.PostgreSQL.DB, &app1), ShouldBeNil)

				app2 := storage.Application{
					Name:             "test-app-2",
					Description:      "test app 2",
					OrganizationID:   org.ID,
					ServiceProfileID: sp2.ServiceProfile.ServiceProfileID,
				}
				So(storage.CreateApplication(config.C.PostgreSQL.DB, &app2), ShouldBeNil)

				Convey("Then List filtered on application id returns the device-profiles accessible by the given application id", func() {
					listResp, err := api.List(ctx, &pb.ListDeviceProfileRequest{
						Limit:         10,
						ApplicationID: app1.ID,
					})
					So(err, ShouldBeNil)
					So(listResp.TotalCount, ShouldEqual, 1)
					So(listResp.Result, ShouldHaveLength, 1)
					So(listResp.Result[0].DeviceProfileID, ShouldEqual, createResp.DeviceProfileID)

					listResp, err = api.List(ctx, &pb.ListDeviceProfileRequest{
						Limit:         10,
						ApplicationID: app2.ID,
					})
					So(err, ShouldBeNil)
					So(listResp.TotalCount, ShouldEqual, 0)
					So(listResp.Result, ShouldHaveLength, 0)
				})
			})
		})
	})
}
