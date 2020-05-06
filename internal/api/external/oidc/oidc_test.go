package oidc

import (
	"context"
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
}
