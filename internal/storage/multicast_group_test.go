package storage

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lorawan"
)

func TestMulticastGroupValidate(t *testing.T) {
	tests := []struct {
		MulticastGroup MulticastGroup
		Error          error
	}{
		{
			MulticastGroup: MulticastGroup{
				Name: "valid-name",
			},
		},
		{
			MulticastGroup: MulticastGroup{
				Name: "",
			},
			Error: ErrMulticastGroupInvalidName,
		},
		{
			MulticastGroup: MulticastGroup{
				Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		},
		{
			MulticastGroup: MulticastGroup{
				Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			Error: ErrMulticastGroupInvalidName,
		},
	}

	assert := require.New(t)

	for _, tst := range tests {
		assert.Equal(tst.Error, tst.MulticastGroup.Validate())
	}
}

func (ts *StorageTestSuite) TestMulticastGroup() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	n := NetworkServer{
		Name:   "test",
		Server: "test:1234",
	}
	assert.NoError(CreateNetworkServer(context.Background(), ts.Tx(), &n))

	org := Organization{
		Name: "test-org-123",
	}
	assert.NoError(CreateOrganization(context.Background(), ts.Tx(), &org))

	sp := ServiceProfile{
		Name:            "test-sp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(CreateServiceProfile(context.Background(), ts.Tx(), &sp))

	app := Application{
		Name:           "test-app",
		OrganizationID: org.ID,
	}
	copy(app.ServiceProfileID[:], sp.ServiceProfile.Id)
	assert.NoError(CreateApplication(context.Background(), ts.Tx(), &app))

	dp := DeviceProfile{
		Name:            "test-dp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(CreateDeviceProfile(context.Background(), ts.Tx(), &dp))
	var dpID uuid.UUID
	copy(dpID[:], dp.DeviceProfile.Id)

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		// create group
		mg := MulticastGroup{
			Name:          "test-mg",
			ApplicationID: app.ID,
			MCAppSKey:     lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
			MulticastGroup: ns.MulticastGroup{
				McAddr:           []byte{1, 2, 3, 4},
				McNwkSKey:        []byte{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1},
				GroupType:        ns.MulticastGroupType_CLASS_B,
				Dr:               3,
				Frequency:        868300000,
				PingSlotPeriod:   32,
				ServiceProfileId: sp.ServiceProfile.Id,
			},
		}
		assert.NoError(CreateMulticastGroup(context.Background(), ts.Tx(), &mg))
		mg.CreatedAt = mg.CreatedAt.Round(time.Second).UTC()
		mg.UpdatedAt = mg.UpdatedAt.Round(time.Second).UTC()

		// validate it has been created on ns
		createReq := <-nsClient.CreateMulticastGroupChan
		assert.Equal(mg.MulticastGroup, *createReq.MulticastGroup)

		nsClient.GetMulticastGroupResponse.MulticastGroup = createReq.MulticastGroup
		var mgID uuid.UUID
		copy(mgID[:], mg.MulticastGroup.Id)

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			mgGet, err := GetMulticastGroup(context.Background(), ts.Tx(), mgID, false, false)
			assert.NoError(err)

			mgGet.CreatedAt = mgGet.CreatedAt.Round(time.Second).UTC()
			mgGet.UpdatedAt = mgGet.UpdatedAt.Round(time.Second).UTC()
			assert.Equal(mg, mgGet)

			getReq := <-nsClient.GetMulticastGroupChan
			assert.Equal(ns.GetMulticastGroupRequest{
				Id: mgID[:],
			}, getReq)
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			mg.Name = "test-mg-updated"
			mg.MCAppSKey = lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
			mg.MulticastGroup = ns.MulticastGroup{
				Id:               mg.MulticastGroup.Id,
				McAddr:           []byte{4, 3, 2, 1},
				McNwkSKey:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
				GroupType:        ns.MulticastGroupType_CLASS_C,
				Dr:               2,
				Frequency:        868500000,
				PingSlotPeriod:   64,
				ServiceProfileId: sp.ServiceProfile.Id,
			}

			assert.NoError(UpdateMulticastGroup(context.Background(), ts.Tx(), &mg))
			mg.UpdatedAt = mg.UpdatedAt.Round(time.Second).UTC()

			updateReq := <-nsClient.UpdateMulticastGroupChan
			assert.Equal(mg.MulticastGroup, *updateReq.MulticastGroup)
			nsClient.GetMulticastGroupResponse.MulticastGroup = updateReq.MulticastGroup

			mgGet, err := GetMulticastGroup(context.Background(), ts.Tx(), mgID, false, false)
			assert.NoError(err)

			mgGet.CreatedAt = mgGet.CreatedAt.Round(time.Second).UTC()
			mgGet.UpdatedAt = mgGet.UpdatedAt.Round(time.Second).UTC()
			assert.Equal(mg, mgGet)
		})

		t.Run("Add device", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetDeviceCountForMulticastGroup(context.Background(), ts.Tx(), mgID)
			assert.NoError(err)
			assert.Equal(0, count)

			d := Device{
				DevEUI:          lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
				Name:            "device-1",
				DeviceProfileID: dpID,
				ApplicationID:   app.ID,
			}
			assert.NoError(CreateDevice(context.Background(), ts.Tx(), &d))
			assert.NoError(AddDeviceToMulticastGroup(context.Background(), ts.Tx(), mgID, d.DevEUI))

			t.Run("List devices", func(t *testing.T) {
				assert := require.New(t)

				count, err := GetDeviceCountForMulticastGroup(context.Background(), ts.Tx(), mgID)
				assert.NoError(err)
				assert.Equal(1, count)

				devices, err := GetDevicesForMulticastGroup(context.Background(), ts.Tx(), mgID, 10, 0)
				assert.NoError(err)
				assert.Len(devices, 1)
			})

			t.Run("Remove device", func(t *testing.T) {
				assert := require.New(t)
				assert.NoError(RemoveDeviceFromMulticastGroup(context.Background(), ts.Tx(), mgID, d.DevEUI))
				count, err := GetDeviceCountForMulticastGroup(context.Background(), ts.Tx(), mgID)
				assert.NoError(err)
				assert.Equal(0, count)
			})
		})

		t.Run("List", func(t *testing.T) {
			assert := require.New(t)

			// having two devices will result in duplicated records and tests
			// that the sql 'distinct' is used correctly
			devices := []Device{
				{
					DevEUI:          lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 2},
					Name:            "device-2",
					DeviceProfileID: dpID,
					ApplicationID:   app.ID,
				},
				{
					DevEUI:          lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 3},
					Name:            "device-3",
					DeviceProfileID: dpID,
					ApplicationID:   app.ID,
				},
			}

			for i := range devices {
				assert.NoError(CreateDevice(context.Background(), ts.Tx(), &devices[i]))
				assert.NoError(AddDeviceToMulticastGroup(context.Background(), ts.Tx(), mgID, devices[i].DevEUI))
			}

			testTable := []struct {
				Name     string
				Filters  MulticastGroupFilters
				Expected []MulticastGroupListItem
			}{
				{
					Name: "no filters",
					Filters: MulticastGroupFilters{
						Limit: 10,
					},
					Expected: []MulticastGroupListItem{
						{
							ID:              mgID,
							Name:            "test-mg-updated",
							ApplicationID:   app.ID,
							ApplicationName: app.Name,
						},
					},
				},
				{
					Name: "org filter",
					Filters: MulticastGroupFilters{
						OrganizationID: org.ID,
						Limit:          10,
					},
					Expected: []MulticastGroupListItem{
						{
							ID:              mgID,
							Name:            "test-mg-updated",
							ApplicationID:   app.ID,
							ApplicationName: app.Name,
						},
					},
				},
				{
					Name: "non-matching org filter",
					Filters: MulticastGroupFilters{
						OrganizationID: org.ID + 1,
					},
				},
				{
					Name: "application filter",
					Filters: MulticastGroupFilters{
						ApplicationID: app.ID,
						Limit:         10,
					},
					Expected: []MulticastGroupListItem{
						{
							ID:              mgID,
							Name:            "test-mg-updated",
							ApplicationID:   app.ID,
							ApplicationName: app.Name,
						},
					},
				},
				{
					Name: "non-matching application filter",
					Filters: MulticastGroupFilters{
						ApplicationID: app.ID + 1,
						Limit:         10,
					},
				},
				{
					Name: "device eui filter",
					Filters: MulticastGroupFilters{
						DevEUI: devices[0].DevEUI,
						Limit:  10,
					},
					Expected: []MulticastGroupListItem{
						{
							ID:              mgID,
							Name:            "test-mg-updated",
							ApplicationID:   app.ID,
							ApplicationName: app.Name,
						},
					},
				},
				{
					Name: "non-matching device eui filter",
					Filters: MulticastGroupFilters{
						DevEUI: lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
						Limit:  10,
					},
				},
				{
					Name: "search filter",
					Filters: MulticastGroupFilters{
						Search: "upda",
						Limit:  10,
					},
					Expected: []MulticastGroupListItem{
						{
							ID:              mgID,
							Name:            "test-mg-updated",
							ApplicationID:   app.ID,
							ApplicationName: app.Name,
						},
					},
				},
				{
					Name: "non-matching search filter",
					Filters: MulticastGroupFilters{
						Search: "foo",
						Limit:  10,
					},
				},
			}

			for _, test := range testTable {
				t.Run(test.Name, func(t *testing.T) {
					assert := require.New(t)

					count, err := GetMulticastGroupCount(context.Background(), ts.Tx(), test.Filters)
					assert.NoError(err)
					assert.Equal(len(test.Expected), count)

					items, err := GetMulticastGroups(context.Background(), ts.Tx(), test.Filters)
					assert.NoError(err)
					for i, item := range items {
						assert.False(item.CreatedAt.IsZero())
						assert.False(item.UpdatedAt.IsZero())
						items[i].CreatedAt = time.Time{}
						items[i].UpdatedAt = time.Time{}
					}
					assert.Equal(test.Expected, items)
				})
			}
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteMulticastGroup(context.Background(), ts.Tx(), mgID))

			delReq := <-nsClient.DeleteMulticastGroupChan
			assert.Equal(ns.DeleteMulticastGroupRequest{
				Id: mgID[:],
			}, delReq)

			assert.Error(DeleteMulticastGroup(context.Background(), ts.Tx(), mgID))
		})
	})
}
