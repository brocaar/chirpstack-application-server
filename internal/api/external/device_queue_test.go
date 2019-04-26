package external

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/codec"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

func (ts *APITestSuite) TestDownlinkQueue() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	nsClient.GetNextDownlinkFCntForDevEUIResponse = ns.GetNextDownlinkFCntForDevEUIResponse{
		FCnt: 12,
	}
	networkserver.SetPool(mock.NewPool(nsClient))

	validator := &TestValidator{}
	api := NewDeviceQueueAPI(validator)

	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(storage.DB(), &org))

	n := storage.NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:1234",
	}
	assert.NoError(storage.CreateNetworkServer(storage.DB(), &n))

	sp := storage.ServiceProfile{
		Name:            "test-sp",
		NetworkServerID: n.ID,
		OrganizationID:  org.ID,
	}
	assert.NoError(storage.CreateServiceProfile(storage.DB(), &sp))
	spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
	assert.NoError(err)

	dp := storage.DeviceProfile{
		Name:            "test-dp",
		NetworkServerID: n.ID,
		OrganizationID:  org.ID,
	}
	assert.NoError(storage.CreateDeviceProfile(storage.DB(), &dp))
	dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
	assert.NoError(err)

	app := storage.Application{
		OrganizationID:   org.ID,
		ServiceProfileID: spID,
		Name:             "test-app",
	}
	assert.NoError(storage.CreateApplication(storage.DB(), &app))

	d := storage.Device{
		ApplicationID:   app.ID,
		DeviceProfileID: dpID,
		Name:            "test-node",
		DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(storage.CreateDevice(storage.DB(), &d))

	da := storage.DeviceActivation{
		DevEUI:  d.DevEUI,
		DevAddr: lorawan.DevAddr{1, 2, 3, 4},
		AppSKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(storage.CreateDeviceActivation(storage.DB(), &da))

	b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, 12, []byte{1, 2, 3, 4})
	assert.NoError(err)

	ts.T().Run("codec configured on application", func(t *testing.T) {
		assert := require.New(t)

		app.PayloadCodec = codec.CustomJSType
		app.PayloadEncoderScript = `
				function Encode(fPort, obj) {
					return [
						obj.Bytes[3],
						obj.Bytes[2],
						obj.Bytes[1],
						obj.Bytes[0]
					];
				}
			`
		assert.NoError(storage.UpdateApplication(storage.DB(), app))

		t.Run("Enqueue with raw JSON", func(t *testing.T) {
			assert := require.New(t)

			resp, err := api.Enqueue(context.Background(), &pb.EnqueueDeviceQueueItemRequest{
				DeviceQueueItem: &pb.DeviceQueueItem{
					DevEui:     d.DevEUI.String(),
					FPort:      10,
					JsonObject: `{"Bytes": [4,3,2,1]}`,
				},
			})
			assert.NoError(err)
			assert.Equal(&pb.EnqueueDeviceQueueItemResponse{
				FCnt: 12,
			}, resp)

			assert.Equal(ns.CreateDeviceQueueItemRequest{
				Item: &ns.DeviceQueueItem{
					DevEui:     d.DevEUI[:],
					FrmPayload: b,
					FCnt:       12,
					FPort:      10,
				},
			}, <-nsClient.CreateDeviceQueueItemChan)
		})
	})

	ts.T().Run("codec configured on device-profile", func(t *testing.T) {
		assert := require.New(t)

		dp.PayloadCodec = codec.CustomJSType
		dp.PayloadEncoderScript = `
				function Encode(fPort, obj) {
					return [
						obj.Bytes[0],
						obj.Bytes[1],
						obj.Bytes[2],
						obj.Bytes[3]
					];
				}
			`
		assert.NoError(storage.UpdateDeviceProfile(storage.DB(), &dp))

		t.Run("Enqueue with raw JSON", func(t *testing.T) {
			assert := require.New(t)

			resp, err := api.Enqueue(context.Background(), &pb.EnqueueDeviceQueueItemRequest{
				DeviceQueueItem: &pb.DeviceQueueItem{
					DevEui:     d.DevEUI.String(),
					FPort:      10,
					JsonObject: `{"Bytes": [4,3,2,1]}`,
				},
			})
			assert.NoError(err)
			assert.Equal(&pb.EnqueueDeviceQueueItemResponse{
				FCnt: 12,
			}, resp)

			assert.Equal(ns.CreateDeviceQueueItemRequest{
				Item: &ns.DeviceQueueItem{
					DevEui:     d.DevEUI[:],
					FrmPayload: []byte{0xa3, 0x9c, 0x42, 0xca},
					FCnt:       12,
					FPort:      10,
				},
			}, <-nsClient.CreateDeviceQueueItemChan)
		})
	})

	ts.T().Run("List with mocked device-queue item", func(t *testing.T) {
		assert := require.New(t)

		nsClient.GetDeviceQueueItemsForDevEUIResponse = ns.GetDeviceQueueItemsForDevEUIResponse{
			Items: []*ns.DeviceQueueItem{
				{
					DevEui:     d.DevEUI[:],
					FrmPayload: b,
					FCnt:       12,
					FPort:      10,
					Confirmed:  true,
				},
			},
		}

		resp, err := api.List(context.Background(), &pb.ListDeviceQueueItemsRequest{
			DevEui: d.DevEUI.String(),
		})
		assert.NoError(err)
		assert.Len(resp.DeviceQueueItems, 1)
		assert.Equal(&pb.DeviceQueueItem{
			DevEui:    d.DevEUI.String(),
			Confirmed: true,
			FPort:     10,
			FCnt:      12,
			Data:      []byte{1, 2, 3, 4},
		}, resp.DeviceQueueItems[0])
	})

	ts.T().Run("Flush", func(t *testing.T) {
		assert := require.New(t)

		_, err := api.Flush(context.Background(), &pb.FlushDeviceQueueRequest{
			DevEui: d.DevEUI.String(),
		})
		assert.NoError(err)
		assert.Equal(ns.FlushDeviceQueueForDevEUIRequest{
			DevEui: d.DevEUI[:],
		}, <-nsClient.FlushDeviceQueueForDevEUIChan)
	})
}
