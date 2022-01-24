package external

import (
	"testing"

	"github.com/brocaar/lorawan"

	uuid "github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

func (ts *APITestSuite) TestMulticastGroup() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	rpID, _ := uuid.NewV4()

	validator := &TestValidator{}
	api := NewMulticastGroupAPI(validator, rpID)

	n := storage.NetworkServer{
		Name:   "test",
		Server: "test:1234",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &n))

	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))

	sp := storage.ServiceProfile{
		Name:            "test-sp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp))
	var spID uuid.UUID
	copy(spID[:], sp.ServiceProfile.Id)

	app := storage.Application{
		Name:           "test-app",
		OrganizationID: org.ID,
	}
	copy(app.ServiceProfileID[:], sp.ServiceProfile.Id)
	assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &app))

	dp := storage.DeviceProfile{
		Name:            "test-dp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(storage.CreateDeviceProfile(context.Background(), storage.DB(), &dp))
	var dpID uuid.UUID
	copy(dpID[:], dp.DeviceProfile.Id)

	adminUser := storage.User{
		Email:    "admin@user.com",
		IsAdmin:  true,
		IsActive: true,
	}
	assert.NoError(storage.CreateUser(context.Background(), storage.DB(), &adminUser))

	user := storage.User{
		Email:    "some@user.com",
		IsAdmin:  false,
		IsActive: true,
	}
	assert.NoError(storage.CreateUser(context.Background(), storage.DB(), &user))

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		createReq := pb.CreateMulticastGroupRequest{
			MulticastGroup: &pb.MulticastGroup{
				Name:           "test-mg",
				McAddr:         "01020304",
				McNwkSKey:      "01020304050607080102030405060708",
				McAppSKey:      "08070605040302010807060504030201",
				FCnt:           10,
				GroupType:      pb.MulticastGroupType_CLASS_B,
				Dr:             5,
				Frequency:      868100000,
				PingSlotPeriod: 32,
				ApplicationId:  app.ID,
			},
		}

		createResp, err := api.Create(context.Background(), &createReq)
		assert.NoError(err)
		assert.NotEqual(uuid.Nil.String(), createResp.Id)
		createReq.MulticastGroup.Id = createResp.Id

		nsCreateReq := <-nsClient.CreateMulticastGroupChan
		nsClient.GetMulticastGroupResponse = ns.GetMulticastGroupResponse{
			MulticastGroup: nsCreateReq.MulticastGroup,
		}

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			getResp, err := api.Get(context.Background(), &pb.GetMulticastGroupRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)
			assert.Equal(createReq.MulticastGroup, getResp.MulticastGroup)
			assert.NotEqual("", getResp.CreatedAt)
			assert.NotEqual("", getResp.UpdatedAt)
		})

		t.Run("List", func(t *testing.T) {
			assert := require.New(t)

			mgID, err := uuid.FromString(createResp.Id)
			assert.NoError(err)

			// assign device to multicast-group for listing
			d := storage.Device{
				DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				ApplicationID:   app.ID,
				DeviceProfileID: dpID,
				Name:            "test-device-1",
			}
			assert.NoError(storage.CreateDevice(context.Background(), storage.DB(), &d))
			assert.NoError(storage.AddDeviceToMulticastGroup(context.Background(), storage.DB(), mgID, d.DevEUI))

			testTable := []struct {
				Name             string
				User             storage.User
				Request          pb.ListMulticastGroupRequest
				ExpectedResponse pb.ListMulticastGroupResponse
				ExpectedError    error
			}{
				{
					Name: "no admin, list all",
					Request: pb.ListMulticastGroupRequest{
						Limit: 10,
					},
					ExpectedError: grpc.Errorf(codes.Unauthenticated, "client must be global admin for unfiltered request"),
				},
				{
					Name: "admin, list all",
					User: adminUser,
					Request: pb.ListMulticastGroupRequest{
						Limit: 10,
					},
					ExpectedResponse: pb.ListMulticastGroupResponse{
						TotalCount: 1,
						Result: []*pb.MulticastGroupListItem{
							{
								Id:              createResp.Id,
								Name:            createReq.MulticastGroup.Name,
								ApplicationId:   app.ID,
								ApplicationName: app.Name,
							},
						},
					},
				},
				{
					Name: "non-matching search",
					User: adminUser,
					Request: pb.ListMulticastGroupRequest{
						Limit:  10,
						Search: "testing",
					},
					ExpectedResponse: pb.ListMulticastGroupResponse{
						TotalCount: 0,
					},
				},
				{
					Name: "matching search",
					User: adminUser,
					Request: pb.ListMulticastGroupRequest{
						Limit:  10,
						Search: "tes",
					},
					ExpectedResponse: pb.ListMulticastGroupResponse{
						TotalCount: 1,
						Result: []*pb.MulticastGroupListItem{
							{
								Id:              createResp.Id,
								Name:            createReq.MulticastGroup.Name,
								ApplicationId:   app.ID,
								ApplicationName: app.Name,
							},
						},
					},
				},
				{
					Name: "non-matching org id",
					User: user,
					Request: pb.ListMulticastGroupRequest{
						Limit:          10,
						OrganizationId: org.ID + 1,
					},
					ExpectedResponse: pb.ListMulticastGroupResponse{
						TotalCount: 0,
					},
				},
				{
					Name: "matching org id",
					User: user,
					Request: pb.ListMulticastGroupRequest{
						Limit:          10,
						OrganizationId: org.ID,
					},
					ExpectedResponse: pb.ListMulticastGroupResponse{
						TotalCount: 1,
						Result: []*pb.MulticastGroupListItem{
							{
								Id:              createResp.Id,
								Name:            createReq.MulticastGroup.Name,
								ApplicationId:   app.ID,
								ApplicationName: app.Name,
							},
						},
					},
				},
				{
					Name: "non-matching application id",
					User: user,
					Request: pb.ListMulticastGroupRequest{
						Limit:         10,
						ApplicationId: app.ID + 1,
					},
					ExpectedResponse: pb.ListMulticastGroupResponse{
						TotalCount: 0,
					},
				},
				{
					Name: "matching application id",
					User: user,
					Request: pb.ListMulticastGroupRequest{
						Limit:         10,
						ApplicationId: app.ID,
					},
					ExpectedResponse: pb.ListMulticastGroupResponse{
						TotalCount: 1,
						Result: []*pb.MulticastGroupListItem{
							{
								Id:              createResp.Id,
								Name:            createReq.MulticastGroup.Name,
								ApplicationId:   app.ID,
								ApplicationName: app.Name,
							},
						},
					},
				},
				{
					Name: "non-matching deveui",
					User: user,
					Request: pb.ListMulticastGroupRequest{
						Limit:  10,
						DevEui: "0807060504030201",
					},
					ExpectedResponse: pb.ListMulticastGroupResponse{
						TotalCount: 0,
					},
				},
				{
					Name: "non-matching deveui",
					User: user,
					Request: pb.ListMulticastGroupRequest{
						Limit:  10,
						DevEui: "0102030405060708",
					},
					ExpectedResponse: pb.ListMulticastGroupResponse{
						TotalCount: 1,
						Result: []*pb.MulticastGroupListItem{
							{
								Id:              createResp.Id,
								Name:            createReq.MulticastGroup.Name,
								ApplicationId:   app.ID,
								ApplicationName: app.Name,
							},
						},
					},
				},
			}

			for _, test := range testTable {
				t.Run(test.Name, func(t *testing.T) {
					assert := require.New(t)

					validator.returnUser = test.User

					resp, err := api.List(context.Background(), &test.Request)
					assert.Equal(test.ExpectedError, err)

					if err == nil {
						assert.Equal(&test.ExpectedResponse, resp)
					}
				})
			}
		})

		t.Run("Add device", func(t *testing.T) {
			assert := require.New(t)

			// assign device to multicast-group for listing
			d := storage.Device{
				DevEUI:          lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
				ApplicationID:   app.ID,
				DeviceProfileID: dpID,
				Name:            "test-device-2",
			}
			assert.NoError(storage.CreateDevice(context.Background(), storage.DB(), &d))

			_, err := api.AddDevice(context.Background(), &pb.AddDeviceToMulticastGroupRequest{
				DevEui:           d.DevEUI.String(),
				MulticastGroupId: createResp.Id,
			})
			assert.NoError(err)

			t.Run("Remove device", func(t *testing.T) {
				assert := require.New(t)

				_, err := api.RemoveDevice(context.Background(), &pb.RemoveDeviceFromMulticastGroupRequest{
					DevEui:           d.DevEUI.String(),
					MulticastGroupId: createResp.Id,
				})
				assert.NoError(err)

				_, err = api.RemoveDevice(context.Background(), &pb.RemoveDeviceFromMulticastGroupRequest{
					DevEui:           d.DevEUI.String(),
					MulticastGroupId: createResp.Id,
				})
				assert.Error(err)
				assert.Equal(codes.NotFound, grpc.Code(err))
			})
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			updateReq := pb.UpdateMulticastGroupRequest{
				MulticastGroup: &pb.MulticastGroup{
					Id:             createResp.Id,
					Name:           "test-mg-updated",
					McAddr:         "04030201",
					McAppSKey:      "01020304050607080102030405060708",
					McNwkSKey:      "08070605040302010807060504030201",
					FCnt:           11,
					GroupType:      pb.MulticastGroupType_CLASS_C,
					Dr:             4,
					Frequency:      868300000,
					PingSlotPeriod: 64,
					ApplicationId:  app.ID,
				},
			}

			_, err := api.Update(context.Background(), &updateReq)
			assert.NoError(err)

			nsUpdateReq := <-nsClient.UpdateMulticastGroupChan
			nsClient.GetMulticastGroupResponse = ns.GetMulticastGroupResponse{
				MulticastGroup: nsUpdateReq.MulticastGroup,
			}

			getResp, err := api.Get(context.Background(), &pb.GetMulticastGroupRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)
			assert.Equal(updateReq.MulticastGroup, getResp.MulticastGroup)
		})

		t.Run("Enqueue", func(t *testing.T) {
			assert := require.New(t)
			mgID, err := uuid.FromString(createResp.Id)
			assert.NoError(err)

			nsClient.GetMulticastGroupResponse = ns.GetMulticastGroupResponse{
				MulticastGroup: &ns.MulticastGroup{
					McAddr: []byte{1, 2, 3, 4},
					FCnt:   15,
				},
			}

			resp, err := api.Enqueue(context.Background(), &pb.EnqueueMulticastQueueItemRequest{
				MulticastQueueItem: &pb.MulticastQueueItem{
					MulticastGroupId: createResp.Id,
					FPort:            10,
					Data:             []byte{1, 2, 3, 4, 5},
				},
			})
			assert.NoError(err)
			assert.EqualValues(15, resp.FCnt)

			nsEnqueueReq := <-nsClient.EnqueueMulticastQueueItemChan
			assert.Equal(&ns.MulticastQueueItem{
				MulticastGroupId: mgID.Bytes(),
				FCnt:             15,
				FPort:            10,
				FrmPayload:       []byte{0x3f, 0xb1, 0xca, 0xb2, 0xc7},
			}, nsEnqueueReq.MulticastQueueItem)

			t.Run("ListQueue", func(t *testing.T) {
				assert := require.New(t)

				nsClient.GetMulticastQueueItemsForMulticastGroupResponse = ns.GetMulticastQueueItemsForMulticastGroupResponse{
					MulticastQueueItems: []*ns.MulticastQueueItem{
						nsEnqueueReq.MulticastQueueItem,
					},
				}

				listResp, err := api.ListQueue(context.Background(), &pb.ListMulticastGroupQueueItemsRequest{
					MulticastGroupId: mgID.String(),
				})
				assert.NoError(err)
				assert.Len(listResp.MulticastQueueItems, 1)
				assert.Equal(&pb.MulticastQueueItem{
					MulticastGroupId: createResp.Id,
					FCnt:             15,
					FPort:            10,
					Data:             []byte{1, 2, 3, 4, 5},
				}, listResp.MulticastQueueItems[0])

				nsReq := <-nsClient.GetMulticastQueueItemsForMulticastGroupChan
				assert.Equal(ns.GetMulticastQueueItemsForMulticastGroupRequest{
					MulticastGroupId: mgID.Bytes(),
				}, nsReq)
			})

			t.Run("FlushQueue", func(t *testing.T) {
				assert := require.New(t)

				_, err := api.FlushQueue(context.Background(), &pb.FlushMulticastGroupQueueItemsRequest{
					MulticastGroupId: mgID.String(),
				})
				assert.NoError(err)

				nsReq := <-nsClient.FlushMulticastQueueForMulticastGroupChan
				assert.Equal(ns.FlushMulticastQueueForMulticastGroupRequest{
					MulticastGroupId: mgID.Bytes(),
				}, nsReq)

			})
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			_, err := api.Delete(context.Background(), &pb.DeleteMulticastGroupRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			_, err = api.Get(context.Background(), &pb.GetMulticastGroupRequest{
				Id: createResp.Id,
			})
			assert.Error(err)
			assert.Equal(codes.NotFound, grpc.Code(err))
		})
	})
}
