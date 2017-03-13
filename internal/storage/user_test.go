package storage

import (
	"testing"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/test"
)

func TestUser(t *testing.T) {
	conf := test.GetConfig()

	// Set a user secret so JWTs can be assigned
	SetUserSecret("DoWahDiddy")

	Convey("Given a clean database", t, func() {
		db, err := OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		Convey("When creating a user with an invalid username", func() {
			user := User{
				Username: "bad characters %",
			}
			_, err := CreateUser(db, &user, "somepassword")

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(errors.Cause(err), ShouldResemble, ErrUserInvalidUsername)
			})
		})

		Convey("When creating a user with an invalid password", func() {
			user := User{
				Username:   "okcharacters",
				IsAdmin:    false,
				SessionTTL: 40,
			}
			_, err := CreateUser(db, &user, "bad")

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(errors.Cause(err), ShouldResemble, ErrUserPasswordLength)
			})
		})

		Convey("When creating a user", func() {
			user := User{
				Username:   "goodusername111",
				IsAdmin:    false,
				SessionTTL: 20,
			}
			password := "somepassword"

			userId, err := CreateUser(db, &user, password)
			So(err, ShouldBeNil)

			Convey("It can be get by id", func() {
				user2, err := GetUser(db, userId)
				So(err, ShouldBeNil)
				So(user2.Username, ShouldResemble, user.Username)
				So(user2.IsAdmin, ShouldResemble, user.IsAdmin)
				So(user2.SessionTTL, ShouldResemble, user.SessionTTL)
			})

			Convey("It can be get by username", func() {
				user2, err := GetUserByUsername(db, user.Username)
				So(err, ShouldBeNil)
				So(user2.Username, ShouldResemble, user.Username)
				So(user2.IsAdmin, ShouldResemble, user.IsAdmin)
				So(user2.SessionTTL, ShouldResemble, user.SessionTTL)
			})

			Convey("Then get users returns a single user", func() {
				users, err := GetUsers(db, 10, 0, "")
				So(err, ShouldBeNil)
				So(users, ShouldHaveLength, 1)
				So(users[0].Username, ShouldResemble, user.Username)
				So(users[0].IsAdmin, ShouldResemble, user.IsAdmin)
				So(users[0].SessionTTL, ShouldResemble, user.SessionTTL)
			})

			Convey("Then get user count returns 1", func() {
				count, err := GetUserCount(db, "")
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then searching for 'good' returns a single item", func() {
				count, err := GetUserCount(db, "good")
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)

				users, err := GetUsers(db, 10, 0, "good")
				So(err, ShouldBeNil)
				So(users, ShouldHaveLength, 1)
			})

			Convey("Then searching for 'foo' returns 0 items", func() {
				count, err := GetUserCount(db, "foo")
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)

				users, err := GetUsers(db, 10, 0, "foo")
				So(err, ShouldBeNil)
				So(users, ShouldHaveLength, 0)
			})

			Convey("Then the user can log in", func() {
				jwt, err := LoginUser(db, user.Username, password)
				So(err, ShouldBeNil)
				So(jwt, ShouldNotBeNil)
			})

			Convey("When updating the user password", func() {
				password = "newrandompassword2*&^"
				So(UpdatePassword(db, user.ID, password), ShouldBeNil)

				Convey("Then the user can log in with the new password", func() {
					jwt, err := LoginUser(db, user.Username, password)
					So(err, ShouldBeNil)
					So(jwt, ShouldNotBeNil)
				})
			})

			Convey("When updating the user", func() {
				userUpdate := UserUpdate{
					ID:         user.ID,
					Username:   "newusername",
					IsAdmin:    true,
					SessionTTL: 30,
				}
				So(UpdateUser(db, userUpdate), ShouldBeNil)

				Convey("Then the user has been updated", func() {
					user2, err := GetUser(db, user.ID)
					So(err, ShouldBeNil)
					So(user2.Username, ShouldResemble, userUpdate.Username)
					So(user2.IsAdmin, ShouldResemble, userUpdate.IsAdmin)
					So(user2.SessionTTL, ShouldResemble, userUpdate.SessionTTL)
				})
			})

			Convey("When deleting the user", func() {
				So(DeleteUser(db, user.ID), ShouldBeNil)

				Convey("Then the user count returns 0", func() {
					count, err := GetUserCount(db, "")
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)
				})
			})

		})
	})
}
