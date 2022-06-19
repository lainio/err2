package err2

import "io"

type _R struct{}

// R is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
// Deprecated: use try package.
var R _R

// Try is a helper method to call func() (io.Reader, error) functions
// with it and be as fast as Check(err).
// Deprecated: use try package.
func (o _R) Try(v io.Reader, err error) io.Reader {
	Check(err)
	return v
}
