package identity

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmptyPhone   = errors.New("phone hash required")
)
