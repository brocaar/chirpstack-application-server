package storage

import (
	"testing"

	"github.com/brocaar/lora-app-server/internal/test"
	. "github.com/smartystreets/goconvey/convey"
)

func TestChannelFunctions(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database", t, func() {
		db, err := OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		Convey("When creating a channel-list", func() {
			cl := ChannelList{
				Name: "test channel-list",
				Channels: []int64{
					868400000,
					868500000,
					868600000,
				},
			}
			So(CreateChannelList(db, &cl), ShouldBeNil)

			Convey("Then the channel-list exists", func() {
				cl2, err := GetChannelList(db, cl.ID)
				So(err, ShouldBeNil)
				So(cl2, ShouldResemble, cl)
			})

			Convey("When updating the channel-list", func() {
				cl.Name = "test channel-list changed"
				So(UpdateChannelList(db, cl), ShouldBeNil)

				Convey("Then the channel-list has been updated", func() {
					cl2, err := GetChannelList(db, cl.ID)
					So(err, ShouldBeNil)
					So(cl2, ShouldResemble, cl)
				})
			})

			Convey("Then listing the channel-lists returns 1 result", func() {
				lists, err := GetChannelLists(db, 10, 0)
				So(err, ShouldBeNil)
				So(lists, ShouldHaveLength, 1)
				So(lists[0], ShouldResemble, cl)
			})

			Convey("Then the channel-list count returns 1", func() {
				count, err := GetChannelListsCount(db)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("When deleting the channel-list", func() {
				So(DeleteChannelList(db, cl.ID), ShouldBeNil)

				Convey("Then the channel-list has been removed", func() {
					count, err := GetChannelListsCount(db)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)
				})
			})
		})
	})
}
