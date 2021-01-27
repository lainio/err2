package err2

import "net/http"

type _Request struct{}

// Request is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
var Request _Request

// Try is a helper method to call func() (*http.Request, error) functions
// with it and be as fast as Check(err).
func (o _Request) Try(v *http.Request, err error) *http.Request {
	Check(err)
	return v
}
