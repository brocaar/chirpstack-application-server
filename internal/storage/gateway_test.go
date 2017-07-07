package storage

import (
	"testing"
	"time"

	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lorawan"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGateway(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database woth an organization", t, func() {
		db, err := OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(db, &org), ShouldBeNil)

		Convey("When creating a gateway with an invalid name", func() {
			gw := Gateway{
				MAC:  lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				Name: "test gateway",
			}
			err := CreateGateway(db, &gw)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(errors.Cause(err), ShouldResemble, ErrGatewayInvalidName)
			})
		})

		Convey("When creating a gateway", func() {
			gw := Gateway{
				MAC:            lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				Name:           "test-gw",
				Description:    "test gateway",
				OrganizationID: org.ID,
			}
			So(CreateGateway(db, &gw), ShouldBeNil)
			gw.CreatedAt = gw.CreatedAt.Truncate(time.Millisecond).UTC()
			gw.UpdatedAt = gw.UpdatedAt.Truncate(time.Millisecond).UTC()

			Convey("Then it can be get by its MAC", func() {
				gw2, err := GetGateway(db, gw.MAC)
				So(err, ShouldBeNil)
				gw2.CreatedAt = gw2.CreatedAt.Truncate(time.Millisecond).UTC()
				gw2.UpdatedAt = gw2.UpdatedAt.Truncate(time.Millisecond).UTC()
				So(gw2, ShouldResemble, gw)
			})

			Convey("Then it can be updated", func() {
				gw.Name = "test-gw2"
				gw.Description = "updated test gateway"
				So(UpdateGateway(db, &gw), ShouldBeNil)
				gw.CreatedAt = gw.CreatedAt.Truncate(time.Millisecond).UTC()
				gw.UpdatedAt = gw.UpdatedAt.Truncate(time.Millisecond).UTC()

				gw2, err := GetGateway(db, gw.MAC)
				So(err, ShouldBeNil)
				gw2.CreatedAt = gw2.CreatedAt.Truncate(time.Millisecond).UTC()
				gw2.UpdatedAt = gw2.UpdatedAt.Truncate(time.Millisecond).UTC()
				So(gw2, ShouldResemble, gw)
			})

			Convey("Then it can be deleted", func() {
				So(DeleteGateway(db, gw.MAC), ShouldBeNil)
				_, err := GetGateway(db, gw.MAC)
				So(errors.Cause(err), ShouldResemble, ErrDoesNotExist)
			})

			Convey("Then getting the total gateway count returns 1", func() {
				c, err := GetGatewayCount(db)
				So(err, ShouldBeNil)
				So(c, ShouldEqual, 1)
			})

			Convey("Then getting all gateways returns the expected gateway", func() {
				gws, err := GetGateways(db, 10, 0)
				So(err, ShouldBeNil)
				So(gws, ShouldHaveLength, 1)
				So(gws[0].MAC, ShouldEqual, gw.MAC)
			})

			Convey("Then getting the total gateway count for the organization returns 1", func() {
				c, err := GetGatewayCountForOrganizationID(db, org.ID)
				So(err, ShouldBeNil)
				So(c, ShouldEqual, 1)
			})

			Convey("Then getting all gateways for the organization returns the exepected gateway", func() {
				gws, err := GetGatewaysForOrganizationID(db, org.ID, 10, 0)
				So(err, ShouldBeNil)
				So(gws, ShouldHaveLength, 1)
				So(gws[0].MAC, ShouldEqual, gw.MAC)
			})

			Convey("When creating an user", func() {
				user := User{
					Username: "testuser",
					IsActive: true,
				}
				_, err := CreateUser(db, &user, "password123")
				So(err, ShouldBeNil)

				Convey("Getting the gateway count for this user returns 0", func() {
					c, err := GetGatewayCountForUser(db, user.Username)
					So(err, ShouldBeNil)
					So(c, ShouldEqual, 0)
				})

				Convey("Then getting the gateways for this user returns 0 items", func() {
					gws, err := GetGatewaysForUser(db, user.Username, 10, 0)
					So(err, ShouldBeNil)
					So(gws, ShouldHaveLength, 0)
				})

				Convey("When assigning the user to the organization", func() {
					So(CreateOrganizationUser(db, org.ID, user.ID, false), ShouldBeNil)

					Convey("Getting the gateway count for this user returns 1", func() {
						c, err := GetGatewayCountForUser(db, user.Username)
						So(err, ShouldBeNil)
						So(c, ShouldEqual, 1)
					})

					Convey("Then getting the gateways for this user returns 1 item", func() {
						gws, err := GetGatewaysForUser(db, user.Username, 10, 0)
						So(err, ShouldBeNil)
						So(gws, ShouldHaveLength, 1)
						So(gws[0].MAC, ShouldEqual, gw.MAC)
					})
				})
			})
		})
	})
}
