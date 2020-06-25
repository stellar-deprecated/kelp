package api

// Query is an interface for a query that returns data
type Query interface {
	// QueryRow executes the query with the passed in runtime parameters
	QueryRow(args ...interface{}) (interface{}, error)
}
