package api

// Query is an interface for a query that returns data
type Query interface {
	// Name returns a constant string name with which to represent the query
	Name() string

	// QueryRow executes the query with the passed in runtime parameters
	QueryRow(args ...interface{}) (interface{}, error)
}
