package storage

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// Action defines the action type.
type Action int

// Possible actions
const (
	Select Action = iota
	Insert
	Update
	Delete
	Scan
)

// errors
var (
	ErrAlreadyExists                   = errors.New("object already exists")
	ErrDoesNotExist                    = errors.New("object does not exist")
	ErrUsedByOtherObjects              = errors.New("this object is used by other objects, remove them first")
	ErrApplicationInvalidName          = errors.New("invalid application name")
	ErrNodeInvalidName                 = errors.New("invalid node name")
	ErrNodeMaxRXDelay                  = errors.New("max value of RXDelay is 15")
	ErrCFListTooManyChannels           = errors.New("too many channels in channel-list")
	ErrUserInvalidUsername             = errors.New("username name may only be composed of upper and lower case characters and digits")
	ErrUserPasswordLength              = errors.New("passwords must be at least 6 characters long")
	ErrInvalidUsernameOrPassword       = errors.New("invalid username or password")
	ErrOrganizationInvalidName         = errors.New("invalid organization name")
	ErrGatewayInvalidName              = errors.New("invalid gateway name")
	ErrInvalidEmail                    = errors.New("invalid e-mail")
	ErrInvalidGatewayDiscoveryInterval = errors.New("invalid gateway-discovery interval, it must be greater than 0")
	ErrDeviceProfileInvalidName        = errors.New("invalid device-profile name")
	ErrServiceProfileInvalidName       = errors.New("invalid service-profile name")
	ErrMulticastGroupInvalidName       = errors.New("invalid multicast-group name")
	ErrOrganizationMaxDeviceCount      = errors.New("organization reached max. device count")
	ErrOrganizationMaxGatewayCount     = errors.New("organization reached max. gateway count")
	ErrNetworkServerInvalidName        = errors.New("invalid network-server name")
	ErrAPIKeyInvalidName               = errors.New("invalid API Key name")
)

func handlePSQLError(action Action, err error, description string) error {
	if err == sql.ErrNoRows {
		return ErrDoesNotExist
	}

	switch err := err.(type) {
	case *pq.Error:
		switch err.Code.Name() {
		case "unique_violation":
			return ErrAlreadyExists
		case "foreign_key_violation":
			switch action {
			case Delete:
				return ErrUsedByOtherObjects
			default:
				return ErrDoesNotExist
			}
		}
	}

	return errors.Wrap(err, description)
}
