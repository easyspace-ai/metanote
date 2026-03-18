package auth

import "errors"

var (
	ErrUsernameOrEmailTaken = errors.New("username or email already registered")
	ErrInvalidCredentials  = errors.New("incorrect username or password")
)
