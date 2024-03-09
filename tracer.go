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

// LogTracer returns a current io.Writer for the explicit try.Result.Logf
// function and automatic logging used in err2.Handle and err2.Catch. The
// default value is nil.
func LogTracer() io.Writer {
	return tracer.Log.Tracer()
}

// SetErrorTracer sets a io.Writer for automatic error stack tracing. The err2
// default is nil. Note that the current function is capable to print error
// stack trace when the function has at least one deferred error handler, e.g:
//
//	func CopyFile(src, dst string) (err error) {
//	     defer err2.Handle(&err) // <- error trace print decision is done here
//
// Remember that you can reset these with Flag package support. See
// documentation of err2 package's flag section.
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
//	     defer err2.Handle(&err) // <- panic trace print decision is done here
//
// Remember that you can reset these with Flag package support. See
// documentation of err2 package's flag section.
func SetPanicTracer(w io.Writer) {
	tracer.Panic.SetTracer(w)
}

// SetLogTracer sets a current io.Writer for the explicit try.Result.Logf
// function and automatic logging used in err2.Handle and err2.Catch. The
// default is nil and then err2 uses std log package for logging.
//
// You can use the std log package to redirect other logging packages like glog
// to automatically work with the err2 package. For the glog, add this line at
// the beginning of your app:
//
//	glog.CopyStandardLogTo("INFO")
//
// Remember that you can reset these with Flag package support. See
// documentation of err2 package's flag section.
func SetLogTracer(w io.Writer) {
	tracer.Log.SetTracer(w)
}

// SetTracers a helper to set a io.Writer for error and panic stack tracing, the
// log tracer is set as well. More information see SetErrorTracer,
// SetPanicTracer, and SetLogTracer functions.
//
// Remember that you can reset these with Flag package support. See
// documentation of err2 package's flag section.
func SetTracers(w io.Writer) {
	tracer.Error.SetTracer(w)
	tracer.Panic.SetTracer(w)
	tracer.Log.SetTracer(w)
}
