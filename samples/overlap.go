package main

import (
	"fmt"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func doOverlapMain() {
	err := FirstHandle()
	fmt.Printf("Final error: %v\n", err)
}

func FirstHandle() (err error) {
	defer err2.Handle(&err)
	try.To(SecondHandle())
	return nil
}

func SecondHandle() (err error) {
	defer err2.Handle(&err)
	try.T(fmt.Errorf("my error"))("my call lvl annotation")
	return nil
}
