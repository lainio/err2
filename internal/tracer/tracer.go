package tracer

import (
	"io"
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
	Panic.SetTracer(nil)
}

func (v *value) Tracer() io.Writer {
	return v.Load().(writer).w
}

func (v *value) SetTracer(w io.Writer) {
	v.Store(writer{w: w})
}
