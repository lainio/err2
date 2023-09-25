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
	flag.Var(&tracer.Log, "err2-log", "stream for logging: nil -> log pkg")
	flag.Var(&tracer.Error, "err2-trace", "stream for error tracing: stderr, stdout")
	flag.Var(&tracer.Panic, "err2-panic-trace", "stream for panic tracing")
	flag.Var(&asserterFlag, "err2-asserter", "asserter: Production, Development, Debug")
}

type flagAsserter struct{}

// String is part of the flag interfaces
func (*flagAsserter) String() string {
	return assert.AsserterString()
}

// Get is part of the flag interfaces, getter.
func (*flagAsserter) Get() any { return nil }

// Set is part of the flag.Value interface.
func (*flagAsserter) Set(value string) error {
	assert.SetDefault(assert.NewDefInd(value))
	return nil
}
