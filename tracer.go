package err2

import (
	"io"

	"github.com/lainio/err2/internal/tracer"
)

// StackTraceWriter allows to set automatic stack tracing.
//
//	err2.StackTraceWriter = os.Stderr // write stack trace to stderr
//	 or
//	err2.StackTraceWriter = log.Writer() // stack trace to std logger
//
// Deprecated: Use SetErrorTracer and SetPanicTracer to set tracers.
var StackTraceWriter io.Writer

// ErrorTracer returns current io.Writer for automatic error stack tracing.
func ErrorTracer() io.Writer {
	// Deprecated: until StackTraceWriter removed
	if StackTraceWriter != nil {
		return StackTraceWriter
	}
	return tracer.Error.Tracer()
}

// PanicTracer returns current io.Writer for automatic panic stack tracing. Note
// that runtime.Error types which are transported by panics are controlled by
// this.
func PanicTracer() io.Writer {
	// Deprecated: until StackTraceWriter removed
	if StackTraceWriter != nil {
		return StackTraceWriter
	}
	return tracer.Panic.Tracer()
}

// SetErrorTracer sets a io.Writer for automatic error stack tracing. Note
// that runtime.Error types which are transported by panics are controlled by
// this.
func SetErrorTracer(w io.Writer) {
	tracer.Error.SetTracer(w)
}

// SetPanicTracer sets a io.Writer for automatic panic stack tracing. Note
// that runtime.Error types which are transported by panics are controlled by
// this.
func SetPanicTracer(w io.Writer) {
	tracer.Panic.SetTracer(w)
}
