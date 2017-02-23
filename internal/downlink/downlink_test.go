package downlink

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

func TestHandleDownlinkQueueItem(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database, an application + node", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		nsClient := test.NewNetworkServerClient()
		nsClient.GetNodeSessionResponse = ns.GetNodeSessionResponse{
			FCntDown: 12,
		}

		//ctx := context.Background()
		lsCtx := common.Context{
			DB:            db,
			NetworkServer: nsClient,
		}

		app := storage.Application{
			Name: "test-app",
		}
		So(storage.CreateApplication(db, &app), ShouldBeNil)
		node := storage.Node{
			ApplicationID: app.ID,
			Name:          "test-node",
			DevEUI:        [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevAddr:       [4]byte{1, 2, 3, 4},
			AppSKey:       [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		}
		So(storage.CreateNode(db, node), ShouldBeNil)

		qi := storage.DownlinkQueueItem{
			Reference: "test",
			DevEUI:    node.DevEUI,
			Confirmed: false,
			FPort:     10,
			Data:      []byte{1, 2, 3, 4},
		}

		b, err := lorawan.EncryptFRMPayload(node.AppSKey, false, node.DevAddr, 12, []byte{1, 2, 3, 4})
		So(err, ShouldBeNil)

		Convey("When calling HandleDownlinkQueueItem for a non class-c device", func() {
			So(HandleDownlinkQueueItem(lsCtx, node, &qi), ShouldBeNil)

			Convey("Then the item was added to the queue", func() {
				items, err := storage.GetDownlinkQueueItems(db, node.DevEUI)
				So(err, ShouldBeNil)
				So(items, ShouldHaveLength, 1)
				So(items[0], ShouldResemble, qi)
			})

			Convey("Then nothing was sent to the network-server", func() {
				So(nsClient.PushDataDownChan, ShouldHaveLength, 0)
			})
		})

		Convey("When calling HandleDownlinkQueueItem for a class-c device", func() {
			node.IsClassC = true
			So(storage.UpdateNode(db, node), ShouldBeNil)

			Convey("When the queue item is confirmed", func() {
				qi.Confirmed = true

				So(HandleDownlinkQueueItem(lsCtx, node, &qi), ShouldBeNil)

				Convey("Then the payload was sent to the network-server", func() {
					So(nsClient.PushDataDownChan, ShouldHaveLength, 1)
					So(<-nsClient.PushDataDownChan, ShouldResemble, ns.PushDataDownRequest{
						DevEUI:    node.DevEUI[:],
						Data:      b,
						Confirmed: true,
						FPort:     10,
						FCnt:      12,
					})
				})

				Convey("Then the item was added as pending to the queue", func() {
					items, err := storage.GetDownlinkQueueItems(db, node.DevEUI)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 1)
					So(items[0].Pending, ShouldBeTrue)
				})
			})

			Convey("When the queue item is unconfirmed", func() {
				So(HandleDownlinkQueueItem(lsCtx, node, &qi), ShouldBeNil)

				Convey("Then the payload was sent to the network-server", func() {
					So(nsClient.PushDataDownChan, ShouldHaveLength, 1)
					So(<-nsClient.PushDataDownChan, ShouldResemble, ns.PushDataDownRequest{
						DevEUI:    node.DevEUI[:],
						Data:      b,
						Confirmed: false,
						FPort:     10,
						FCnt:      12,
					})
				})

				Convey("Then the queue is empty", func() {
					items, err := storage.GetDownlinkQueueItems(db, node.DevEUI)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 0)
				})
			})
		})

	})
}
