package err2

import (
	"flag"

	"github.com/lainio/err2/assert"
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
