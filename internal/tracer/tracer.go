// Package tracer implements thread safe storage for trace writers.
package tracer

import (
	"flag"
	"io"
	"os"
	"sync/atomic"

	"github.com/lainio/err2/internal/x"
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
	Log   value
)

func init() {
	Error.SetTracer(nil)
	// Because we stop panics as default, we need to output as default
	Panic.SetTracer(os.Stderr)

	// nil is a good default for try.Out().Logf() because then we use std log.
	Log.SetTracer(nil)

	flag.Var(&Log, "err2-log", "`stream` for logging: nil -> log pkg")
	flag.Var(
		&Error,
		"err2-trace",
		"`stream` for error tracing: stderr, stdout",
	)
	flag.Var(&Panic, "err2-panic-trace", "`stream` for panic tracing")
}

func (v *value) Tracer() io.Writer {
	if w, ok := v.Load().(writer); ok {
		return w.w
	}
	return nil
}

func (v *value) SetTracer(w io.Writer) {
	v.Store(writer{w: w})
}

// String is part of the flag interfaces
func (v *value) String() string {
	if v == nil {
		return "null"
	}
	return x.Whom(v.Tracer() != nil, "stderr", "nil")
}

// Get is part of the flag interfaces, getter.
func (v *value) Get() any {
	return v.Tracer()
}

// Set is part of the flag.Value interface.
func (v *value) Set(value string) error {
	switch value {
	case "stderr":
		v.SetTracer(os.Stderr)
	case "stdout":
		v.SetTracer(os.Stdout)
	case "nil":
		v.SetTracer(nil)
	}
	return nil
}
