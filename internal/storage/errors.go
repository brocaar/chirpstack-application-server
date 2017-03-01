package storage

import "errors"

// errors
var (
	ErrAlreadyExists          = errors.New("object already exists")
	ErrDoesNotExist           = errors.New("object does not exist")
	ErrApplicationInvalidName = errors.New("invalid application name")
)
