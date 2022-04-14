package err2

import "io"

type _W struct{}

// W is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
var W _W

// Try is a helper method to call func() (io.Writer, error) functions
// with it and be as fast as Check(err).
func (o _W) Try(v io.Writer, err error) io.Writer {
	Check(err)
	return v
}
