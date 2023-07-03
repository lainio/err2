package main

import (
	"flag"

	"github.com/lainio/err2"
)

var (
	mode = flag.String("mode", "play", "runs the wanted playground: db, play,")
)

func main() {
	defer err2.Catch()
	
	flag.Parse()

	switch *mode {
	case "db":
		doDbMain()
	case "play":
		doPlayMain()
	default:
		err2.Throwf("unknown (%v) playground given", *mode)
	}
}
