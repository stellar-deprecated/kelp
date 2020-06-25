package api

// SimpleMockQuery is a helper that allows a test to force a return value from this query
type SimpleMockQuery struct {
	name        string
	returnValue interface{}
	returnError error
}

func (q *SimpleMockQuery) setName(name string) {
	q.name = name
}

func (q *SimpleMockQuery) setOutput(value interface{}, err error) {
	q.returnValue = value
	q.returnError = err
}

// Name impl.
func (q *SimpleMockQuery) Name() string {
	return q.name
}

// QueryRow impl.
func (q *SimpleMockQuery) QueryRow(args ...interface{}) (interface{}, error) {
	return q.returnValue, q.returnError
}
