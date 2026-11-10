package user

import "errors"

var ErrNotFound = errors.New("user not found")
var ErrStreamNotFound = errors.New("stream not found")
var ErrInvalidCredentials = errors.New("invalid email or password")
