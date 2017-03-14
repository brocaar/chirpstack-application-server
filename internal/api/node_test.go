package api

import (
	"testing"

	"golang.org/x/net/context"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
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

		Convey("When creating a node without a name set", func() {
			_, err := api.Create(ctx, &pb.CreateNodeRequest{
				ApplicationID:      app.ID,
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
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the DevEUI is used as name", func() {
				node, err := api.Get(ctx, &pb.GetNodeRequest{
					DevEUI: "0807060504030201",
				})
				So(err, ShouldBeNil)
				So(node.Name, ShouldEqual, "0807060504030201")
			})
		})

		Convey("When creating a node", func() {
			_, err := api.Create(ctx, &pb.CreateNodeRequest{
				ApplicationID:      app.ID,
				Name:               "test-node",
				Description:        "test node description",
				DevEUI:             "0807060504030201",
				AppEUI:             "0102030405060708",
				AppKey:             "01020304050607080102030405060708",
				IsABP:              true,
				IsClassC:           true,
				RxDelay:            1,
				Rx1DROffset:        3,
				RxWindow:           pb.RXWindow_RX2,
				Rx2DR:              3,
				AdrInterval:        20,
				InstallationMargin: 5,
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("The node has been created", func() {
				node, err := api.Get(ctx, &pb.GetNodeRequest{
					DevEUI: "0807060504030201",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				So(node, ShouldResemble, &pb.GetNodeResponse{
					Name:               "test-node",
					Description:        "test node description",
					DevEUI:             "0807060504030201",
					AppEUI:             "0102030405060708",
					AppKey:             "01020304050607080102030405060708",
					IsABP:              true,
					IsClassC:           true,
					RxDelay:            1,
					Rx1DROffset:        3,
					RxWindow:           pb.RXWindow_RX2,
					Rx2DR:              3,
					AdrInterval:        20,
					InstallationMargin: 5,
					ApplicationID:      app.ID,
				})
			})

			Convey("Then listing the nodes for the application returns a single items", func() {
				nodes, err := api.ListByApplicationID(ctx, &pb.ListNodeByApplicationIDRequest{
					ApplicationID: app.ID,
					Limit:         10,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				So(nodes.Result, ShouldHaveLength, 1)
				So(nodes.TotalCount, ShouldEqual, 1)
				So(nodes.Result[0], ShouldResemble, &pb.GetNodeResponse{
					Name:               "test-node",
					Description:        "test node description",
					DevEUI:             "0807060504030201",
					AppEUI:             "0102030405060708",
					AppKey:             "01020304050607080102030405060708",
					IsABP:              true,
					IsClassC:           true,
					RxDelay:            1,
					Rx1DROffset:        3,
					RxWindow:           pb.RXWindow_RX2,
					Rx2DR:              3,
					AdrInterval:        20,
					InstallationMargin: 5,
					ApplicationID:      app.ID,
				})
			})

			Convey("When updating the node", func() {
				_, err := api.Update(ctx, &pb.UpdateNodeRequest{
					ApplicationID:      app.ID,
					DevEUI:             "0807060504030201",
					Name:               "test-node-updated",
					Description:        "test node description updated",
					AppEUI:             "0102030405060708",
					AppKey:             "08070605040302010807060504030201",
					IsABP:              false,
					IsClassC:           false,
					RxDelay:            3,
					Rx1DROffset:        1,
					RxWindow:           pb.RXWindow_RX2,
					Rx2DR:              4,
					AdrInterval:        30,
					InstallationMargin: 10,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the node has been updated", func() {
					node, err := api.Get(ctx, &pb.GetNodeRequest{
						DevEUI: "0807060504030201",
					})
					So(err, ShouldBeNil)
					So(node, ShouldResemble, &pb.GetNodeResponse{
						Name:               "test-node-updated",
						Description:        "test node description updated",
						DevEUI:             "0807060504030201",
						AppEUI:             "0102030405060708",
						AppKey:             "08070605040302010807060504030201",
						IsABP:              false,
						IsClassC:           false,
						RxDelay:            3,
						Rx1DROffset:        1,
						RxWindow:           pb.RXWindow_RX2,
						Rx2DR:              4,
						AdrInterval:        30,
						InstallationMargin: 10,
						ApplicationID:      app.ID,
					})
				})
			})

			Convey("After deleting the node", func() {
				_, err := api.Delete(ctx, &pb.DeleteNodeRequest{
					DevEUI: "0807060504030201",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then an attempt was made to delete the node-session", func() {
					So(nsClient.DeleteNodeSessionChan, ShouldHaveLength, 1)
					So(<-nsClient.DeleteNodeSessionChan, ShouldResemble, ns.DeleteNodeSessionRequest{
						DevEUI: []byte{8, 7, 6, 5, 4, 3, 2, 1},
					})
				})

				Convey("Then listing the nodes returns zero nodes", func() {
					nodes, err := api.ListByApplicationID(ctx, &pb.ListNodeByApplicationIDRequest{
						ApplicationID: app.ID,
						Limit:         10,
					})
					So(err, ShouldBeNil)
					So(nodes.TotalCount, ShouldEqual, 0)
					So(nodes.Result, ShouldHaveLength, 0)
				})
			})

			Convey("When activating the node (ABP)", func() {
				_, err := api.Activate(ctx, &pb.ActivateNodeRequest{
					DevEUI:   "0807060504030201",
					DevAddr:  "01020304",
					AppSKey:  "01020304050607080102030405060708",
					NwkSKey:  "08070605040302010807060504030201",
					FCntUp:   10,
					FCntDown: 11,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then an attempt was made to delete the node-session", func() {
					So(nsClient.DeleteNodeSessionChan, ShouldHaveLength, 1)
					So(<-nsClient.DeleteNodeSessionChan, ShouldResemble, ns.DeleteNodeSessionRequest{
						DevEUI: []byte{8, 7, 6, 5, 4, 3, 2, 1},
					})
				})

				Convey("Then a node-session was created", func() {
					So(nsClient.CreateNodeSessionChan, ShouldHaveLength, 1)
					So(<-nsClient.CreateNodeSessionChan, ShouldResemble, ns.CreateNodeSessionRequest{
						DevAddr:            []uint8{1, 2, 3, 4},
						AppEUI:             []uint8{1, 2, 3, 4, 5, 6, 7, 8},
						DevEUI:             []uint8{8, 7, 6, 5, 4, 3, 2, 1},
						NwkSKey:            []uint8{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1},
						FCntUp:             10,
						FCntDown:           11,
						RxDelay:            1,
						Rx1DROffset:        3,
						RxWindow:           1,
						Rx2DR:              3,
						RelaxFCnt:          false,
						AdrInterval:        20,
						InstallationMargin: 5,
					})
				})

				Convey("Then the node was updated", func() {
					node, err := storage.GetNode(db, [8]byte{8, 7, 6, 5, 4, 3, 2, 1})
					So(err, ShouldBeNil)
					So(node.AppSKey, ShouldEqual, lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8})
					So(node.NwkSKey, ShouldEqual, lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1})
					So(node.DevAddr, ShouldEqual, lorawan.DevAddr{1, 2, 3, 4})
				})
			})
		})
	})
}
