package storage

import (
	"testing"
	"time"

	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOrganization(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database", t, func() {
		db, err := OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		Convey("When creating an organization with an invalid name", func() {
			org := Organization{
				Name:        "invalid name",
				DisplayName: "invalid organization",
			}
			err := CreateOrganization(db, &org)

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
			So(CreateOrganization(db, &org), ShouldBeNil)
			org.CreatedAt = org.CreatedAt.Truncate(time.Millisecond).UTC()
			org.UpdatedAt = org.UpdatedAt.Truncate(time.Millisecond).UTC()

			Convey("Then it can be retrieved by its id", func() {
				o, err := GetOrganization(db, org.ID)
				So(err, ShouldBeNil)
				o.CreatedAt = o.CreatedAt.Truncate(time.Millisecond).UTC()
				o.UpdatedAt = o.UpdatedAt.Truncate(time.Millisecond).UTC()
				So(o, ShouldResemble, org)
			})

			Convey("When updating the organization", func() {
				org.Name = "test-organization-updated"
				org.DisplayName = "test organization updated"
				org.CanHaveGateways = true
				So(UpdateOrganization(db, &org), ShouldBeNil)
				org.CreatedAt = org.CreatedAt.Truncate(time.Millisecond).UTC()
				org.UpdatedAt = org.UpdatedAt.Truncate(time.Millisecond).UTC()

				Convey("Then it has been updated", func() {
					o, err := GetOrganization(db, org.ID)
					So(err, ShouldBeNil)
					o.CreatedAt = o.CreatedAt.Truncate(time.Millisecond).UTC()
					o.UpdatedAt = o.UpdatedAt.Truncate(time.Millisecond).UTC()
					So(o, ShouldResemble, org)
				})
			})

			// first org is created in the migrations
			Convey("Then get organization count returns 2", func() {
				count, err := GetOrganizationCount(db, "")
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 2)
			})

			Convey("Then get organizations returns the expected items", func() {
				items, err := GetOrganizations(db, 10, 0, "")
				So(err, ShouldBeNil)
				So(items, ShouldHaveLength, 2)
				items[1].CreatedAt = items[1].CreatedAt.Truncate(time.Millisecond).UTC()
				items[1].UpdatedAt = items[1].UpdatedAt.Truncate(time.Millisecond).UTC()
				So(items[1], ShouldResemble, org)
			})

			Convey("When deleting the organization", func() {
				So(DeleteOrganization(db, org.ID), ShouldBeNil)

				Convey("Then it has been deleted", func() {
					_, err := GetOrganization(db, org.ID)
					So(errors.Cause(err), ShouldResemble, ErrDoesNotExist)
				})
			})

			Convey("Then the organization has 0 users", func() {
				count, err := GetOrganizationUserCount(db, org.ID)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
			})

			Convey("When adding an user to the organization", func() {
				So(CreateOrganizationUser(db, org.ID, 1, false), ShouldBeNil) // admin user

				Convey("Then it can be retrieved", func() {
					u, err := GetOrganizationUser(db, org.ID, 1)
					So(err, ShouldBeNil)
					So(u.UserID, ShouldEqual, 1)
					So(u.Username, ShouldEqual, "admin")
					So(u.IsAdmin, ShouldBeFalse)
				})

				Convey("Then the organization has 1 user", func() {
					c, err := GetOrganizationUserCount(db, org.ID)
					So(err, ShouldBeNil)
					So(c, ShouldEqual, 1)

					users, err := GetOrganizationUsers(db, org.ID, 10, 0)
					So(err, ShouldBeNil)
					So(users, ShouldHaveLength, 1)
				})

				Convey("Then it can be updated", func() {
					So(UpdateOrganizationUser(db, org.ID, 1, true), ShouldBeNil) // admin user

					u, err := GetOrganizationUser(db, org.ID, 1)
					So(err, ShouldBeNil)
					So(u.UserID, ShouldEqual, 1)
					So(u.Username, ShouldEqual, "admin")
					So(u.IsAdmin, ShouldBeTrue)
				})

				Convey("Then it can be deleted", func() {
					So(DeleteOrganizationUser(db, org.ID, 1), ShouldBeNil) // admin user
					c, err := GetOrganizationUserCount(db, org.ID)
					So(err, ShouldBeNil)
					So(c, ShouldEqual, 0)
				})
			})

			Convey("Given an user and application linked to the organization", func() {
				user := User{
					Username: "testuser",
					IsActive: true,
				}
				_, err := CreateUser(db, &user, "password123")
				So(err, ShouldBeNil)

				app := Application{
					OrganizationID: org.ID,
					Name:           "test-app",
				}
				So(CreateApplication(db, &app), ShouldBeNil)

				Convey("Then no organizations are related to this user", func() {
					c, err := GetOrganizationCountForUser(db, user.Username, "")
					So(err, ShouldBeNil)
					So(c, ShouldEqual, 0)

					orgs, err := GetOrganizationsForUser(db, user.Username, 10, 0, "")
					So(err, ShouldBeNil)
					So(orgs, ShouldHaveLength, 0)
				})

				Convey("When the user is linked to the application", func() {
					So(CreateUserForApplication(db, app.ID, user.ID, false), ShouldBeNil)

					Convey("Then the test organization is returned for the user", func() {
						c, err := GetOrganizationCountForUser(db, user.Username, "")
						So(err, ShouldBeNil)
						So(c, ShouldEqual, 1)

						orgs, err := GetOrganizationsForUser(db, user.Username, 10, 0, "")
						So(err, ShouldBeNil)
						So(orgs, ShouldHaveLength, 1)
						So(orgs[0].ID, ShouldEqual, org.ID)
					})
				})

				Convey("When the user is linked to the organization", func() {
					So(CreateOrganizationUser(db, org.ID, user.ID, false), ShouldBeNil)

					Convey("Then the test organization is returned for the user", func() {
						c, err := GetOrganizationCountForUser(db, user.Username, "")
						So(err, ShouldBeNil)
						So(c, ShouldEqual, 1)

						orgs, err := GetOrganizationsForUser(db, user.Username, 10, 0, "")
						So(err, ShouldBeNil)
						So(orgs, ShouldHaveLength, 1)
						So(orgs[0].ID, ShouldEqual, org.ID)
					})
				})

				Convey("When the user is linked to both the organization and application", func() {
					So(CreateUserForApplication(db, app.ID, user.ID, false), ShouldBeNil)
					So(CreateOrganizationUser(db, org.ID, user.ID, false), ShouldBeNil)

					Convey("Then the test organization is returned for the user", func() {
						c, err := GetOrganizationCountForUser(db, user.Username, "")
						So(err, ShouldBeNil)
						So(c, ShouldEqual, 1)

						orgs, err := GetOrganizationsForUser(db, user.Username, 10, 0, "")
						So(err, ShouldBeNil)
						So(orgs, ShouldHaveLength, 1)
						So(orgs[0].ID, ShouldEqual, org.ID)
					})
				})
			})
		})
	})
}
