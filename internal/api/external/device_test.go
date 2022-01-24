package external

import (
	"net"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/eventlog"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

func (ts *APITestSuite) TestDevice() {
	assert := require.New(ts.T())

	assert.NoError(storage.SetAggregationIntervals([]storage.AggregationInterval{storage.AggregationMinute}))
	storage.SetMetricsTTL(time.Minute, time.Minute, time.Minute, time.Minute)

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	validator := &TestValidator{}

	grpcServer := grpc.NewServer()
	apiServer := NewDeviceAPI(validator)
	pb.RegisterDeviceServiceServer(grpcServer, apiServer)

	ln, err := net.Listen("tcp", "localhost:0")
	assert.NoError(err)
	go grpcServer.Serve(ln)
	defer func() {
		grpcServer.Stop()
		ln.Close()
	}()

	apiClient, err := grpc.Dial(ln.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
	assert.NoError(err)
	defer apiClient.Close()

	api := pb.NewDeviceServiceClient(apiClient)

	org := storage.Organization{
		Name:           "test-org",
		MaxDeviceCount: 1,
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))
	org2 := storage.Organization{
		Name: "test-org-2",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org2))

	n := storage.NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:1234",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &n))

	sp := storage.ServiceProfile{
		Name:            "test-sp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp))
	spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
	assert.NoError(err)

	sp2 := storage.ServiceProfile{
		Name:            "test-sp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp2))
	sp2ID, err := uuid.FromBytes(sp2.ServiceProfile.Id)
	assert.NoError(err)

	app := storage.Application{
		OrganizationID:   org.ID,
		Name:             "test-app",
		ServiceProfileID: spID,
	}
	assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &app))

	app2 := storage.Application{
		OrganizationID:   org.ID,
		Name:             "test-app-2",
		ServiceProfileID: spID,
	}
	assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &app2))

	app3 := storage.Application{
		OrganizationID:   org.ID,
		Name:             "test-app-3",
		ServiceProfileID: sp2ID,
	}
	assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &app3))

	dp := storage.DeviceProfile{
		Name:            "test-dp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(storage.CreateDeviceProfile(context.Background(), storage.DB(), &dp))
	dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
	assert.NoError(err)
	dp2 := storage.DeviceProfile{
		Name:            "test-dp",
		OrganizationID:  org2.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(storage.CreateDeviceProfile(context.Background(), storage.DB(), &dp2))
	dpID2, err := uuid.FromBytes(dp2.DeviceProfile.Id)
	assert.NoError(err)

	adminUser := storage.User{
		Email:    "admin@user.com",
		IsActive: true,
		IsAdmin:  true,
	}
	assert.NoError(storage.CreateUser(context.Background(), storage.DB(), &adminUser))

	user := storage.User{
		Email:    "some@user.com",
		IsActive: true,
		IsAdmin:  false,
	}
	assert.NoError(storage.CreateUser(context.Background(), storage.DB(), &user))

	ts.T().Run("Create without name", func(t *testing.T) {
		assert := require.New(t)

		_, err := api.Create(context.Background(), &pb.CreateDeviceRequest{
			Device: &pb.Device{
				DevEui:          "0807060504030202",
				ApplicationId:   app.ID,
				Description:     "test device description",
				DeviceProfileId: dpID.String(),
			},
		})
		assert.NoError(err)

		nsReq := <-nsClient.CreateDeviceChan
		nsClient.GetDeviceResponse = ns.GetDeviceResponse{
			Device: nsReq.Device,
		}

		dGet, err := api.Get(context.Background(), &pb.GetDeviceRequest{
			DevEui: "0807060504030202",
		})
		assert.NoError(err)
		assert.Equal("0807060504030202", dGet.Device.Name)

		_, err = api.Delete(context.Background(), &pb.DeleteDeviceRequest{
			DevEui: "0807060504030202",
		})
		assert.NoError(err)
	})

	ts.T().Run("Create with device-profile under different organization", func(t *testing.T) {
		assert := require.New(t)

		createReq := pb.CreateDeviceRequest{
			Device: &pb.Device{
				ApplicationId:     app.ID,
				Name:              "test-device",
				Description:       "test device description",
				DevEui:            "0807060504030201",
				DeviceProfileId:   dpID2.String(),
				SkipFCntCheck:     true,
				ReferenceAltitude: 5.6,
				Variables: map[string]string{
					"var_1": "test var 1",
				},
				Tags: map[string]string{
					"foo": "bar",
				},
			},
		}
		_, err := api.Create(context.Background(), &createReq)
		assert.Equal(codes.InvalidArgument, grpc.Code(err))
	})

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		createReq := pb.CreateDeviceRequest{
			Device: &pb.Device{
				ApplicationId:     app.ID,
				Name:              "test-device",
				Description:       "test device description",
				DevEui:            "0807060504030201",
				DeviceProfileId:   dpID.String(),
				SkipFCntCheck:     true,
				ReferenceAltitude: 5.6,
				IsDisabled:        true,
				Variables: map[string]string{
					"var_1": "test var 1",
				},
				Tags: map[string]string{
					"foo": "bar",
				},
			},
		}
		_, err := api.Create(context.Background(), &createReq)
		assert.NoError(err)

		nsReq := <-nsClient.CreateDeviceChan
		nsClient.GetDeviceResponse = ns.GetDeviceResponse{
			Device: nsReq.Device,
		}

		t.Run("Create second exceeds max device count", func(t *testing.T) {
			assert := require.New(t)

			createReq := pb.CreateDeviceRequest{
				Device: &pb.Device{
					ApplicationId:   app.ID,
					Name:            "test-device-2",
					Description:     "test device description",
					DevEui:          "0807060504030202",
					DeviceProfileId: dpID.String(),
				},
			}
			_, err := api.Create(context.Background(), &createReq)
			assert.Equal(codes.FailedPrecondition, grpc.Code(err))
			assert.Equal("rpc error: code = FailedPrecondition desc = organization reached max. device count", err.Error())
		})

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			d, err := api.Get(context.Background(), &pb.GetDeviceRequest{
				DevEui: "0807060504030201",
			})
			assert.NoError(err)
			assert.True(proto.Equal(createReq.Device, d.Device))
			assert.True(proto.Equal(createReq.Device, d.Device))
			assert.Nil(d.LastSeenAt)
			assert.EqualValues(256, d.DeviceStatusBattery)
			assert.EqualValues(256, d.DeviceStatusMargin)

			t.Run("Set battery and margin status", func(t *testing.T) {
				assert := require.New(t)

				ten := float32(10)
				eleven := 11

				d, err := storage.GetDevice(context.Background(), storage.DB(), lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, false, true)
				assert.NoError(err)

				d.DeviceStatusBattery = &ten
				d.DeviceStatusMargin = &eleven
				assert.NoError(storage.UpdateDevice(context.Background(), storage.DB(), &d, true))

				t.Run("Get", func(t *testing.T) {
					assert := require.New(t)

					d, err := api.Get(context.Background(), &pb.GetDeviceRequest{
						DevEui: "0807060504030201",
					})
					assert.NoError(err)
					assert.EqualValues(10, d.DeviceStatusBattery)
					assert.EqualValues(11, d.DeviceStatusMargin)
				})
			})

			t.Run("Set LastSeenAt", func(t *testing.T) {
				assert := require.New(t)

				now := time.Now().Truncate(time.Millisecond)

				d, err := storage.GetDevice(context.Background(), storage.DB(), lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, false, true)
				assert.NoError(err)
				d.LastSeenAt = &now
				assert.NoError(storage.UpdateDevice(context.Background(), storage.DB(), &d, true))

				t.Run("Get", func(t *testing.T) {
					d, err := api.Get(context.Background(), &pb.GetDeviceRequest{
						DevEui: "0807060504030201",
					})
					assert.NoError(err)
					assert.NotNil(d.LastSeenAt)
				})
			})

			t.Run("Set location", func(t *testing.T) {
				assert := require.New(t)

				lat := 1.123
				long := 2.123
				alt := 3.123

				d, err := storage.GetDevice(context.Background(), storage.DB(), lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, false, true)
				assert.NoError(err)
				d.Latitude = &lat
				d.Longitude = &long
				d.Altitude = &alt
				assert.NoError(storage.UpdateDevice(context.Background(), storage.DB(), &d, true))

				t.Run("Get", func(t *testing.T) {
					assert := require.New(t)

					d, err := api.Get(context.Background(), &pb.GetDeviceRequest{
						DevEui: "0807060504030201",
					})
					assert.NoError(err)
					assert.Equal(&common.Location{
						Latitude:  1.123,
						Longitude: 2.123,
						Altitude:  3.123,
					}, d.Location)
				})
			})

			t.Run("List", func(t *testing.T) {
				t.Run("Filter by tag", func(t *testing.T) {
					assert := require.New(t)
					validator.returnUser = adminUser

					devices, err := api.List(context.Background(), &pb.ListDeviceRequest{
						Limit:  10,
						Offset: 0,
						Tags:   map[string]string{"foo": "bar"},
					})
					assert.NoError(err)
					assert.EqualValues(1, devices.TotalCount)
					assert.Len(devices.Result, 1)

					devices, err = api.List(context.Background(), &pb.ListDeviceRequest{
						Limit:  10,
						Offset: 0,
						Tags:   map[string]string{"foo": "bas"},
					})
					assert.NoError(err)
					assert.EqualValues(0, devices.TotalCount)
					assert.Len(devices.Result, 0)
				})

				t.Run("Global admin can list all devices", func(t *testing.T) {
					assert := require.New(t)
					validator.returnUser = adminUser

					devices, err := api.List(context.Background(), &pb.ListDeviceRequest{
						Limit:  10,
						Offset: 0,
					})
					assert.NoError(err)
					assert.EqualValues(1, devices.TotalCount)
					assert.Len(devices.Result, 1)

					devices, err = api.List(context.Background(), &pb.ListDeviceRequest{
						Limit:         10,
						Offset:        0,
						ApplicationId: app.ID,
					})
					assert.NoError(err)
					assert.EqualValues(1, devices.TotalCount)
					assert.Len(devices.Result, 1)
				})

				t.Run("Non-admin can not list the devices", func(t *testing.T) {
					assert := require.New(t)
					validator.returnUser = user

					_, err := api.List(context.Background(), &pb.ListDeviceRequest{
						Limit:  10,
						Offset: 0,
					})
					assert.NotNil(err)
				})

				t.Run("Non-admin can list devices by application id", func(t *testing.T) {
					assert := require.New(t)
					validator.returnUser = user

					devices, err := api.List(context.Background(), &pb.ListDeviceRequest{
						Limit:         10,
						Offset:        0,
						ApplicationId: app.ID,
					})
					assert.NoError(err)
					assert.EqualValues(1, devices.TotalCount)
					assert.Len(devices.Result, 1)
				})
			})

			t.Run("Update with device-profile under different organization", func(t *testing.T) {
				assert := require.New(t)

				updateReq := pb.UpdateDeviceRequest{
					Device: &pb.Device{
						ApplicationId:     app.ID,
						DevEui:            "0807060504030201",
						Name:              "test-device-updated",
						Description:       "test device description updated",
						DeviceProfileId:   dpID2.String(),
						SkipFCntCheck:     true,
						ReferenceAltitude: 6.7,
						Variables: map[string]string{
							"var_2": "test var 2",
						},
						Tags: map[string]string{
							"bar": "foo",
						},
					},
				}

				_, err := api.Update(context.Background(), &updateReq)
				assert.Equal(codes.InvalidArgument, grpc.Code(err))
			})

			t.Run("Update", func(t *testing.T) {
				assert := require.New(t)

				updateReq := pb.UpdateDeviceRequest{
					Device: &pb.Device{
						ApplicationId:     app.ID,
						DevEui:            "0807060504030201",
						Name:              "test-device-updated",
						Description:       "test device description updated",
						DeviceProfileId:   dpID.String(),
						SkipFCntCheck:     true,
						ReferenceAltitude: 6.7,
						IsDisabled:        true,
						Variables: map[string]string{
							"var_2": "test var 2",
						},
						Tags: map[string]string{
							"bar": "foo",
						},
					},
				}

				_, err := api.Update(context.Background(), &updateReq)
				assert.NoError(err)

				nsUpdateReq := <-nsClient.UpdateDeviceChan
				nsClient.GetDeviceResponse = ns.GetDeviceResponse{
					Device: nsUpdateReq.Device,
				}

				d, err := api.Get(context.Background(), &pb.GetDeviceRequest{
					DevEui: "0807060504030201",
				})
				assert.NoError(err)
				assert.True(proto.Equal(updateReq.Device, d.Device))
			})

			t.Run("Update and move to different application", func(t *testing.T) {
				t.Run("Same service-profile", func(t *testing.T) {
					assert := require.New(t)

					updateReq := pb.UpdateDeviceRequest{
						Device: &pb.Device{
							ApplicationId:     app2.ID,
							DevEui:            "0807060504030201",
							Name:              "test-device-updated",
							Description:       "test device description updated",
							DeviceProfileId:   dpID.String(),
							SkipFCntCheck:     true,
							ReferenceAltitude: 6.7,
						},
					}

					_, err := api.Update(context.Background(), &updateReq)
					assert.NoError(err)

					<-nsClient.UpdateDeviceChan

					d, err := api.Get(context.Background(), &pb.GetDeviceRequest{
						DevEui: "0807060504030201",
					})
					assert.NoError(err)
					assert.Equal(app2.ID, d.Device.ApplicationId)
				})

				t.Run("Different service-profile", func(t *testing.T) {
					assert := require.New(t)

					updateReq := pb.UpdateDeviceRequest{
						Device: &pb.Device{
							ApplicationId:     app3.ID,
							DevEui:            "0807060504030201",
							Name:              "test-device-updated",
							Description:       "test device description updated",
							DeviceProfileId:   dpID.String(),
							SkipFCntCheck:     true,
							ReferenceAltitude: 6.7,
						},
					}

					_, err := api.Update(context.Background(), &updateReq)
					assert.Equal(codes.InvalidArgument, grpc.Code(err))
					assert.Error(err, "rpc error: code = InvalidArgument desc = when moving a device from application A to B, both A and B must share the same service-profile")
				})
			})

			t.Run("CreateKeys", func(t *testing.T) {
				assert := require.New(t)

				createReq := pb.CreateDeviceKeysRequest{
					DeviceKeys: &pb.DeviceKeys{
						DevEui: "0807060504030201",
						NwkKey: "01020304050607080807060504030201",
					},
				}
				_, err := api.CreateKeys(context.Background(), &createReq)
				assert.NoError(err)

				t.Run("GetKeys", func(t *testing.T) {
					assert := require.New(t)

					dk, err := api.GetKeys(context.Background(), &pb.GetDeviceKeysRequest{
						DevEui: "0807060504030201",
					})
					assert.NoError(err)
					assert.Equal(&pb.DeviceKeys{
						DevEui: "0807060504030201",
						NwkKey: "01020304050607080807060504030201",
						AppKey: "00000000000000000000000000000000",
					}, dk.DeviceKeys)
				})

				t.Run("UpdateKeys", func(t *testing.T) {
					assert := require.New(t)

					updateReq := pb.UpdateDeviceKeysRequest{
						DeviceKeys: &pb.DeviceKeys{
							DevEui: "0807060504030201",
							NwkKey: "08070605040302010102030405060708",
						},
					}

					_, err := api.UpdateKeys(context.Background(), &updateReq)
					assert.NoError(err)

					dk, err := api.GetKeys(context.Background(), &pb.GetDeviceKeysRequest{
						DevEui: "0807060504030201",
					})
					assert.NoError(err)

					assert.Equal(&pb.DeviceKeys{
						DevEui: "0807060504030201",
						NwkKey: "08070605040302010102030405060708",
						AppKey: "00000000000000000000000000000000",
					}, dk.DeviceKeys)
				})

				t.Run("DeleteKeys", func(t *testing.T) {
					assert := require.New(t)

					_, err := api.DeleteKeys(context.Background(), &pb.DeleteDeviceKeysRequest{
						DevEui: "0807060504030201",
					})
					assert.NoError(err)

					_, err = api.GetKeys(context.Background(), &pb.GetDeviceKeysRequest{
						DevEui: "0807060504030201",
					})
					assert.Equal(codes.NotFound, grpc.Code(err))
				})
			})

			t.Run("Deactivate", func(t *testing.T) {
				assert := require.New(t)

				deactivateReq := pb.DeactivateDeviceRequest{
					DevEui: "0807060504030201",
				}

				_, err := api.Deactivate(context.Background(), &deactivateReq)
				assert.NoError(err)

				// test that the device was de-activated in the NS
				assert.Equal(ns.DeactivateDeviceRequest{
					DevEui: []byte{8, 7, 6, 5, 4, 3, 2, 1},
				}, <-nsClient.DeactivateDeviceChan)
			})

			t.Run("ABP activate", func(t *testing.T) {
				assert := require.New(t)

				nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
					DeviceProfile: &ns.DeviceProfile{
						SupportsJoin: false,
					},
				}

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

				_, err := api.Activate(context.Background(), &activateReq)
				assert.NoError(err)

				// the device was first de-activated
				assert.Equal(ns.DeactivateDeviceRequest{
					DevEui: []byte{8, 7, 6, 5, 4, 3, 2, 1},
				}, <-nsClient.DeactivateDeviceChan)

				// device was activated
				assert.Equal(ns.ActivateDeviceRequest{
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
				}, <-nsClient.ActivateDeviceChan)

				// activation was stored
				d, err := storage.GetDevice(context.Background(), storage.DB(), lorawan.EUI64{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}, false, true)
				assert.NoError(err)
				assert.Equal(lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}, d.AppSKey)
				assert.Equal(lorawan.DevAddr{1, 2, 3, 4}, d.DevAddr)
			})

			t.Run("StreamEventLogs", func(t *testing.T) {
				assert := require.New(t)

				respChan := make(chan *pb.StreamDeviceEventLogsResponse)

				client, err := api.StreamEventLogs(context.Background(), &pb.StreamDeviceEventLogsRequest{
					DevEui: "0807060504030201",
				})
				assert.NoError(err)

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

				assert.NoError(eventlog.LogEventForDevice(lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, eventlog.Join, &integration.JoinEvent{}))

				resp := <-respChan
				assert.Equal(eventlog.Join, resp.Type)
			})

			t.Run("GetStats", func(t *testing.T) {
				assert := require.New(t)

				metrics := storage.MetricsRecord{
					Time: time.Now(),
					Metrics: map[string]float64{
						"rx_count":          2,
						"gw_rssi_sum":       -120.0,
						"gw_snr_sum":        10.0,
						"rx_freq_868100000": 2,
						"rx_dr_2":           2,
						"error_TOO_LATE":    1,
					},
				}
				assert.NoError(storage.SaveMetrics(context.Background(), "device:0807060504030201", metrics))

				resp, err := api.GetStats(context.Background(), &pb.GetDeviceStatsRequest{
					DevEui:         "0807060504030201",
					Interval:       "MINUTE",
					StartTimestamp: ptypes.TimestampNow(),
					EndTimestamp:   ptypes.TimestampNow(),
				})
				assert.NoError(err)
				assert.Len(resp.Result, 1)
				resp.Result[0].Timestamp = nil
				assert.Equal(&pb.DeviceStats{
					RxPackets: 2,
					GwRssi:    -60,
					GwSnr:     5,
					RxPacketsPerFrequency: map[uint32]uint32{
						868100000: 2,
					},
					RxPacketsPerDr: map[uint32]uint32{
						2: 2,
					},
					Errors: map[string]uint32{
						"TOO_LATE": 1,
					},
				}, resp.Result[0])
			})

			t.Run("Delete", func(t *testing.T) {
				assert := require.New(t)

				_, err := api.Delete(context.Background(), &pb.DeleteDeviceRequest{
					DevEui: "0807060504030201",
				})
				assert.NoError(err)

				_, err = api.Get(context.Background(), &pb.GetDeviceRequest{
					DevEui: "0807060504030201",
				})
				assert.Equal(codes.NotFound, grpc.Code(err))
			})
		})
	})
}
