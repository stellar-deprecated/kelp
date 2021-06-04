package kelpos

import (
	"fmt"
)

// User is a struct that represents a user
type User struct {
	// for now this is just userID, but later this may be expanded to become host:userID (UUID) for example
	ID string
}

// MakeUser is a factory method to create a User
func MakeUser(
	ID string,
) *User {
	return &User{
		ID: ID,
	}
}

// String is the standard stringer method
func (u *User) String() string {
	return fmt.Sprintf("User[ID=%s]", u.ID)
}
