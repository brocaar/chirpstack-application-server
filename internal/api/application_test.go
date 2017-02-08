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

func TestApplicationAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database and api instance", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		ctx := context.Background()
		lsCtx := common.Context{DB: db}
		validator := &TestValidator{}
		api := NewApplicationAPI(lsCtx, validator)

		Convey("When creating an application", func() {
			_, err := api.Create(ctx, &pb.CreateApplicationRequest{
				Name:        "test-app",
				Description: "A test application",
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 2)

			Convey("Then the application has been created", func() {
				app, err := api.Get(ctx, &pb.GetApplicationRequest{
					ApplicationName: "test-app",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 2)
				So(app, ShouldResemble, &pb.GetApplicationResponse{
					Name:        "test-app",
					Description: "A test application",
				})
			})

			Convey("Then listing the applications returns a single item", func() {
				apps, err := api.List(ctx, &pb.ListApplicationRequest{
					Limit:  10,
					Offset: 0,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				So(apps.Result, ShouldHaveLength, 1)
				So(apps.TotalCount, ShouldEqual, 1)
				So(apps.Result[0], ShouldResemble, &pb.GetApplicationResponse{
					Name:        "test-app",
					Description: "A test application",
				})
			})

			Convey("When updating the application", func() {
				_, err := api.Update(ctx, &pb.UpdateApplicationRequest{
					ApplicationName: "test-app",
					Name:            "test-app-updated",
					Description:     "An updated test description",
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 3)

				Convey("Then the application has been updated", func() {
					app, err := api.Get(ctx, &pb.GetApplicationRequest{
						ApplicationName: "test-app-updated",
					})
					So(err, ShouldBeNil)
					So(app, ShouldResemble, &pb.GetApplicationResponse{
						Name:        "test-app-updated",
						Description: "An updated test description",
					})
				})
			})

			Convey("When deleting the application", func() {
				_, err := api.Delete(ctx, &pb.DeleteApplicationRequest{
					ApplicationName: "test-app",
				})
				So(err, ShouldBeNil)

				Convey("Then the application has been deleted", func() {
					apps, err := api.List(ctx, &pb.ListApplicationRequest{Limit: 10})
					So(err, ShouldBeNil)
					So(apps.TotalCount, ShouldEqual, 0)
					So(apps.Result, ShouldHaveLength, 0)
				})
			})
		})
	})
}
