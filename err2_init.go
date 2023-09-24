package err2

import (
	"flag"

	"github.com/lainio/err2/internal/tracer"
)

func init() {
	flag.Var(&tracer.Log, "err2-log", "stream for logging")
	flag.Var(&tracer.Error, "err2-trace", "stream for error tracing")
	flag.Var(&tracer.Panic, "err2-panic-trace", "stream for panic tracing")
}
