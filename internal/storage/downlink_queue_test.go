package storage

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/test"
)

func TestGetNextDownlinkQeueueItem(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database with application and node", t, func() {
		db, err := OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		app := Application{
			Name: "test",
		}
		So(CreateApplication(db, &app), ShouldBeNil)

		node := Node{
			ApplicationID: app.ID,
			Name:          "test-node",
			DevEUI:        [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(CreateNode(db, node), ShouldBeNil)

		Convey("Given a set of four queue items", func() {
			queue := []*DownlinkQueueItem{
				{DevEUI: node.DevEUI, Reference: "a", Data: []byte{1, 2, 3, 4, 5, 6, 7}},
				{DevEUI: node.DevEUI, Reference: "b", Data: []byte{1, 2, 3, 4, 5, 6}},
				{DevEUI: node.DevEUI, Reference: "c", Data: []byte{1, 2, 3, 4, 5}},
				{DevEUI: node.DevEUI, Reference: "d", Data: []byte{1, 2, 3, 4}},
			}
			for _, qi := range queue {
				So(CreateDownlinkQueueItem(db, qi), ShouldBeNil)
			}

			Convey("When getting calling GetNextDownlinkQueueItem with maxPayloadSize=7", func() {
				qi, err := GetNextDownlinkQueueItem(db, node.DevEUI, 7)
				So(err, ShouldBeNil)
				Convey("Then the first item is returned", func() {
					So(qi, ShouldResemble, queue[0])
				})
			})

			Convey("When getting calling GetNextDownlinkQueueItem with maxPayloadSize=5", func() {
				qi, err := GetNextDownlinkQueueItem(db, node.DevEUI, 5)
				So(err, ShouldBeNil)
				Convey("Then the third item is returned", func() {
					So(qi, ShouldResemble, queue[2])
				})
			})

			Convey("When getting calling GetNextDownlinkQueueItem with maxPayloadSize=3", func() {
				qi, err := GetNextDownlinkQueueItem(db, node.DevEUI, 3)
				So(err, ShouldBeNil)
				Convey("Then no item is returned", func() {
					So(qi, ShouldBeNil)
				})
			})
		})
	})
}

func TestDownlinkQueueFuncs(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database with application and node", t, func() {
		db, err := OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		app := Application{
			Name: "test",
		}
		So(CreateApplication(db, &app), ShouldBeNil)

		node := Node{
			ApplicationID: app.ID,
			Name:          "test-node",
			DevEUI:        [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(CreateNode(db, node), ShouldBeNil)

		Convey("When creating a downlink queue item", func() {
			qi := DownlinkQueueItem{
				DevEUI:    node.DevEUI,
				Reference: "abcd1234-1",
				Confirmed: true,
				FPort:     10,
				Data:      []byte{1, 2, 3, 4},
			}
			So(CreateDownlinkQueueItem(db, &qi), ShouldBeNil)

			Convey("Then GetDownlinkQueueSize returns 1", func() {
				size, err := GetDownlinkQueueSize(db, node.DevEUI)
				So(err, ShouldBeNil)
				So(size, ShouldEqual, 1)
			})

			Convey("Then the downlink queue item can be retrieved", func() {
				qi2, err := GetDownlinkQueueItem(db, qi.ID)
				So(err, ShouldBeNil)
				So(qi2, ShouldResemble, qi)
			})

			Convey("Then getting a pending downlink queue item returns an error", func() {
				_, err := GetPendingDownlinkQueueItem(db, node.DevEUI)
				So(err, ShouldNotBeNil)
			})

			Convey("When updating a downlink queue item to pending", func() {
				qi.Pending = true
				So(UpdateDownlinkQueueItem(db, qi), ShouldBeNil)

				Convey("Then the downlink queue item has been updated", func() {
					qi2, err := GetDownlinkQueueItem(db, qi.ID)
					So(err, ShouldBeNil)
					So(qi2, ShouldResemble, qi)
				})

				Convey("Then getting a pending downlink queue item returns this item", func() {
					qi2, err := GetPendingDownlinkQueueItem(db, node.DevEUI)
					So(err, ShouldBeNil)
					So(qi2, ShouldResemble, qi)
				})
			})

			Convey("When listing the queue items", func() {
				Convey("Then it returns one item for the existing DevEUI", func() {
					items, err := GetDownlinkQueueItems(db, node.DevEUI)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 1)
					So(items[0], ShouldResemble, qi)
				})

				Convey("Then it returns zero items for a non-existing DevEUI", func() {
					items, err := GetDownlinkQueueItems(db, [8]byte{8, 7, 6, 5, 4, 3, 2, 1})
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 0)
				})
			})

			Convey("When deleting the downlink queue item", func() {
				So(DeleteDownlinkQueueItem(db, qi.ID), ShouldBeNil)

				Convey("Then the item has been deleted", func() {
					_, err := GetDownlinkQueueItem(db, qi.ID)
					So(err, ShouldNotBeNil)
				})
			})

			Convey("When deleting all queue items for a node", func() {
				So(DeleteDownlinkQueueItemsForDevEUI(db, node.DevEUI), ShouldBeNil)

				Convey("Then the queue is empty", func() {
					items, err := GetDownlinkQueueItems(db, node.DevEUI)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 0)
				})
			})
		})
	})
}
