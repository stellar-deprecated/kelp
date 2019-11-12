package networking

// HeaderFn represents a function that transforms headers
type HeaderFn func(string, string, string) string // (string httpMethod, string requestPath, string body)

// MakeStaticHeaderFn is a convenience method
func MakeStaticHeaderFn(value string) HeaderFn {
	// need to convert to HeaderFn to work as a api.ExchangeHeader.Value
	return HeaderFn(func(method string, requestPath string, body string) string {
		return value
	})
}
