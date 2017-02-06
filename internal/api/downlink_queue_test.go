package api

import (
	"context"
	"testing"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDownlinkQueueAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database, an application + node and api instance", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		ctx := context.Background()
		lsCtx := common.Context{DB: db}
		validator := &TestValidator{}
		api := NewDownlinkQueueAPI(lsCtx, validator)

		app := storage.Application{
			Name: "test-app",
		}
		So(storage.CreateApplication(db, &app), ShouldBeNil)
		node := storage.Node{
			ApplicationID: app.ID,
			DevEUI:        [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(storage.CreateNode(db, node), ShouldBeNil)

		Convey("When enqueueing a downlink queue item", func() {
			_, err := api.Enqueue(ctx, &pb.EnqueueDownlinkQueueItemRequest{
				ApplicationName: "test-app",
				DevEUI:          "0102030405060708",
				Confirmed:       true,
				FPort:           10,
				Data:            []byte{1, 2, 3, 4},
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 3)

			Convey("Then the queue contains a single item", func() {
				resp, err := api.List(ctx, &pb.ListDownlinkQueueItemsRequest{
					ApplicationName: "test-app",
					DevEUI:          "0102030405060708",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)

				So(resp.Items, ShouldHaveLength, 1)
				So(resp.Items[0], ShouldResemble, &pb.DownlinkQueueItem{
					Id:        1,
					DevEUI:    "0102030405060708",
					Confirmed: true,
					Pending:   false,
					FPort:     10,
					Data:      []byte{1, 2, 3, 4},
				})
			})

			Convey("When removing the queue item", func() {
				_, err := api.Delete(ctx, &pb.DeleteDownlinkQeueueItemRequest{
					ApplicationName: "test-app",
					DevEUI:          "0102030405060708",
					Id:              1,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)

				Convey("Then the downlink queue item has been deleted", func() {
					resp, err := api.List(ctx, &pb.ListDownlinkQueueItemsRequest{
						ApplicationName: "test-app",
						DevEUI:          "0102030405060708",
					})
					So(err, ShouldBeNil)
					So(resp.Items, ShouldHaveLength, 0)
				})
			})
		})
	})
}
