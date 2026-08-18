package user

import "errors"

var (
	// ErrUserNotFound indicates that the requested user does not exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists indicates that a user with the given identifier already exists.
	ErrUserAlreadyExists = errors.New("user already exists")
)
