package storage

import (
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/test"
)

func TestValidateDevNonce(t *testing.T) {
	Convey("Given a Node", t, func() {
		n := Node{
			Name:          "test-node",
			UsedDevNonces: make([][2]byte, 0),
		}

		Convey("Then any given dev-nonce is valid", func() {
			dn := [2]byte{1, 2}
			So(n.ValidateDevNonce(dn), ShouldBeTrue)

			Convey("Then the dev-nonce is added to the used nonces", func() {
				So(n.UsedDevNonces, ShouldContain, dn)
			})
		})

		Convey("Given a Node has 10 used nonces", func() {
			n.UsedDevNonces = [][2]byte{
				{1, 1},
				{2, 2},
				{3, 3},
				{4, 4},
				{5, 5},
				{6, 6},
				{7, 7},
				{8, 8},
				{9, 9},
				{10, 10},
			}

			Convey("Then an already used nonce is invalid", func() {
				So(n.ValidateDevNonce([2]byte{1, 1}), ShouldBeFalse)

				Convey("Then the UsedDevNonces still has length 10", func() {
					So(n.UsedDevNonces, ShouldHaveLength, 10)
				})
			})

			Convey("Then a new nonce is valid", func() {
				So(n.ValidateDevNonce([2]byte{0, 0}), ShouldBeTrue)

				Convey("Then the UsedDevNonces has length 11", func() {
					So(n.UsedDevNonces, ShouldHaveLength, 11)
				})
			})
		})
	})
}

func TestGetCFListForNode(t *testing.T) {
	conf := test.GetConfig()
	Convey("Given an application, node (without channel-list) and channel-list with 2 channels", t, func() {
		db, err := OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)
		ctx := common.Context{
			DB: db,
		}
		channelList := ChannelList{
			Name: "test channels",
			Channels: []int64{
				868400000,
				868500000,
			},
		}
		So(CreateChannelList(ctx.DB, &channelList), ShouldBeNil)
		app := Application{
			Name: "test",
		}
		So(CreateApplication(db, &app), ShouldBeNil)
		node := Node{
			ApplicationID: app.ID,
			Name:          "test-node",
			DevEUI:        [8]byte{8, 7, 6, 5, 4, 3, 2, 1},
		}
		So(CreateNode(ctx.DB, node), ShouldBeNil)
		Convey("Then GetCFListForNode returns nil", func() {
			cFList, err := GetCFListForNode(ctx.DB, node)
			So(err, ShouldBeNil)
			So(cFList, ShouldBeNil)
		})
		Convey("Given the node has the channel-list configured", func() {
			node.ChannelListID = &channelList.ID
			So(UpdateNode(ctx.DB, node), ShouldBeNil)
			Convey("Then GetCFListForNode returns the CFList with the configured channels", func() {
				cFList, err := GetCFListForNode(ctx.DB, node)
				So(err, ShouldBeNil)
				So(cFList, ShouldResemble, &lorawan.CFList{
					868400000,
					868500000,
					0,
					0,
					0,
				})
			})
		})
	})
}

func TestNodeMethods(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database with an application", t, func() {
		db, err := OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		app := Application{
			Name: "test",
		}
		So(CreateApplication(db, &app), ShouldBeNil)

		Convey("The application doesn't contain any nodes", func() {
			count, err := GetNodesCountForApplicationID(db, app.ID)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)

			nodes, err := GetNodesForApplicationID(db, app.ID, 10, 0)
			So(err, ShouldBeNil)
			So(nodes, ShouldHaveLength, 0)
		})

		Convey("When creating a node with an invalid name", func() {
			node := Node{
				ApplicationID: app.ID,
				Name:          "test node",
				Description:   "test node description",
				DevEUI:        [8]byte{8, 7, 6, 5, 4, 3, 2, 1},
				AppKey:        [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},

				RXDelay:            2,
				RX1DROffset:        3,
				RXWindow:           RX2,
				RX2DR:              3,
				ADRInterval:        20,
				InstallationMargin: 5,
			}
			err := CreateNode(db, node)

			Convey("Then a validation error is returned", func() {
				So(errors.Cause(err), ShouldResemble, ErrNodeInvalidName)
			})
		})

		Convey("When creating a node", func() {
			node := Node{
				ApplicationID: app.ID,
				Name:          "test-node",
				Description:   "test node description",
				DevEUI:        [8]byte{8, 7, 6, 5, 4, 3, 2, 1},
				AppKey:        [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
				IsABP:         true,
				IsClassC:      true,

				RXDelay:            2,
				RX1DROffset:        3,
				RXWindow:           RX2,
				RX2DR:              3,
				ADRInterval:        20,
				InstallationMargin: 5,
			}
			So(CreateNode(db, node), ShouldBeNil)

			Convey("It can be get by DevEUI", func() {
				node2, err := GetNode(db, node.DevEUI)
				node2.UsedDevNonces = nil
				So(err, ShouldBeNil)
				So(node2, ShouldResemble, node)
			})

			Convey("Then getting all nodes for the application returns this node", func() {
				nodes, err := GetNodesForApplicationID(db, app.ID, 10, 0)
				So(err, ShouldBeNil)
				So(nodes, ShouldHaveLength, 1)
				nodes[0].UsedDevNonces = nil
				So(nodes[0], ShouldResemble, node)
			})

			Convey("Then the nodes count for the application returns 1", func() {
				count, err := GetNodesCountForApplicationID(db, app.ID)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("When updating the node", func() {
				node.AppKey = [16]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
				node.IsABP = false
				So(UpdateNode(db, node), ShouldBeNil)

				Convey("Then the nodes has been updated", func() {
					node2, err := GetNode(db, node.DevEUI)
					So(err, ShouldBeNil)
					node2.UsedDevNonces = nil
					So(node2, ShouldResemble, node)
				})
			})

			Convey("When deleting the node", func() {
				So(DeleteNode(db, node.DevEUI), ShouldBeNil)

				Convey("Then get nodes count returns 0", func() {
					count, err := GetNodesCountForApplicationID(db, app.ID)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)
				})
			})

			Convey("When updating the node with UseApplicationSettings=true", func() {
				node.UseApplicationSettings = true
				So(UpdateNode(db, node), ShouldBeNil)

				Convey("Then the application settings are set", func() {
					node2, err := GetNode(db, node.DevEUI)
					So(err, ShouldBeNil)
					So(node2, ShouldResemble, Node{
						ApplicationID:          app.ID,
						UseApplicationSettings: true,
						Name:        "test-node",
						Description: "test node description",
						DevEUI:      [8]byte{8, 7, 6, 5, 4, 3, 2, 1},
						AppKey:      [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
					})
				})

				Convey("When updating the application settings", func() {
					app.IsClassC = true
					So(UpdateApplication(db, app), ShouldBeNil)

					Convey("Then the node has been updated", func() {
						node2, err := GetNode(db, node.DevEUI)
						So(err, ShouldBeNil)
						So(node2, ShouldResemble, Node{
							ApplicationID:          app.ID,
							UseApplicationSettings: true,
							Name:        "test-node",
							Description: "test node description",
							DevEUI:      [8]byte{8, 7, 6, 5, 4, 3, 2, 1},
							AppKey:      [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
							IsClassC:    true,
						})
					})
				})
			})
		})
	})
}
