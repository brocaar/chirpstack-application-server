package external

import (
	"testing"

	uuid "github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
)

func (ts *APITestSuite) TestDeviceProfile() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	validator := &TestValidator{}
	api := NewDeviceProfileServiceAPI(validator)

	n := storage.NetworkServer{
		Name:   "test",
		Server: "test:1234",
	}
	assert.NoError(storage.CreateNetworkServer(storage.DB(), &n))

	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(storage.DB(), &org))

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		createReq := pb.CreateDeviceProfileRequest{
			DeviceProfile: &pb.DeviceProfile{
				Name:                 "test-dp",
				OrganizationId:       org.ID,
				NetworkServerId:      n.ID,
				SupportsClassB:       true,
				ClassBTimeout:        10,
				PingSlotPeriod:       20,
				PingSlotDr:           5,
				PingSlotFreq:         868100000,
				SupportsClassC:       true,
				ClassCTimeout:        30,
				MacVersion:           "1.0.2",
				RegParamsRevision:    "B",
				RxDelay_1:            1,
				RxDrOffset_1:         1,
				RxDatarate_2:         6,
				RxFreq_2:             868300000,
				FactoryPresetFreqs:   []uint32{868100000, 868300000, 868500000},
				MaxEirp:              14,
				MaxDutyCycle:         10,
				SupportsJoin:         true,
				RfRegion:             "EU868",
				Supports_32BitFCnt:   true,
				PayloadCodec:         "CUSTOM_JS",
				PayloadEncoderScript: "Encode() {}",
				PayloadDecoderScript: "Decode() {}",
			},
		}

		createResp, err := api.Create(context.Background(), &createReq)
		assert.NoError(err)

		assert.NotEqual("", createResp.Id)
		assert.NotEqual(uuid.Nil.String(), "")

		// set network-server mock
		nsCreate := <-nsClient.CreateDeviceProfileChan
		nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
			DeviceProfile: nsCreate.DeviceProfile,
		}

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			getResp, err := api.Get(context.Background(), &pb.GetDeviceProfileRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			createReq.DeviceProfile.Id = createResp.Id
			assert.Equal(createReq.DeviceProfile, getResp.DeviceProfile)
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)
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

			_, err := api.Update(context.Background(), &updateReq)
			assert.NoError(err)

			nsUpdate := <-nsClient.UpdateDeviceProfileChan
			nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
				DeviceProfile: nsUpdate.DeviceProfile,
			}

			getResp, err := api.Get(context.Background(), &pb.GetDeviceProfileRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)
			assert.Equal(updateReq.DeviceProfile, getResp.DeviceProfile)
		})

		t.Run("Global admin user", func(t *testing.T) {
			validator.returnIsAdmin = true

			t.Run("List", func(t *testing.T) {
				assert := require.New(t)
				listResp, err := api.List(context.Background(), &pb.ListDeviceProfileRequest{
					Limit: 10,
				})
				assert.NoError(err)
				assert.EqualValues(1, listResp.TotalCount)
				assert.Len(listResp.Result, 1)
			})

			t.Run("List with org id", func(t *testing.T) {
				assert := require.New(t)

				listResp, err := api.List(context.Background(), &pb.ListDeviceProfileRequest{
					Limit:          10,
					OrganizationId: org.ID,
				})
				assert.NoError(err)
				assert.EqualValues(1, listResp.TotalCount)
				assert.Len(listResp.Result, 1)

				listResp, err = api.List(context.Background(), &pb.ListDeviceProfileRequest{
					Limit:          10,
					OrganizationId: org.ID + 1,
				})
				assert.NoError(err)
				assert.EqualValues(0, listResp.TotalCount)
				assert.Len(listResp.Result, 0)
			})

			t.Run("List with app id", func(t *testing.T) {
				// we test here that only the device-profiles that are indirectly
				// linked through the application service-profile are returned

				assert := require.New(t)

				n2 := storage.NetworkServer{
					Name:   "ns-server-2",
					Server: "ns-server-2:1234",
				}
				assert.NoError(storage.CreateNetworkServer(storage.DB(), &n2))

				sp1 := storage.ServiceProfile{
					Name:            "test-sp",
					NetworkServerID: n.ID,
					OrganizationID:  org.ID,
				}
				assert.NoError(storage.CreateServiceProfile(storage.DB(), &sp1))
				sp1ID, err := uuid.FromBytes(sp1.ServiceProfile.Id)
				assert.NoError(err)

				sp2 := storage.ServiceProfile{
					Name:            "test-sp-2",
					NetworkServerID: n2.ID,
					OrganizationID:  org.ID,
				}
				assert.NoError(storage.CreateServiceProfile(storage.DB(), &sp2))
				sp2ID, err := uuid.FromBytes(sp2.ServiceProfile.Id)
				assert.NoError(err)

				app1 := storage.Application{
					Name:             "test-app",
					Description:      "test app",
					OrganizationID:   org.ID,
					ServiceProfileID: sp1ID,
				}
				assert.NoError(storage.CreateApplication(storage.DB(), &app1))

				app2 := storage.Application{
					Name:             "test-app-2",
					Description:      "test app 2",
					OrganizationID:   org.ID,
					ServiceProfileID: sp2ID,
				}
				assert.NoError(storage.CreateApplication(storage.DB(), &app2))

				listResp, err := api.List(context.Background(), &pb.ListDeviceProfileRequest{
					Limit:         10,
					ApplicationId: app1.ID,
				})
				assert.NoError(err)
				assert.EqualValues(1, listResp.TotalCount)
				assert.Len(listResp.Result, 1)
				assert.Equal(createResp.Id, listResp.Result[0].Id)

				listResp, err = api.List(context.Background(), &pb.ListDeviceProfileRequest{
					Limit:         10,
					ApplicationId: app2.ID,
				})
				assert.NoError(err)
				assert.EqualValues(0, listResp.TotalCount)
				assert.Len(listResp.Result, 0)
			})
		})

		t.Run("Organization user", func(t *testing.T) {
			assert := require.New(t)

			userID, err := storage.CreateUser(storage.DB(), &storage.User{
				Username: "testuser",
				IsActive: true,
				Email:    "foo@bar.com",
			}, "testpassword")
			assert.NoError(err)

			assert.NoError(storage.CreateOrganizationUser(storage.DB(), org.ID, userID, false))

			t.Run("List without org id returns all device-profiles for user", func(t *testing.T) {
				assert := require.New(t)

				validator.returnIsAdmin = false
				validator.returnUsername = "testuser"

				listResp, err := api.List(context.Background(), &pb.ListDeviceProfileRequest{
					Limit: 10,
				})
				assert.NoError(err)
				assert.EqualValues(1, listResp.TotalCount)
				assert.Len(listResp.Result, 1)
			})

			t.Run("List with different user", func(t *testing.T) {
				assert := require.New(t)

				validator.returnIsAdmin = false
				validator.returnUsername = "differentuser"

				listResp, err := api.List(context.Background(), &pb.ListDeviceProfileRequest{
					Limit: 10,
				})
				assert.NoError(err)
				assert.EqualValues(0, listResp.TotalCount)
				assert.Len(listResp.Result, 0)
			})
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			_, err := api.Delete(context.Background(), &pb.DeleteDeviceProfileRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			_, err = api.Get(context.Background(), &pb.GetDeviceProfileRequest{
				Id: createResp.Id,
			})
			assert.NotNil(err)
			assert.Equal(codes.NotFound, grpc.Code(err))
		})
	})
}
