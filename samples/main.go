package main

import (
	"flag"
	"log"

	"github.com/lainio/err2"
)

var (
	mode  = flag.String("mode", "play", "runs the wanted playground: db, play, nil")
	isErr = flag.Bool("err", false, "tells if we have error")
)

func main() {
	defer err2.Catch()
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
	default:
		err2.Throwf("unknown (%v) playground given", *mode)
	}
}
