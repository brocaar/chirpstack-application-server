package downlink

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gusseleet/lora-app-server/internal/codec"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/handler"
	"github.com/pkg/errors"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/gusseleet/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

func TestHandleDownlinkQueueItem(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	config.C.PostgreSQL.DB = db

	Convey("Given a clean database an organization, application + node", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		nsClient := test.NewNetworkServerClient()
		nsClient.GetNextDownlinkFCntForDevEUIResponse = ns.GetNextDownlinkFCntForDevEUIResponse{
			FCnt: 12,
		}
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

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

		dp := storage.DeviceProfile{
			Name:            "test-dp",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			DeviceProfile:   backend.DeviceProfile{},
		}
		So(storage.CreateDeviceProfile(config.C.PostgreSQL.DB, &dp), ShouldBeNil)

		app := storage.Application{
			OrganizationID:   org.ID,
			Name:             "test-app",
			ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
		}
		So(storage.CreateApplication(config.C.PostgreSQL.DB, &app), ShouldBeNil)

		device := storage.Device{
			ApplicationID:   app.ID,
			DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
			Name:            "test-node",
			DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(storage.CreateDevice(config.C.PostgreSQL.DB, &device), ShouldBeNil)

		da := storage.DeviceActivation{
			DevEUI:  [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevAddr: [4]byte{1, 2, 3, 4},
			AppSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		}
		So(storage.CreateDeviceActivation(config.C.PostgreSQL.DB, &da), ShouldBeNil)

		b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, 12, []byte{1, 2, 3, 4})
		So(err, ShouldBeNil)

		Convey("Given a set of tests", func() {
			tests := []struct {
				Name                 string
				Payload              handler.DataDownPayload
				PayloadCodec         codec.Type
				PayloadEncoderScript string

				ExpectedDeviceQueueMapping           bool
				ExpectedError                        error
				ExpectedCreateDeviceQueueItemRequest ns.CreateDeviceQueueItemRequest
			}{
				{
					Name: "unconfirmed payload",
					Payload: handler.DataDownPayload{
						ApplicationID: app.ID,
						DevEUI:        device.DevEUI,
						Reference:     "test-123",
						Confirmed:     false,
						FPort:         2,
						Data:          []byte{1, 2, 3, 4},
					},

					ExpectedCreateDeviceQueueItemRequest: ns.CreateDeviceQueueItemRequest{
						Item: &ns.DeviceQueueItem{
							DevEUI:     device.DevEUI[:],
							FrmPayload: b,
							FCnt:       12,
							FPort:      2,
							Confirmed:  false,
						},
					},
				},
				{
					Name: "confirmed payload",
					Payload: handler.DataDownPayload{
						ApplicationID: app.ID,
						DevEUI:        device.DevEUI,
						Reference:     "test-123",
						Confirmed:     true,
						FPort:         2,
						Data:          []byte{1, 2, 3, 4},
					},

					ExpectedCreateDeviceQueueItemRequest: ns.CreateDeviceQueueItemRequest{
						Item: &ns.DeviceQueueItem{
							DevEUI:     device.DevEUI[:],
							FrmPayload: b,
							FCnt:       12,
							FPort:      2,
							Confirmed:  true,
						},
					},
					ExpectedDeviceQueueMapping: true,
				},
				{
					Name: "invalid application id",
					Payload: handler.DataDownPayload{
						ApplicationID: app.ID + 1,
						DevEUI:        device.DevEUI,
						Reference:     "test-123",
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
					Payload: handler.DataDownPayload{
						ApplicationID: app.ID,
						DevEUI:        device.DevEUI,
						FPort:         2,
						Object:        json.RawMessage(`{"Bytes": [4, 3, 2, 1]}`),
					},

					ExpectedCreateDeviceQueueItemRequest: ns.CreateDeviceQueueItemRequest{
						Item: &ns.DeviceQueueItem{
							DevEUI:     device.DevEUI[:],
							FrmPayload: b,
							FCnt:       12,
							FPort:      2,
							Confirmed:  false,
						},
					},
				},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("Testint: %s [%d]", test.Name, i), func() {
					// update application
					app.PayloadCodec = test.PayloadCodec
					app.PayloadEncoderScript = test.PayloadEncoderScript
					So(storage.UpdateApplication(config.C.PostgreSQL.DB, app), ShouldBeNil)

					err := handleDataDownPayload(test.Payload)
					if test.ExpectedError != nil {
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, test.ExpectedError.Error())
						return
					}

					So(err, ShouldEqual, nil)
					So(nsClient.GetNextDownlinkFCntForDevEUIChan, ShouldHaveLength, 1)
					So(nsClient.CreateDeviceQueueItemChan, ShouldHaveLength, 1)
					So(<-nsClient.CreateDeviceQueueItemChan, ShouldResemble, test.ExpectedCreateDeviceQueueItemRequest)

					if test.ExpectedDeviceQueueMapping {
						dqm, err := storage.GetDeviceQueueMappingForDevEUIAndFCnt(config.C.PostgreSQL.DB, device.DevEUI, 12)
						So(err, ShouldBeNil)
						So(dqm.Reference, ShouldEqual, test.Payload.Reference)
					} else {
						_, err := storage.GetDeviceQueueMappingForDevEUIAndFCnt(config.C.PostgreSQL.DB, device.DevEUI, 12)
						So(err, ShouldEqual, storage.ErrDoesNotExist)
					}
				})
			}
		})
	})
}
