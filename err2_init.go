package err2

import (
	"flag"

	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/internal/tracer"
)

var (
	// asserter def index flag has following values:
	//
	//	Production
	//	Development
	//	Test
	//	TestFull
	//	Debug
	asserterFlag flagAsserter
)

func init() {
	flag.Var(&tracer.Log, "err2-log", "stream for logging")
	flag.Var(&tracer.Error, "err2-trace", "stream for error tracing")
	flag.Var(&tracer.Panic, "err2-panic-trace", "stream for panic tracing")
	flag.Var(&asserterFlag, "err2-asserter", "asserter: Production, Development, Debug")
}

type flagAsserter struct {
	v string
}

// String is part of the flag interfaces
func (f *flagAsserter) String() string {
	return f.v
}

// Get is part of the flag interfaces, getter.
func (*flagAsserter) Get() any { return nil }

// Set is part of the flag.Value interface.
func (f *flagAsserter) Set(flagAsserter string) error {
	assert.SetDefault(assert.NewDefInd(flagAsserter))
	return nil
}
