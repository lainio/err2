package err2

import "net/http"

type _Response struct{}

// Response is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
// Deprecated: use try package.
var Response _Response

// Try is a helper method to call func() (*http.Response, error) functions
// with it and be as fast as Check(err).
// Deprecated: use try package.
func (o _Response) Try(v *http.Response, err error) *http.Response {
	Check(err)
	return v
}
