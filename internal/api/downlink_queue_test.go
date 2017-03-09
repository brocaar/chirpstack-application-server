package api

import (
	"context"
	"testing"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDownlinkQueueAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database, an application + node and api instance", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		nsClient := test.NewNetworkServerClient()
		ctx := context.Background()
		lsCtx := common.Context{
			DB:            db,
			NetworkServer: nsClient,
		}
		validator := &TestValidator{}
		api := NewDownlinkQueueAPI(lsCtx, validator)

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

		Convey("Given the node is a class-c device", func() {
			node.IsClassC = true
			So(storage.UpdateNode(db, node), ShouldBeNil)
			nsClient.GetNodeSessionResponse = ns.GetNodeSessionResponse{
				FCntDown: 12,
			}

			Convey("When enqueueing a unconfirmed queue item", func() {
				_, err := api.Enqueue(ctx, &pb.EnqueueDownlinkQueueItemRequest{
					DevEUI:    node.DevEUI.String(),
					Confirmed: false,
					FPort:     10,
					Data:      []byte{1, 2, 3, 4},
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the payload was directly sent to the network-server", func() {
					b, err := lorawan.EncryptFRMPayload(node.AppSKey, false, node.DevAddr, 12, []byte{1, 2, 3, 4})
					So(err, ShouldBeNil)

					So(nsClient.PushDataDownChan, ShouldHaveLength, 1)
					So(<-nsClient.PushDataDownChan, ShouldResemble, ns.PushDataDownRequest{
						DevEUI:    node.DevEUI[:],
						Data:      b,
						Confirmed: false,
						FPort:     10,
						FCnt:      12,
					})
				})

				Convey("Then the item was not added to the queue", func() {
					items, err := storage.GetDownlinkQueueItems(db, node.DevEUI)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 0)
				})
			})

			Convey("When enqueueing a confirmed queue item", func() {
				_, err := api.Enqueue(ctx, &pb.EnqueueDownlinkQueueItemRequest{
					DevEUI:    node.DevEUI.String(),
					Confirmed: true,
					FPort:     10,
					Data:      []byte{1, 2, 3, 4},
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the payload was directly sent to the network-server", func() {
					b, err := lorawan.EncryptFRMPayload(node.AppSKey, false, node.DevAddr, 12, []byte{1, 2, 3, 4})
					So(err, ShouldBeNil)

					So(nsClient.PushDataDownChan, ShouldHaveLength, 1)
					So(<-nsClient.PushDataDownChan, ShouldResemble, ns.PushDataDownRequest{
						DevEUI:    node.DevEUI[:],
						Data:      b,
						Confirmed: true,
						FPort:     10,
						FCnt:      12,
					})
				})

				Convey("Then the item was added as pending item to the queue", func() {
					items, err := storage.GetDownlinkQueueItems(db, node.DevEUI)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 1)
					So(items[0].Pending, ShouldBeTrue)
				})
			})
		})

		Convey("When enqueueing a downlink queue item", func() {
			_, err := api.Enqueue(ctx, &pb.EnqueueDownlinkQueueItemRequest{
				DevEUI:    node.DevEUI.String(),
				Confirmed: true,
				FPort:     10,
				Data:      []byte{1, 2, 3, 4},
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the queue contains a single item", func() {
				resp, err := api.List(ctx, &pb.ListDownlinkQueueItemsRequest{
					DevEUI: node.DevEUI.String(),
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				So(resp.Items, ShouldHaveLength, 1)
				So(resp.Items[0], ShouldResemble, &pb.DownlinkQueueItem{
					Id:        1,
					DevEUI:    node.DevEUI.String(),
					Confirmed: true,
					Pending:   false,
					FPort:     10,
					Data:      []byte{1, 2, 3, 4},
				})
			})

			Convey("Then nothing was sent to the network-server", func() {
				So(nsClient.PushDataDownChan, ShouldHaveLength, 0)
			})

			Convey("When removing the queue item", func() {
				_, err := api.Delete(ctx, &pb.DeleteDownlinkQeueueItemRequest{
					DevEUI: node.DevEUI.String(),
					Id:     1,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the downlink queue item has been deleted", func() {
					resp, err := api.List(ctx, &pb.ListDownlinkQueueItemsRequest{
						DevEUI: node.DevEUI.String(),
					})
					So(err, ShouldBeNil)
					So(resp.Items, ShouldHaveLength, 0)
				})
			})
		})
	})
}
