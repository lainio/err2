package err2

import "net/url"

type _URL struct{}

// URL is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
var URL _URL

// Try is a helper method to call func() (*url.URL, error) functions
// with it and be as fast as Check(err).
func (o _URL) Try(v *url.URL, err error) *url.URL {
	Check(err)
	return v
}
