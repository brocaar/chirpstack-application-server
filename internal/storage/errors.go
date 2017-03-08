package storage

import "errors"

// errors
var (
	ErrAlreadyExists             = errors.New("object already exists")
	ErrDoesNotExist              = errors.New("object does not exist")
	ErrApplicationInvalidName    = errors.New("invalid application name")
	ErrNodeInvalidName           = errors.New("invalid node name")
	ErrNodeMaxRXDelay            = errors.New("max value of RXDelay is 15")
	ErrCFListTooManyChannels     = errors.New("too many channels in channel-list")
	ErrUserInvalidUsername       = errors.New("username name may only be composed of upper and lower case characters and digits")
	ErrUserPasswordLength        = errors.New("passwords must be at least 6 characters long")
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
)
