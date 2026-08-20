package user

// UserStatus represents the lifecycle state of a user.
type UserStatus string

const (
	// Active indicates the user is currently active.
	Active UserStatus = "active"

	// Suspended indicates the user is currently suspended.
	Suspended UserStatus = "suspended"

	// Deleted indicates the user has been deleted.
	Deleted UserStatus = "deleted"
)
