package storage

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/test"
)

func TestDownlinkQueueFuncs(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database with node", t, func() {
		db, err := OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		node := Node{
			DevEUI: [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(CreateNode(db, node), ShouldBeNil)

		Convey("When creating a downlink queue item", func() {
			qi := DownlinkQueueItem{
				DevEUI:    node.DevEUI,
				Confirmed: true,
				FPort:     10,
				Data:      []byte{1, 2, 3, 4},
			}
			So(CreateDownlinkQueueItem(db, &qi), ShouldBeNil)

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
		})
	})
}
