package err2

import (
	"io"

	"github.com/lainio/err2/internal/tracer"
)

// ErrorTracer returns current io.Writer for automatic error stack tracing.
func ErrorTracer() io.Writer {
	return tracer.Error.Tracer()
}

// PanicTracer returns current io.Writer for automatic panic stack tracing. Note
// that runtime.Error types which are transported by panics are controlled by
// this.
func PanicTracer() io.Writer {
	return tracer.Panic.Tracer()
}

// SetErrorTracer sets a io.Writer for automatic error stack tracing. Note
// that runtime.Error types which are transported by panics are controlled by
// this. Note also that the current function is capable to print error stack
// trace when the function has at least one deferred error handler, e.g:
//
//	func CopyFile(src, dst string) (err error) {
//	     defer err2.Handle(&err) // <- error trace print decision is done here
func SetErrorTracer(w io.Writer) {
	tracer.Error.SetTracer(w)
}

// SetPanicTracer sets a io.Writer for automatic panic stack tracing. Note
// that runtime.Error types which are transported by panics are controlled by
// this. Note also that the current function is capable to print panic stack
// trace when the function has at least one deferred error handler, e.g:
//
//	func CopyFile(src, dst string) (err error) {
//	     defer err2.Handle(&err) // <- error trace print decision is done here
func SetPanicTracer(w io.Writer) {
	tracer.Panic.SetTracer(w)
}

// SetTracers a convenient helper to set a io.Writer for error and panic stack
// tracing. More information see SetErrorTracer and SetPanicTracer functions.
func SetTracers(w io.Writer) {
	tracer.Error.SetTracer(w)
	tracer.Panic.SetTracer(w)
}
