package api

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
)

func TestChannelListAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database and api instances", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		ctx := context.Background()
		lsCtx := common.Context{DB: db}
		validator := &TestValidator{}

		clAPI := NewChannelListAPI(lsCtx, validator)

		Convey("When creating a channel-list", func() {
			resp, err := clAPI.Create(ctx, &pb.CreateChannelListRequest{
				Name: "test channel-list",
				Channels: []uint32{
					868400000,
					868500000,
				},
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			clID := resp.Id

			Convey("Then the channel-list has been created", func() {
				cl, err := clAPI.Get(ctx, &pb.GetChannelListRequest{Id: clID})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				So(cl, ShouldResemble, &pb.GetChannelListResponse{
					Id:   clID,
					Name: "test channel-list",
					Channels: []uint32{
						868400000,
						868500000,
					},
				})
			})

			Convey("When updating the channel-list", func() {
				_, err := clAPI.Update(ctx, &pb.UpdateChannelListRequest{
					Id:   clID,
					Name: "test channel-list changed",
					Channels: []uint32{
						868400000,
						868500000,
						868600000,
					},
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the channel-list has been updated", func() {
					cl, err := clAPI.Get(ctx, &pb.GetChannelListRequest{Id: clID})
					So(err, ShouldBeNil)
					So(cl, ShouldResemble, &pb.GetChannelListResponse{
						Id:   clID,
						Name: "test channel-list changed",
						Channels: []uint32{
							868400000,
							868500000,
							868600000,
						},
					})
				})
			})

			Convey("Then listing the channel-lists returns 1 result", func() {
				resp, err := clAPI.List(ctx, &pb.ListChannelListRequest{Limit: 10, Offset: 0})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				So(resp.TotalCount, ShouldEqual, 1)
				So(resp.Result, ShouldHaveLength, 1)
				So(resp.Result[0], ShouldResemble, &pb.GetChannelListResponse{
					Id:   clID,
					Name: "test channel-list",
					Channels: []uint32{
						868400000,
						868500000,
					},
				})
			})

			Convey("When deleting the channel-list", func() {
				_, err := clAPI.Delete(ctx, &pb.DeleteChannelListRequest{Id: clID})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the channel-list has been deleted", func() {
					resp, err := clAPI.List(ctx, &pb.ListChannelListRequest{Limit: 10, Offset: 0})
					So(err, ShouldBeNil)

					So(resp.TotalCount, ShouldEqual, 0)
				})
			})
		})
	})
}
