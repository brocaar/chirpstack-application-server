package oidc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAuthenticator(t *testing.T) {
	assert := require.New(t)

	t.Run("Disabled", func(t *testing.T) {
		// make sure that this errors when not setup
		_, err := newAuthenticator(context.Background())
		assert.Equal("openid connect is not properly configured", err.Error())
	})

	t.Run("Claims Unmarshalling", func(t *testing.T) {
		emailVerifiedAsString := `
		{
			"sub": "chirpstack-oidc",
			"name": "brocaar",
			"email": "chirpstack@chirpstack.io",
			"email_verified": "true",
			"user_info_claims": {
				"some_key": "some_value",
				"some_other": "another_value"
			}
		}`

		var userStr User

		err := json.Unmarshal([]byte(emailVerifiedAsString), &userStr)
		assert.NoError(err)
		assert.Equal(true, userStr.EmailVerified, "string parsing should return true")

		emailVerifiedAsBool := `
		{
			"sub": "chirpstack-oidc",
			"name": "brocaar",
			"email": "chirpstack@chirpstack.io",
			"email_verified": true,
			"user_info_claims": {
				"some_key": "some_value",
				"some_other": "another_value"
			}
		}`

		var userBool User
		err = json.Unmarshal([]byte(emailVerifiedAsBool), &userBool)
		assert.NoError(err)
		assert.Equal(true, userBool.EmailVerified, "bool parsing should return true")

		emailVerifiedMissing := `{
			"sub": "chirpstack-oidc",
			"name": "brocaar",
			"email": "chirpstack@chirpstack.io",
			"user_info_claims": {
				"some_key": "some_value",
				"some_other": "another_value"
			}
		}`

		var userMiss User
		err = json.Unmarshal([]byte(emailVerifiedMissing), &userMiss)
		assert.NoError(err)
		assert.Equal(false, userMiss.EmailVerified, "should default to false if missing")
	})
}
