package user

import "errors"

var (
	// ErrUserNotFound indicates that the requested user does not exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists indicates that a user with the given identifier already exists.
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrUserSuspended indicates that the user is suspended.
	ErrUserSuspended = errors.New("user is suspended")

	// ErrUserDeleted indicates that the user has been deleted.
	ErrUserDeleted = errors.New("user is deleted")

	// ErrInvalidAPIKey indicates that the provided API key is invalid.
	ErrInvalidAPIKey = errors.New("invalid API key")
)
