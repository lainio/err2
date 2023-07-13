package main

import (
	"flag"
	"log"

	"github.com/lainio/err2"
)

var (
	mode = flag.String("mode", "play", "runs the wanted playground: db, play,")
)

func main() {
	defer err2.Catch()
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	flag.Parse()

	switch *mode {
	case "db":
		doDBMain()
	case "play":
		doPlayMain()
	default:
		err2.Throwf("unknown (%v) playground given", *mode)
	}
}
