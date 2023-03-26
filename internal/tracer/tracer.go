// Package tracer implements thread safe storage for trace writers.
package tracer

import (
	"io"
	"os"
	"sync/atomic"
)

type value struct {
	atomic.Value
}

type writer struct {
	w io.Writer
}

var (
	Error value
	Panic value
)

func init() {
	Error.SetTracer(nil)
	// Because we stop panics as default, we need to output as default
	Panic.SetTracer(os.Stderr)
}

func (v *value) Tracer() io.Writer {
	return v.Load().(writer).w
}

func (v *value) SetTracer(w io.Writer) {
	v.Store(writer{w: w})
}
