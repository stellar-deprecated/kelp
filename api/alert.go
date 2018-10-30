package api

// Alert interface is used for the various monitoring and alerting tools for Kelp.
type Alert interface {
	Trigger(description string, details interface{}) error
}
