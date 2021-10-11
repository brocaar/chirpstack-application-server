package helpers

import (
	"reflect"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/brocaar/chirpstack-application-server/internal/integration/http"
	"github.com/brocaar/chirpstack-application-server/internal/integration/influxdb"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

var errToCode = map[error]codes.Code{
	storage.ErrAlreadyExists:                   codes.AlreadyExists,
	storage.ErrDoesNotExist:                    codes.NotFound,
	storage.ErrUsedByOtherObjects:              codes.FailedPrecondition,
	storage.ErrApplicationInvalidName:          codes.InvalidArgument,
	storage.ErrNodeInvalidName:                 codes.InvalidArgument,
	storage.ErrNodeMaxRXDelay:                  codes.InvalidArgument,
	storage.ErrCFListTooManyChannels:           codes.InvalidArgument,
	storage.ErrUserInvalidUsername:             codes.InvalidArgument,
	storage.ErrUserPasswordLength:              codes.InvalidArgument,
	storage.ErrInvalidUsernameOrPassword:       codes.Unauthenticated,
	storage.ErrInvalidEmail:                    codes.InvalidArgument,
	storage.ErrInvalidGatewayDiscoveryInterval: codes.InvalidArgument,
	storage.ErrDeviceProfileInvalidName:        codes.InvalidArgument,
	storage.ErrServiceProfileInvalidName:       codes.InvalidArgument,
	storage.ErrMulticastGroupInvalidName:       codes.InvalidArgument,
	storage.ErrOrganizationMaxDeviceCount:      codes.FailedPrecondition,
	storage.ErrOrganizationMaxGatewayCount:     codes.FailedPrecondition,
	storage.ErrNetworkServerInvalidName:        codes.InvalidArgument,
	storage.ErrAPIKeyInvalidName:               codes.InvalidArgument,
	http.ErrInvalidHeaderName:                  codes.InvalidArgument,
	influxdb.ErrInvalidPrecision:               codes.InvalidArgument,
}

// ErrToRPCError converts the given error into a gRPC error.
func ErrToRPCError(err error) error {
	cause := errors.Cause(err)

	// if the err has already a gRPC status (it is a gRPC error), just
	// return the error.
	if code := status.Code(cause); code != codes.Unknown {
		return cause
	}

	// Some error values (error returned by Otto) can not be hashed:
	// https://github.com/brocaar/chirpstack-application-server/issues/631
	k := reflect.TypeOf(cause).Kind()
	if k == reflect.Slice {
		return grpc.Errorf(codes.Unknown, cause.Error())
	}

	code, ok := errToCode[cause]
	if !ok {
		code = codes.Unknown
	}
	return grpc.Errorf(code, cause.Error())
}
