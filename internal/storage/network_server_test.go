package storage

import (
	"testing"
	"time"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/test"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNetworkServer(t *testing.T) {
	conf := test.GetConfig()
	db, err := OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	common.DB = db

	Convey("Given a clean database with an organization", t, func() {
		test.MustResetDB(db)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(db, &org), ShouldBeNil)

		Convey("Then CreateNetworkServer creates a network-server", func() {
			ns := NetworkServer{
				Name:   "test-ns",
				Server: "test-ns:123",
			}
			So(CreateNetworkServer(db, &ns), ShouldBeNil)
			ns.CreatedAt = ns.CreatedAt.UTC().Truncate(time.Millisecond)
			ns.UpdatedAt = ns.UpdatedAt.UTC().Truncate(time.Millisecond)

			Convey("Then GetNetworkServer returns the network-server", func() {
				nsGet, err := GetNetworkServer(db, ns.ID)
				So(err, ShouldBeNil)
				nsGet.CreatedAt = nsGet.CreatedAt.UTC().Truncate(time.Millisecond)
				nsGet.UpdatedAt = nsGet.UpdatedAt.UTC().Truncate(time.Millisecond)

				So(nsGet, ShouldResemble, ns)
			})

			Convey("Then UpdateNetworkServer updates the network-server", func() {
				ns.Name = "new-nw-server"
				ns.Server = "new-nw-server:123"
				So(UpdateNetworkServer(db, &ns), ShouldBeNil)
				ns.UpdatedAt = ns.UpdatedAt.UTC().Truncate(time.Millisecond)

				nsGet, err := GetNetworkServer(db, ns.ID)
				So(err, ShouldBeNil)
				nsGet.CreatedAt = nsGet.CreatedAt.UTC().Truncate(time.Millisecond)
				nsGet.UpdatedAt = nsGet.UpdatedAt.UTC().Truncate(time.Millisecond)

				So(nsGet, ShouldResemble, ns)
			})

			Convey("Then DeleteNetworkServer deletes the network-server", func() {
				So(DeleteNetworkServer(db, ns.ID), ShouldBeNil)
				_, err := GetNetworkServer(db, ns.ID)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, ErrDoesNotExist)
			})
		})
	})
}
