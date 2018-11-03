package ucp

// operation interface is implemented by a valid UCP operation
// such as session management, submit short message, delivery notification etc.
type operation interface {
	// Type() returns the operation type either "R" for result or "O" for operation
	Type() []byte
	// Code() returns the operation code i.e. 51, 52, 53, 60
	Code() []byte
}
