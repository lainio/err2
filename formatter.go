package err2

import (
	"github.com/lainio/err2/formatter"
	fmtstore "github.com/lainio/err2/internal/formatter"
)

func init() {
	SetFormatter(formatter.Decamel)
}

// SetFormatter sets the current formatter for the err2 package. The default
// formatter.Decamel tries to process function names to human readable and the
// idiomatic Go format, i.e. all lowercase, space delimiter, etc.
//
// Following line sets a noop formatter where errors are taken as function names
// are in the call stack.
//
//	err2.SetFormatter(formatter.Noop)
//
// You can make your own implementations of formatters. See more information
// in formatter package.
func SetFormatter(f formatter.Interface) {
	fmtstore.SetFormatter(f)
}

// Returns the current formatter. See more information from [SetFormatter] and
// [formatter] package.
func Formatter() formatter.Interface {
	return fmtstore.Formatter()
}
