package storage

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/test"
)

func TestOrganization(t *testing.T) {
	conf := test.GetConfig()
	if err := Setup(conf); err != nil {
		t.Fatal(err)
	}

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	Convey("Given a clean database", t, func() {
		test.MustResetDB(DB().DB)

		Convey("When creating an organization with an invalid name", func() {
			org := Organization{
				Name:        "invalid name",
				DisplayName: "invalid organization",
			}
			err := CreateOrganization(DB(), &org)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(errors.Cause(err), ShouldResemble, ErrOrganizationInvalidName)
			})
		})

		Convey("When creating an organization", func() {
			org := Organization{
				Name:        "test-organization",
				DisplayName: "test organization",
			}
			So(CreateOrganization(DB(), &org), ShouldBeNil)
			org.CreatedAt = org.CreatedAt.Truncate(time.Millisecond).UTC()
			org.UpdatedAt = org.UpdatedAt.Truncate(time.Millisecond).UTC()

			Convey("Then it can be retrieved by its id", func() {
				o, err := GetOrganization(DB(), org.ID)
				So(err, ShouldBeNil)
				o.CreatedAt = o.CreatedAt.Truncate(time.Millisecond).UTC()
				o.UpdatedAt = o.UpdatedAt.Truncate(time.Millisecond).UTC()
				So(o, ShouldResemble, org)
			})

			Convey("When updating the organization", func() {
				org.Name = "test-organization-updated"
				org.DisplayName = "test organization updated"
				org.CanHaveGateways = true
				So(UpdateOrganization(DB(), &org), ShouldBeNil)
				org.CreatedAt = org.CreatedAt.Truncate(time.Millisecond).UTC()
				org.UpdatedAt = org.UpdatedAt.Truncate(time.Millisecond).UTC()

				Convey("Then it has been updated", func() {
					o, err := GetOrganization(DB(), org.ID)
					So(err, ShouldBeNil)
					o.CreatedAt = o.CreatedAt.Truncate(time.Millisecond).UTC()
					o.UpdatedAt = o.UpdatedAt.Truncate(time.Millisecond).UTC()
					So(o, ShouldResemble, org)
				})
			})

			// first org is created in the migrations
			Convey("Then get organization count returns 2", func() {
				count, err := GetOrganizationCount(DB(), "")
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 2)
			})

			Convey("Then get organizations returns the expected items", func() {
				items, err := GetOrganizations(DB(), 10, 0, "")
				So(err, ShouldBeNil)
				So(items, ShouldHaveLength, 2)
				items[1].CreatedAt = items[1].CreatedAt.Truncate(time.Millisecond).UTC()
				items[1].UpdatedAt = items[1].UpdatedAt.Truncate(time.Millisecond).UTC()
				So(items[1], ShouldResemble, org)
			})

			Convey("When deleting the organization", func() {
				So(DeleteOrganization(DB(), org.ID), ShouldBeNil)

				Convey("Then it has been deleted", func() {
					_, err := GetOrganization(DB(), org.ID)
					So(errors.Cause(err), ShouldResemble, ErrDoesNotExist)
				})
			})

			Convey("Then the organization has 0 users", func() {
				count, err := GetOrganizationUserCount(DB(), org.ID)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
			})

			Convey("When adding an user to the organization", func() {
				So(CreateOrganizationUser(DB(), org.ID, 1, false), ShouldBeNil) // admin user

				Convey("Then it can be retrieved", func() {
					u, err := GetOrganizationUser(DB(), org.ID, 1)
					So(err, ShouldBeNil)
					So(u.UserID, ShouldEqual, 1)
					So(u.Username, ShouldEqual, "admin")
					So(u.IsAdmin, ShouldBeFalse)
				})

				Convey("Then the organization has 1 user", func() {
					c, err := GetOrganizationUserCount(DB(), org.ID)
					So(err, ShouldBeNil)
					So(c, ShouldEqual, 1)

					users, err := GetOrganizationUsers(DB(), org.ID, 10, 0)
					So(err, ShouldBeNil)
					So(users, ShouldHaveLength, 1)
				})

				Convey("Then it can be updated", func() {
					So(UpdateOrganizationUser(DB(), org.ID, 1, true), ShouldBeNil) // admin user

					u, err := GetOrganizationUser(DB(), org.ID, 1)
					So(err, ShouldBeNil)
					So(u.UserID, ShouldEqual, 1)
					So(u.Username, ShouldEqual, "admin")
					So(u.IsAdmin, ShouldBeTrue)
				})

				Convey("Then it can be deleted", func() {
					So(DeleteOrganizationUser(DB(), org.ID, 1), ShouldBeNil) // admin user
					c, err := GetOrganizationUserCount(DB(), org.ID)
					So(err, ShouldBeNil)
					So(c, ShouldEqual, 0)
				})
			})

			Convey("Given an user", func() {
				user := User{
					Username: "testuser",
					IsActive: true,
					Email:    "foo@bar.com",
				}
				_, err := CreateUser(DB(), &user, "password123")
				So(err, ShouldBeNil)

				Convey("Then no organizations are related to this user", func() {
					c, err := GetOrganizationCountForUser(DB(), user.Username, "")
					So(err, ShouldBeNil)
					So(c, ShouldEqual, 0)

					orgs, err := GetOrganizationsForUser(DB(), user.Username, 10, 0, "")
					So(err, ShouldBeNil)
					So(orgs, ShouldHaveLength, 0)
				})

				Convey("When the user is linked to the organization", func() {
					So(CreateOrganizationUser(DB(), org.ID, user.ID, false), ShouldBeNil)

					Convey("Then the test organization is returned for the user", func() {
						c, err := GetOrganizationCountForUser(DB(), user.Username, "")
						So(err, ShouldBeNil)
						So(c, ShouldEqual, 1)

						orgs, err := GetOrganizationsForUser(DB(), user.Username, 10, 0, "")
						So(err, ShouldBeNil)
						So(orgs, ShouldHaveLength, 1)
						So(orgs[0].ID, ShouldEqual, org.ID)
					})
				})
			})
		})
	})
}
