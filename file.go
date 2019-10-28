package err2

import "os"

type _File struct{}

// File is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
var File _File

// Try is a helper method to call func() (*os.File, error) functions
// with it and be as fast as Check(err).
func (o _File) Try(v *os.File, err error) *os.File {
	Check(err)
	return v
}
