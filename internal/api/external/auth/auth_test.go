package auth

import (
	"testing"
	"time"

	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func testValidator(pass bool, err error) ValidatorFunc {
	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return pass, err
	}
}

func TestJWTValidator(t *testing.T) {
	Convey("Given a JWT validator", t, func() {
		v := JWTValidator{
			secret:    "verysecret",
			algorithm: "HS256",
		}

		testTable := []struct {
			Description   string
			Key           string
			Claims        Claims
			ValidatorFunc ValidatorFunc
			Error         string
		}{
			{
				Description:   "valid key and passing validation",
				Key:           v.secret,
				Claims:        Claims{Username: "foobar"},
				ValidatorFunc: testValidator(true, nil),
			},
			{
				Description:   "valid key and expired token",
				Key:           v.secret,
				Claims:        Claims{StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Unix() - 1}},
				ValidatorFunc: testValidator(true, nil),
				Error:         "token is expired by 1s",
			},
			{
				Description:   "invalid key",
				Key:           "differentsecret",
				Claims:        Claims{},
				ValidatorFunc: testValidator(true, nil),
				Error:         "signature is invalid",
			},
			{
				Description:   "valid key but failing validation",
				Key:           v.secret,
				Claims:        Claims{},
				ValidatorFunc: testValidator(false, nil),
				Error:         "not authorized",
			},
			{
				Description:   "valid key but validation returning error",
				Key:           v.secret,
				Claims:        Claims{},
				ValidatorFunc: testValidator(true, errors.New("boom")),
				Error:         "boom",
			},
		}

		for _, test := range testTable {
			Convey("Test: "+test.Description, func() {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, test.Claims)
				ss, err := token.SignedString([]byte(test.Key))
				So(err, ShouldBeNil)

				ctx := context.Background()
				ctx = metadata.NewIncomingContext(ctx, metadata.MD{
					"authorization": []string{ss},
				})

				if test.Error != "" {
					So(errors.Cause(v.Validate(ctx, test.ValidatorFunc)).Error(), ShouldResemble, test.Error)
				} else {
					username, err := v.GetUsername(ctx)
					So(err, ShouldBeNil)
					So(username, ShouldEqual, test.Claims.Username)
				}
			})
		}
	})
}
