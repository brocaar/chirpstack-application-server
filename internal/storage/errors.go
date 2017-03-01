package storage

import "errors"

// errors
var (
	ErrAlreadyExists          = errors.New("object already exists")
	ErrDoesNotExist           = errors.New("object does not exist")
	ErrApplicationInvalidName = errors.New("invalid application name")
	ErrNodeInvalidName        = errors.New("invalid node name")
	ErrNodeMaxRXDelay         = errors.New("max value of RXDelay is 15")
	ErrCFListTooManyChannels  = errors.New("too many channels in channel-list")
)
