// Package formatter implements formatters and helper types for err2. See more
// information from err2.SetFormatter.
package formatter

import (
	"github.com/lainio/err2/internal/str"
)

// Interface is a formatter interface. The implementers are used for automatic
// error message generation from function names. See more information from
// err2.Handle.
type Interface interface {
	Format(input string) string
}

// DoFmt is a helper function type which allows reuse Formatter struct for the
// implementations.
type DoFmt func(i string) string

// Formatter is a helper struct which wraps the actual formatting function which
// is called during the function name processing to produce errors
// automatically.
type Formatter struct {
	DoFmt
}

var (
	// Decamel is preimplemented and default formatter to produce human
	// readable error strings from function names.
	//   func CopyFile(..)  -> "copy file: file not exists"
	//                          ^-------^ -> generated from CopyFile
	Decamel = &Formatter{DoFmt: str.Decamel}

	// Noop is preimplemented formatter that does nothing to function name.
	//   func CopyFile(..)  -> "CopyFile: file not exists"
	//                          ^------^ -> function name as it is: CopyFile
	Noop = &Formatter{DoFmt: func(i string) string { return i }}
)

// Format just calls function set in the DoFmt field.
func (f *Formatter) Format(input string) string {
	return f.DoFmt(input)
}
