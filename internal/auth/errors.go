package auth

import "errors"

var (
	ErrInvalidToken        = errors.New("Invalid Token")
	ErrInvalidClaims       = errors.New("Invalid Claims")
	ErrInvalidTokenSubject = errors.New("Invalid Token User")
)
