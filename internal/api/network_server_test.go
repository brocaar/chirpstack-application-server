package api

import (
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

func TestNetworkServerAPI(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	common.DB = db

	Convey("Given a clean database and api instance", t, func() {
		test.MustResetDB(common.DB)

		nsClient := test.NewNetworkServerClient()
		common.NetworkServer = nsClient

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewNetworkServerAPI(validator)

		Convey("Then Create creates a network-server", func() {
			resp, err := api.Create(ctx, &pb.CreateNetworkServerRequest{
				Name:   "test ns",
				Server: "test-ns:1234",
			})
			So(err, ShouldBeNil)
			So(resp.Id, ShouldBeGreaterThan, 0)

			Convey("Then Get returns the network-server", func() {
				getResp, err := api.Get(ctx, &pb.GetNetworkServerRequest{
					Id: resp.Id,
				})
				So(err, ShouldBeNil)
				So(getResp.Name, ShouldEqual, "test ns")
				So(getResp.Server, ShouldEqual, "test-ns:1234")
			})

			Convey("Then Update updates the network-server", func() {
				_, err := api.Update(ctx, &pb.UpdateNetworkServerRequest{
					Id:     resp.Id,
					Name:   "updated-test-ns",
					Server: "updated-test-ns:1234",
				})
				So(err, ShouldBeNil)

				getResp, err := api.Get(ctx, &pb.GetNetworkServerRequest{
					Id: resp.Id,
				})
				So(err, ShouldBeNil)
				So(getResp.Name, ShouldEqual, "updated-test-ns")
				So(getResp.Server, ShouldEqual, "updated-test-ns:1234")
			})

			Convey("Then Delete deletes the network-server", func() {
				_, err := api.Delete(ctx, &pb.DeleteNetworkServerRequest{
					Id: resp.Id,
				})
				So(err, ShouldBeNil)

				_, err = api.Get(ctx, &pb.GetNetworkServerRequest{
					Id: resp.Id,
				})
				So(err, ShouldNotBeNil)
				So(grpc.Code(err), ShouldEqual, codes.NotFound)
			})

			Convey("Then List lists the network-servers", func() {
				listResp, err := api.List(ctx, &pb.ListNetworkServerRequest{
					Limit:  10,
					Offset: 0,
				})
				So(err, ShouldBeNil)
				So(listResp.TotalCount, ShouldEqual, 1)
				So(listResp.Result, ShouldHaveLength, 1)
				So(listResp.Result[0].Name, ShouldEqual, "test ns")
				So(listResp.Result[0].Server, ShouldEqual, "test-ns:1234")
			})
		})
	})
}
