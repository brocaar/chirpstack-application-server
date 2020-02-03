package external

import (
	"github.com/brocaar/chirpstack-application-server/internal/api/external/auth"
	"github.com/gofrs/uuid"
	"golang.org/x/net/context"
)

type TestValidator struct {
	ctx            context.Context
	validatorFuncs []auth.ValidatorFunc
	returnError    error
	returnUsername string
	returnIsAdmin  bool
	returnSubject  string
	returnAPIKeyID uuid.UUID
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

func (v *TestValidator) GetSubject(ctx context.Context) (string, error) {
	return v.returnSubject, v.returnError
}

func (v *TestValidator) GetAPIKeyID(ctx context.Context) (uuid.UUID, error) {
	return v.returnAPIKeyID, v.returnError
}
