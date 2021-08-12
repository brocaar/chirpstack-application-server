package auth

import (
	"testing"
	"time"

	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"

	"github.com/gofrs/uuid"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func testValidator(pass bool, err error) ValidatorFunc {
	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return pass, err
	}
}

func TestJWTValidator(t *testing.T) {
	assert := require.New(t)

	apiKeyID, err := uuid.NewV4()
	assert.NoError(err)

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
			Claims:        Claims{APIKeyID: apiKeyID, StandardClaims: jwt.StandardClaims{Audience: "as"}},
			ValidatorFunc: testValidator(true, nil),
		},
		{
			Description:   "valid key and expired token",
			Key:           v.secret,
			Claims:        Claims{StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Unix() - 1, Audience: "as"}},
			ValidatorFunc: testValidator(true, nil),
			Error:         "token is expired by 1s",
		},
		{
			Description:   "invalid key",
			Key:           "differentsecret",
			Claims:        Claims{StandardClaims: jwt.StandardClaims{Audience: "as"}},
			ValidatorFunc: testValidator(true, nil),
			Error:         "signature is invalid",
		},
		{
			Description:   "valid key but failing validation",
			Key:           v.secret,
			Claims:        Claims{StandardClaims: jwt.StandardClaims{Audience: "as"}},
			ValidatorFunc: testValidator(false, nil),
			Error:         "not authorized",
		},
		{
			Description:   "valid key but validation returning error",
			Key:           v.secret,
			Claims:        Claims{StandardClaims: jwt.StandardClaims{Audience: "as"}},
			ValidatorFunc: testValidator(true, errors.New("boom")),
			Error:         "boom",
		},
	}

	for _, tst := range testTable {
		t.Run(tst.Description, func(t *testing.T) {
			assert := require.New(t)

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, tst.Claims)
			ss, err := token.SignedString([]byte(tst.Key))
			assert.NoError(err)

			ctx := context.Background()
			ctx = metadata.NewIncomingContext(ctx, metadata.MD{
				"authorization": []string{ss},
			})

			if tst.Error != "" {
				assert.Equal(tst.Error, errors.Cause(v.Validate(ctx, tst.ValidatorFunc)).Error())
			} else {
				apiKeyID, err := v.GetAPIKeyID(ctx)
				assert.NoError(err)
				assert.Equal(tst.Claims.APIKeyID, apiKeyID)
			}
		})
	}

}
