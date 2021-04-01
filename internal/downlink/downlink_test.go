package downlink

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/lib/pq/hstore"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
)

func TestHandleDownlinkQueueItem(t *testing.T) {
	conf := test.GetConfig()
	if err := storage.Setup(conf); err != nil {
		t.Fatal(err)
	}

	Convey("Given a clean database an organization, application + node", t, func() {
		So(storage.MigrateDown(storage.DB().DB), ShouldBeNil)
		So(storage.MigrateUp(storage.DB().DB), ShouldBeNil)

		nsClient := mock.NewClient()
		nsClient.GetNextDownlinkFCntForDevEUIResponse = ns.GetNextDownlinkFCntForDevEUIResponse{
			FCnt: 12,
		}
		networkserver.SetPool(mock.NewPool(nsClient))

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(context.Background(), storage.DB(), &org), ShouldBeNil)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(context.Background(), storage.DB(), &n), ShouldBeNil)

		sp := storage.ServiceProfile{
			Name:            "test-sp",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
		}
		So(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp), ShouldBeNil)
		spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
		So(err, ShouldBeNil)

		dp := storage.DeviceProfile{
			Name:            "test-dp",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
		}
		So(storage.CreateDeviceProfile(context.Background(), storage.DB(), &dp), ShouldBeNil)
		dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
		So(err, ShouldBeNil)

		app := storage.Application{
			OrganizationID:   org.ID,
			Name:             "test-app",
			ServiceProfileID: spID,
		}
		So(storage.CreateApplication(context.Background(), storage.DB(), &app), ShouldBeNil)

		device := storage.Device{
			ApplicationID:   app.ID,
			DeviceProfileID: dpID,
			Name:            "test-node",
			DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Tags: hstore.Hstore{
				Map: map[string]sql.NullString{
					"foo": sql.NullString{String: "bar", Valid: true},
				},
			},
			Variables: hstore.Hstore{
				Map: map[string]sql.NullString{
					"secret_token": sql.NullString{String: "secret value", Valid: true},
				},
			},
			DevAddr: lorawan.DevAddr{0x01, 0x02, 0x03, 0x04},
			AppSKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		}
		So(storage.CreateDevice(context.Background(), storage.DB(), &device), ShouldBeNil)

		b, err := lorawan.EncryptFRMPayload(device.AppSKey, false, device.DevAddr, 12, []byte{1, 2, 3, 4})
		So(err, ShouldBeNil)

		Convey("Given a set of tests", func() {
			tests := []struct {
				Name                 string
				Payload              models.DataDownPayload
				PayloadCodec         codec.Type
				PayloadEncoderScript string

				ExpectedError                        error
				ExpectedCreateDeviceQueueItemRequest ns.CreateDeviceQueueItemRequest
			}{
				{
					Name: "unconfirmed payload",
					Payload: models.DataDownPayload{
						ApplicationID: app.ID,
						DevEUI:        device.DevEUI,
						Confirmed:     false,
						FPort:         2,
						Data:          []byte{1, 2, 3, 4},
					},

					ExpectedCreateDeviceQueueItemRequest: ns.CreateDeviceQueueItemRequest{
						Item: &ns.DeviceQueueItem{
							DevAddr:    device.DevAddr[:],
							DevEui:     device.DevEUI[:],
							FrmPayload: b,
							FCnt:       12,
							FPort:      2,
							Confirmed:  false,
						},
					},
				},
				{
					Name: "confirmed payload",
					Payload: models.DataDownPayload{
						ApplicationID: app.ID,
						DevEUI:        device.DevEUI,
						Confirmed:     true,
						FPort:         2,
						Data:          []byte{1, 2, 3, 4},
					},

					ExpectedCreateDeviceQueueItemRequest: ns.CreateDeviceQueueItemRequest{
						Item: &ns.DeviceQueueItem{
							DevAddr:    device.DevAddr[:],
							DevEui:     device.DevEUI[:],
							FrmPayload: b,
							FCnt:       12,
							FPort:      2,
							Confirmed:  true,
						},
					},
				},
				{
					Name: "invalid application id",
					Payload: models.DataDownPayload{
						ApplicationID: app.ID + 1,
						DevEUI:        device.DevEUI,
						Confirmed:     true,
						FPort:         2,
						Data:          []byte{1, 2, 3, 4},
					},
					ExpectedError: errors.New("enqueue downlink payload: device does not exist for given application"),
				},
				{
					Name:         "custom payload encoder",
					PayloadCodec: codec.CustomJSType,
					PayloadEncoderScript: `
						function Encode(fPort, obj) {
							return [
								obj.Bytes[3],
								obj.Bytes[2],
								obj.Bytes[1],
								obj.Bytes[0]
							];
						}
					`,
					Payload: models.DataDownPayload{
						ApplicationID: app.ID,
						DevEUI:        device.DevEUI,
						FPort:         2,
						Object:        json.RawMessage(`{"Bytes": [4, 3, 2, 1]}`),
					},

					ExpectedCreateDeviceQueueItemRequest: ns.CreateDeviceQueueItemRequest{
						Item: &ns.DeviceQueueItem{
							DevAddr:    device.DevAddr[:],
							DevEui:     device.DevEUI[:],
							FrmPayload: b,
							FCnt:       12,
							FPort:      2,
							Confirmed:  false,
						},
					},
				},
				{
					Name:         "custom payload encoder - payload is null",
					PayloadCodec: codec.CustomJSType,
					PayloadEncoderScript: `
						function Encode(fPort, obj) {
							return [
								obj.Bytes[3],
								obj.Bytes[2],
								obj.Bytes[1],
								obj.Bytes[0]
							];
						}
					`,
					Payload: models.DataDownPayload{
						ApplicationID: app.ID,
						DevEUI:        device.DevEUI,
						FPort:         2,
						Object:        json.RawMessage(`null`),
					},

					ExpectedCreateDeviceQueueItemRequest: ns.CreateDeviceQueueItemRequest{
						Item: &ns.DeviceQueueItem{
							DevAddr:   device.DevAddr[:],
							DevEui:    device.DevEUI[:],
							FCnt:      12,
							FPort:     2,
							Confirmed: false,
						},
					},
				},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
					// update device-profile
					dp.PayloadCodec = test.PayloadCodec
					dp.PayloadEncoderScript = test.PayloadEncoderScript
					So(storage.UpdateDeviceProfile(context.Background(), storage.DB(), &dp), ShouldBeNil)

					err := handleDataDownPayload(context.Background(), test.Payload)
					if test.ExpectedError != nil {
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, test.ExpectedError.Error())
						return
					}

					So(err, ShouldEqual, nil)
					So(nsClient.GetNextDownlinkFCntForDevEUIChan, ShouldHaveLength, 1)
					So(nsClient.CreateDeviceQueueItemChan, ShouldHaveLength, 1)
					So(<-nsClient.CreateDeviceQueueItemChan, ShouldResemble, test.ExpectedCreateDeviceQueueItemRequest)
				})
			}
		})
	})
}
