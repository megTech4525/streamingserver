package stream

import "errors"

var ErrStreamNotFound = errors.New("stream not found")
var ErrNotFound = errors.New("user not found")
var ErrInvalidKey = errors.New("invalid stream key")
