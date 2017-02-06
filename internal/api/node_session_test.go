package api

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	. "github.com/smartystreets/goconvey/convey"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
)

func TestNodeSessionAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database and an api instance", t, func() {
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
		api := NewNodeSessionAPI(lsCtx, validator)

		Convey("Given an application + node in the database", func() {
			app := storage.Application{
				Name: "test-app",
			}
			So(storage.CreateApplication(db, &app), ShouldBeNil)
			node := storage.Node{
				ApplicationID: app.ID,
				AppEUI:        [8]byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:        [8]byte{2, 2, 2, 2, 2, 2, 2, 2},
				AppSKey:       [16]byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2},
			}
			So(storage.CreateNode(db, node), ShouldBeNil)

			Convey("When creating a node-session for this node", func() {
				_, err := api.Create(ctx, &pb.CreateNodeSessionRequest{
					ApplicationName:    "test-app",
					DevAddr:            "01020304",
					DevEUI:             "0202020202020202",
					AppSKey:            "01010101010101010202020202020202",
					NwkSKey:            "02020202020202020101010101010101",
					FCntUp:             1,
					FCntDown:           2,
					RxDelay:            3,
					Rx1DROffset:        4,
					CFList:             []uint32{868400000},
					RxWindow:           pb.RXWindow_RX2,
					Rx2DR:              5,
					AdrInterval:        20,
					InstallationMargin: 5,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)

				Convey("Then the NetworkServerClient was called with the expected parameters", func() {
					So(nsClient.CreateNodeSessionChan, ShouldHaveLength, 1)
					So(<-nsClient.CreateNodeSessionChan, ShouldResemble, ns.CreateNodeSessionRequest{
						DevAddr:            []byte{1, 2, 3, 4},
						AppEUI:             []byte{1, 1, 1, 1, 1, 1, 1, 1},
						DevEUI:             []byte{2, 2, 2, 2, 2, 2, 2, 2},
						NwkSKey:            []byte{2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1},
						FCntUp:             1,
						FCntDown:           2,
						RxDelay:            3,
						Rx1DROffset:        4,
						CFList:             []uint32{868400000},
						RxWindow:           ns.RXWindow_RX2,
						Rx2DR:              5,
						AdrInterval:        20,
						InstallationMargin: 5,
					})
				})

				Convey("Then the AppSKey and DevAddr fields on the node are updated", func() {
					node, err := storage.GetNode(db, node.DevEUI)
					So(err, ShouldBeNil)
					So(node.AppSKey[:], ShouldResemble, []byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2})
					So(node.DevAddr[:], ShouldResemble, []byte{1, 2, 3, 4})
				})
			})

			Convey("When creating a node-session for this node but with a different application name", func() {
				_, err := api.Create(ctx, &pb.CreateNodeSessionRequest{
					ApplicationName: "test-app-2",
					DevAddr:         "01020304",
					DevEUI:          "0202020202020202",
					AppSKey:         "01010101010101010202020202020202",
					NwkSKey:         "02020202020202020101010101010101",
				})
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)

				Convey("Then an error is returned", func() {
					So(err, ShouldResemble, grpc.Errorf(codes.Unknown, "get application error: sql: no rows in result set"))
				})
			})

			Convey("When updating a node-session for this node", func() {
				_, err := api.Update(ctx, &pb.UpdateNodeSessionRequest{
					ApplicationName:    "test-app",
					DevAddr:            "04030201",
					DevEUI:             "0202020202020202",
					NwkSKey:            "01010101010101010202020202020202",
					AppSKey:            "02020202020202020101010101010101",
					FCntUp:             1,
					FCntDown:           2,
					RxDelay:            3,
					Rx1DROffset:        4,
					CFList:             []uint32{868400000},
					RxWindow:           pb.RXWindow_RX2,
					Rx2DR:              5,
					AdrInterval:        30,
					InstallationMargin: 10,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)

				Convey("Then the NetworkServerClient was called with the expected arguments", func() {
					So(nsClient.UpdateNodeSessionChan, ShouldHaveLength, 1)
					So(<-nsClient.UpdateNodeSessionChan, ShouldResemble, ns.UpdateNodeSessionRequest{
						DevAddr:            []byte{4, 3, 2, 1},
						AppEUI:             []byte{1, 1, 1, 1, 1, 1, 1, 1},
						DevEUI:             []byte{2, 2, 2, 2, 2, 2, 2, 2},
						NwkSKey:            []byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2},
						FCntUp:             1,
						FCntDown:           2,
						RxDelay:            3,
						Rx1DROffset:        4,
						CFList:             []uint32{868400000},
						RxWindow:           ns.RXWindow_RX2,
						Rx2DR:              5,
						AdrInterval:        30,
						InstallationMargin: 10,
					})
				})

				Convey("Then the AppSKey and DevAddr fields on the node are updated", func() {
					node, err := storage.GetNode(db, node.DevEUI)
					So(err, ShouldBeNil)
					So(node.AppSKey[:], ShouldResemble, []byte{2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1})
					So(node.DevAddr[:], ShouldResemble, []byte{4, 3, 2, 1})
				})
			})

			Convey("When updating a node-session for this node but with a different application name", func() {
				_, err := api.Update(ctx, &pb.UpdateNodeSessionRequest{
					ApplicationName: "test-app-2",
					DevAddr:         "01020304",
					DevEUI:          "0202020202020202",
					AppSKey:         "01010101010101010202020202020202",
					NwkSKey:         "02020202020202020101010101010101",
				})
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)

				Convey("Then an error is returned", func() {
					So(err, ShouldResemble, grpc.Errorf(codes.Unknown, "get application error: sql: no rows in result set"))
				})
			})

			Convey("Given a mocked node-session in the network-server", func() {
				nsClient.GetNodeSessionResponse = ns.GetNodeSessionResponse{
					DevAddr:            []byte{1, 2, 3, 4},
					AppEUI:             []byte{1, 1, 1, 1, 1, 1, 1, 1},
					DevEUI:             []byte{2, 2, 2, 2, 2, 2, 2, 2},
					NwkSKey:            []byte{2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1},
					FCntUp:             1,
					FCntDown:           2,
					RxDelay:            3,
					Rx1DROffset:        4,
					CFList:             []uint32{868400000},
					RxWindow:           ns.RXWindow_RX2,
					Rx2DR:              5,
					RelaxFCnt:          true,
					AdrInterval:        20,
					InstallationMargin: 5,
					NbTrans:            1,
					TxPower:            20,
				}

				Convey("When getting the node-session", func() {
					resp, err := api.Get(ctx, &pb.GetNodeSessionRequest{
						ApplicationName: "test-app",
						DevEUI:          node.DevEUI.String(),
					})
					So(err, ShouldBeNil)
					So(validator.ctx, ShouldResemble, ctx)
					So(validator.validatorFuncs, ShouldHaveLength, 3)

					Convey("Then the expected request was made", func() {
						So(nsClient.GetNodeSessionChan, ShouldHaveLength, 1)
						So(<-nsClient.GetNodeSessionChan, ShouldResemble, ns.GetNodeSessionRequest{
							DevEUI: node.DevEUI[:],
						})
					})

					Convey("Then the expected response is returned", func() {
						So(resp, ShouldResemble, &pb.GetNodeSessionResponse{
							DevAddr:            "01020304",
							AppEUI:             "0101010101010101",
							DevEUI:             "0202020202020202",
							AppSKey:            "01010101010101010202020202020202",
							NwkSKey:            "02020202020202020101010101010101",
							FCntUp:             1,
							FCntDown:           2,
							RxDelay:            3,
							Rx1DROffset:        4,
							CFList:             []uint32{868400000},
							RxWindow:           pb.RXWindow_RX2,
							Rx2DR:              5,
							RelaxFCnt:          true,
							AdrInterval:        20,
							InstallationMargin: 5,
							NbTrans:            1,
							TxPower:            20,
						})
					})
				})
			})

			Convey("Given a mocked GetRandomDevAddrResponse for the network-server client", func() {
				nsClient.GetRandomDevAddrResponse = ns.GetRandomDevAddrResponse{
					DevAddr: []byte{1, 2, 3, 4},
				}

				Convey("Then GetRandomDevAddr returns the expected result", func() {
					resp, err := api.GetRandomDevAddr(ctx, &pb.GetRandomDevAddrRequest{})
					So(err, ShouldBeNil)
					So(validator.ctx, ShouldResemble, ctx)
					So(validator.validatorFuncs, ShouldHaveLength, 1)
					So(resp.DevAddr, ShouldEqual, "01020304")
				})
			})

			Convey("When deleting a node-session", func() {
				_, err := api.Delete(ctx, &pb.DeleteNodeSessionRequest{
					ApplicationName: "test-app",
					DevEUI:          "0202020202020202",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)

				Convey("Then the expected request was made", func() {
					So(nsClient.DeleteNodeSessionChan, ShouldHaveLength, 1)
					So(<-nsClient.DeleteNodeSessionChan, ShouldResemble, ns.DeleteNodeSessionRequest{
						DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
					})
				})
			})
		})
	})
}
