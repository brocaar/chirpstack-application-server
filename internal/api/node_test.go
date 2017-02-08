package api

import (
	"testing"

	"golang.org/x/net/context"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNodeAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database with an application and api instance", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		nsClient := test.NewNetworkServerClient()
		ctx := context.Background()
		lsCtx := common.Context{DB: db, NetworkServer: nsClient}
		validator := &TestValidator{}
		api := NewNodeAPI(lsCtx, validator)

		app := storage.Application{
			Name: "test-app",
		}
		So(storage.CreateApplication(db, &app), ShouldBeNil)

		Convey("When creating a node", func() {
			_, err := api.Create(ctx, &pb.CreateNodeRequest{
				ApplicationName:    "test-app",
				Name:               "test-node",
				Description:        "test node description",
				DevEUI:             "0807060504030201",
				AppEUI:             "0102030405060708",
				AppKey:             "01020304050607080102030405060708",
				RxDelay:            1,
				Rx1DROffset:        3,
				RxWindow:           pb.RXWindow_RX2,
				Rx2DR:              3,
				AdrInterval:        20,
				InstallationMargin: 5,
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 3)

			Convey("The node has been created", func() {
				node, err := api.Get(ctx, &pb.GetNodeRequest{
					ApplicationName: "test-app",
					NodeName:        "test-node",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)
				So(node, ShouldResemble, &pb.GetNodeResponse{
					Name:               "test-node",
					Description:        "test node description",
					DevEUI:             "0807060504030201",
					AppEUI:             "0102030405060708",
					AppKey:             "01020304050607080102030405060708",
					RxDelay:            1,
					Rx1DROffset:        3,
					RxWindow:           pb.RXWindow_RX2,
					Rx2DR:              3,
					AdrInterval:        20,
					InstallationMargin: 5,
				})
			})

			Convey("Then listing the nodes for the application returns a single items", func() {
				nodes, err := api.List(ctx, &pb.ListNodeRequest{
					ApplicationName: "test-app",
					Limit:           10,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 2)
				So(nodes.Result, ShouldHaveLength, 1)
				So(nodes.TotalCount, ShouldEqual, 1)
				So(nodes.Result[0], ShouldResemble, &pb.GetNodeResponse{
					Name:               "test-node",
					Description:        "test node description",
					DevEUI:             "0807060504030201",
					AppEUI:             "0102030405060708",
					AppKey:             "01020304050607080102030405060708",
					RxDelay:            1,
					Rx1DROffset:        3,
					RxWindow:           pb.RXWindow_RX2,
					Rx2DR:              3,
					AdrInterval:        20,
					InstallationMargin: 5,
				})
			})

			Convey("When updating the node", func() {
				_, err := api.Update(ctx, &pb.UpdateNodeRequest{
					ApplicationName:    "test-app",
					NodeName:           "test-node",
					Name:               "test-node-updated",
					Description:        "test node description updated",
					AppEUI:             "0102030405060708",
					AppKey:             "08070605040302010807060504030201",
					RxDelay:            3,
					Rx1DROffset:        1,
					RxWindow:           pb.RXWindow_RX2,
					Rx2DR:              4,
					AdrInterval:        30,
					InstallationMargin: 10,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)

				Convey("Then the node has been updated", func() {
					node, err := api.Get(ctx, &pb.GetNodeRequest{
						ApplicationName: "test-app",
						NodeName:        "test-node-updated",
					})
					So(err, ShouldBeNil)
					So(node, ShouldResemble, &pb.GetNodeResponse{
						Name:               "test-node-updated",
						Description:        "test node description updated",
						DevEUI:             "0807060504030201",
						AppEUI:             "0102030405060708",
						AppKey:             "08070605040302010807060504030201",
						RxDelay:            3,
						Rx1DROffset:        1,
						RxWindow:           pb.RXWindow_RX2,
						Rx2DR:              4,
						AdrInterval:        30,
						InstallationMargin: 10,
					})
				})
			})

			Convey("After deleting the node", func() {
				_, err := api.Delete(ctx, &pb.DeleteNodeRequest{
					ApplicationName: "test-app",
					NodeName:        "test-node",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)

				Convey("Then an attempt was made to delete the node-session", func() {
					So(nsClient.DeleteNodeSessionChan, ShouldHaveLength, 1)
					So(<-nsClient.DeleteNodeSessionChan, ShouldResemble, ns.DeleteNodeSessionRequest{
						DevEUI: []byte{8, 7, 6, 5, 4, 3, 2, 1},
					})
				})

				Convey("Then listing the nodes returns zero nodes", func() {
					nodes, err := api.List(ctx, &pb.ListNodeRequest{
						ApplicationName: "test-app",
						Limit:           10,
					})
					So(err, ShouldBeNil)
					So(nodes.TotalCount, ShouldEqual, 0)
					So(nodes.Result, ShouldHaveLength, 0)
				})
			})
		})
	})
}
