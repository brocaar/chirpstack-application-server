package external

import (
	"github.com/brocaar/chirpstack-application-server/internal/api/external/auth"
	"golang.org/x/net/context"
)

type TestValidator struct {
	ctx            context.Context
	validatorFuncs []auth.ValidatorFunc
	returnError    error
	returnUsername string
	returnIsAdmin  bool
}

func (v *TestValidator) Validate(ctx context.Context, funcs ...auth.ValidatorFunc) error {
	v.ctx = ctx
	v.validatorFuncs = funcs
	return v.returnError
}

func (v *TestValidator) GetUsername(ctx context.Context) (string, error) {
	return v.returnUsername, v.returnError
}

func (v *TestValidator) GetIsAdmin(ctx context.Context) (bool, error) {
	return v.returnIsAdmin, v.returnError
}
