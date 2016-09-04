package api

import (
	"testing"

	"golang.org/x/net/context"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNodeAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database with an application and api instance", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		ctx := context.Background()
		lsCtx := common.Context{DB: db}
		validator := &TestValidator{}
		api := NewNodeAPI(lsCtx, validator)

		Convey("When creating a node", func() {
			_, err := api.Create(ctx, &pb.CreateNodeRequest{
				DevEUI:      "0807060504030201",
				AppEUI:      "0102030405060708",
				AppKey:      "01020304050607080102030405060708",
				RxDelay:     1,
				Rx1DROffset: 3,
				RxWindow:    pb.RXWindow_RX2,
				Rx2DR:       3,
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 3)

			Convey("The node has been created", func() {
				node, err := api.Get(ctx, &pb.GetNodeRequest{DevEUI: "0807060504030201"})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)
				So(node, ShouldResemble, &pb.GetNodeResponse{
					DevEUI:      "0807060504030201",
					AppEUI:      "0102030405060708",
					AppKey:      "01020304050607080102030405060708",
					RxDelay:     1,
					Rx1DROffset: 3,
					RxWindow:    pb.RXWindow_RX2,
					Rx2DR:       3,
				})
			})

			Convey("Then listing the nodes returns a single items", func() {
				nodes, err := api.List(ctx, &pb.ListNodeRequest{
					Limit: 10,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				So(nodes.Result, ShouldHaveLength, 1)
				So(nodes.TotalCount, ShouldEqual, 1)
				So(nodes.Result[0], ShouldResemble, &pb.GetNodeResponse{
					DevEUI:      "0807060504030201",
					AppEUI:      "0102030405060708",
					AppKey:      "01020304050607080102030405060708",
					RxDelay:     1,
					Rx1DROffset: 3,
					RxWindow:    pb.RXWindow_RX2,
					Rx2DR:       3,
				})
			})

			Convey("When updating the node", func() {
				_, err := api.Update(ctx, &pb.UpdateNodeRequest{
					DevEUI:      "0807060504030201",
					AppEUI:      "0102030405060708",
					AppKey:      "08070605040302010807060504030201",
					RxDelay:     3,
					Rx1DROffset: 1,
					RxWindow:    pb.RXWindow_RX2,
					Rx2DR:       4,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)

				Convey("Then the node has been updated", func() {
					node, err := api.Get(ctx, &pb.GetNodeRequest{DevEUI: "0807060504030201"})
					So(err, ShouldBeNil)
					So(node, ShouldResemble, &pb.GetNodeResponse{
						DevEUI:      "0807060504030201",
						AppEUI:      "0102030405060708",
						AppKey:      "08070605040302010807060504030201",
						RxDelay:     3,
						Rx1DROffset: 1,
						RxWindow:    pb.RXWindow_RX2,
						Rx2DR:       4,
					})
				})
			})

			Convey("After deleting the node", func() {
				_, err := api.Delete(ctx, &pb.DeleteNodeRequest{DevEUI: "0807060504030201"})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)

				Convey("Then listing the nodes returns zero nodes", func() {
					nodes, err := api.List(ctx, &pb.ListNodeRequest{Limit: 10})
					So(err, ShouldBeNil)
					So(nodes.TotalCount, ShouldEqual, 0)
					So(nodes.Result, ShouldHaveLength, 0)
				})
			})
		})
	})
}
