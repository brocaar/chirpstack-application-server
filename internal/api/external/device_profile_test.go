package external

import (
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/loraserver/api/ns"
	uuid "github.com/gofrs/uuid"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
)

func TestDeviceProfileServiceAPI(t *testing.T) {
	conf := test.GetConfig()
	if err := storage.Setup(conf); err != nil {
		t.Fatal(err)
	}

	Convey("Given a clean database and api instance", t, func() {
		test.MustResetDB(storage.DB().DB)

		nsClient := mock.NewClient()
		networkserver.SetPool(mock.NewPool(nsClient))

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewDeviceProfileServiceAPI(validator)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(storage.DB(), &n), ShouldBeNil)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(storage.DB(), &org), ShouldBeNil)

		Convey("Then Create creates a device-profile", func() {
			createReq := pb.CreateDeviceProfileRequest{
				DeviceProfile: &pb.DeviceProfile{
					Name:               "test-dp",
					OrganizationId:     org.ID,
					NetworkServerId:    n.ID,
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
					RfRegion:           "EU868",
					Supports_32BitFCnt: true,
				},
			}

			createResp, err := api.Create(ctx, &createReq)
			So(err, ShouldBeNil)
			So(createResp.Id, ShouldNotEqual, "")
			So(createResp.Id, ShouldNotEqual, uuid.Nil.String())
			So(nsClient.CreateDeviceProfileChan, ShouldHaveLength, 1)

			// set network-server mock
			nsCreate := <-nsClient.CreateDeviceProfileChan
			nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
				DeviceProfile: nsCreate.DeviceProfile,
			}

			Convey("Then Get returns the device-profile", func() {
				getResp, err := api.Get(ctx, &pb.GetDeviceProfileRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)

				createReq.DeviceProfile.Id = createResp.Id
				So(getResp.DeviceProfile, ShouldResemble, createReq.DeviceProfile)
			})

			Convey("Then Update updates the device-profile", func() {
				updateReq := pb.UpdateDeviceProfileRequest{
					DeviceProfile: &pb.DeviceProfile{
						Id:                 createResp.Id,
						OrganizationId:     org.ID,
						NetworkServerId:    n.ID,
						Name:               "updated-dp",
						SupportsClassB:     true,
						ClassBTimeout:      20,
						PingSlotPeriod:     30,
						PingSlotDr:         4,
						PingSlotFreq:       868300000,
						SupportsClassC:     true,
						ClassCTimeout:      20,
						MacVersion:         "1.1.0",
						RegParamsRevision:  "C",
						RxDelay_1:          2,
						RxDrOffset_1:       3,
						RxDatarate_2:       5,
						RxFreq_2:           868500000,
						FactoryPresetFreqs: []uint32{868100000, 868300000, 868500000, 868700000},
						MaxEirp:            17,
						MaxDutyCycle:       1,
						SupportsJoin:       true,
						RfRegion:           "EU868",
						Supports_32BitFCnt: true,
					},
				}

				_, err := api.Update(ctx, &updateReq)
				So(err, ShouldBeNil)
				So(nsClient.UpdateDeviceProfileChan, ShouldHaveLength, 1)

				nsUpdate := <-nsClient.UpdateDeviceProfileChan
				nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
					DeviceProfile: nsUpdate.DeviceProfile,
				}

				getResp, err := api.Get(ctx, &pb.GetDeviceProfileRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)
				So(getResp.DeviceProfile, ShouldResemble, updateReq.DeviceProfile)
			})

			Convey("Then Delete deletes the device-profile", func() {
				_, err := api.Delete(ctx, &pb.DeleteDeviceProfileRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)

				_, err = api.Get(ctx, &pb.GetDeviceProfileRequest{
					Id: createResp.Id,
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
				userID, err := storage.CreateUser(storage.DB(), &storage.User{
					Username: "testuser",
					IsActive: true,
					Email:    "foo@bar.com",
				}, "testpassword")
				So(err, ShouldBeNil)
				So(storage.CreateOrganizationUser(storage.DB(), org.ID, userID, false), ShouldBeNil)

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
					OrganizationId: org.ID,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 1)
				So(listResp.Result, ShouldHaveLength, 1)

				listResp, err = api.List(ctx, &pb.ListDeviceProfileRequest{
					Limit:          10,
					OrganizationId: org.ID + 1,
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
				So(storage.CreateNetworkServer(storage.DB(), &n2), ShouldBeNil)

				sp1 := storage.ServiceProfile{
					Name:            "test-sp",
					NetworkServerID: n.ID,
					OrganizationID:  org.ID,
				}
				So(storage.CreateServiceProfile(storage.DB(), &sp1), ShouldBeNil)
				sp1ID, err := uuid.FromBytes(sp1.ServiceProfile.Id)
				So(err, ShouldBeNil)

				sp2 := storage.ServiceProfile{
					Name:            "test-sp-2",
					NetworkServerID: n2.ID,
					OrganizationID:  org.ID,
				}
				So(storage.CreateServiceProfile(storage.DB(), &sp2), ShouldBeNil)
				sp2ID, err := uuid.FromBytes(sp2.ServiceProfile.Id)
				So(err, ShouldBeNil)

				app1 := storage.Application{
					Name:             "test-app",
					Description:      "test app",
					OrganizationID:   org.ID,
					ServiceProfileID: sp1ID,
				}
				So(storage.CreateApplication(storage.DB(), &app1), ShouldBeNil)

				app2 := storage.Application{
					Name:             "test-app-2",
					Description:      "test app 2",
					OrganizationID:   org.ID,
					ServiceProfileID: sp2ID,
				}
				So(storage.CreateApplication(storage.DB(), &app2), ShouldBeNil)

				Convey("Then List filtered on application id returns the device-profiles accessible by the given application id", func() {
					listResp, err := api.List(ctx, &pb.ListDeviceProfileRequest{
						Limit:         10,
						ApplicationId: app1.ID,
					})
					So(err, ShouldBeNil)
					So(listResp.TotalCount, ShouldEqual, 1)
					So(listResp.Result, ShouldHaveLength, 1)

					So(listResp.Result[0].Id, ShouldEqual, createResp.Id)

					listResp, err = api.List(ctx, &pb.ListDeviceProfileRequest{
						Limit:         10,
						ApplicationId: app2.ID,
					})
					So(err, ShouldBeNil)
					So(listResp.TotalCount, ShouldEqual, 0)
					So(listResp.Result, ShouldHaveLength, 0)
				})
			})
		})
	})
}
