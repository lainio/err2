package main

import (
	"flag"
	"log"
	"os"

	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
)

var (
	mode = flag.String(
		"mode",
		"play",
		"runs the wanted playground: db, play, nil, assert,"+
			"\nassert-keep (= uses assert.Debug in GLS)",
	)
	isErr = flag.Bool("err", false, "tells if we want to have an error")
)

func init() {
	// highlight that this is before flag.Parse to allow it to work properly.
	err2.SetLogTracer(os.Stderr) // for import
	err2.SetLogTracer(nil)
}

func main() {
	defer err2.Catch(err2.Stderr)
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	flag.Parse()

	switch *mode {
	case "db":
		doDBMain()
	case "nil":
		doMainAll()
	case "nil1":
		doMain1()
	case "nil2":
		doMain2()
	case "play":
		doPlayMain()
	case "assert":
		doAssertMainKeepGLSAsserter(false)
	case "assert-keep":
		doAssertMainKeepGLSAsserter(true)
	default:
		err2.Throwf("unknown (%v) playground given", *mode)
	}
}

func doAssertMainKeepGLSAsserter(keep bool) {
	asserterPusher(keep)
	asserterTester()
}

func asserterTester() {
	//defer assert.PushAsserter(assert.Development)()
	assert.That(false)
}

func asserterPusher(keep bool) {
	pop := assert.PushAsserter(assert.Debug)
	if !keep { // if not keep we free
		pop()
	}
}
