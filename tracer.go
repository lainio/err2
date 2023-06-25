package err2

import (
	"io"

	"github.com/lainio/err2/internal/tracer"
)

// ErrorTracer returns current io.Writer for automatic error stack tracing.
// The default value is nil.
func ErrorTracer() io.Writer {
	return tracer.Error.Tracer()
}

// PanicTracer returns current io.Writer for automatic panic stack tracing. Note
// that runtime.Error types which are transported by panics are controlled by
// this. The default value is os.Stderr.
func PanicTracer() io.Writer {
	return tracer.Panic.Tracer()
}

// LogTracer returns current io.Writer for try.Out().Logf().
// The default value is nil.
func LogTracer() io.Writer {
	return tracer.Log.Tracer()
}

// SetErrorTracer sets a io.Writer for automatic error stack tracing. The err2
// default is nil. Note that the current function is capable to print error
// stack trace when the function has at least one deferred error handler, e.g:
//
//	func CopyFile(src, dst string) (err error) {
//	     defer err2.Handle(&err) // <- error trace print decision is done here
func SetErrorTracer(w io.Writer) {
	tracer.Error.SetTracer(w)
}

// SetPanicTracer sets a io.Writer for automatic panic stack tracing. The err2
// default is os.Stderr. Note that runtime.Error types which are transported by
// panics are controlled by this. Note also that the current function is capable
// to print panic stack trace when the function has at least one deferred error
// handler, e.g:
//
//	func CopyFile(src, dst string) (err error) {
//	     defer err2.Handle(&err) // <- error trace print decision is done here
func SetPanicTracer(w io.Writer) {
	tracer.Panic.SetTracer(w)
}

// SetLogTracer sets a io.Writer for try.Out().Logf() function.
// The default is nil.
func SetLogTracer(w io.Writer) {
	tracer.Log.SetTracer(w)
}

// SetTracers a convenient helper to set a io.Writer for error and panic stack
// tracing. More information see SetErrorTracer and SetPanicTracer functions.
func SetTracers(w io.Writer) {
	tracer.Error.SetTracer(w)
	tracer.Panic.SetTracer(w)
	tracer.Log.SetTracer(w)
}
