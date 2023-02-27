// Package main includes samples of err2. It works as a playground for users of
// the err2 package to test how different APIs work. We suggest you take your
// favorite editor and start to play with the main.go file. The comments in it
// guide you.
//
// Currently we have only one example build over CopyFile function, but with it
// you can try all the important APIs from err2, try, and assert.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
)

// CopyFile copies source file to the given destination. If any error occurs it
// returns error value describing the reason.
func CopyFile(src, dst string) (err error) {
	defer err2.Handle(&err) // automatic error message: see err2.Formatter

	// the next line is example how to enrich error message when starting with
	// automatic.
	//defer err2.Handle(&err, "(%v, %v)", src, dst)

	// You can comment above handler line(s) out an see what happens.
	// You'll learn that call stacks are for every function level 'catch'
	// statement like defer err2.Handle() is.

	assert.NotEmpty(src)
	assert.NotEmpty(dst)

	r := try.To1(os.Open(src))
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("mixing traditional error checking: %w", err)
	}
	defer err2.Handle(&err, func() {
		os.Remove(dst)
	})
	defer w.Close()
	try.To1(io.Copy(w, r))
	return nil
}

func callRecur(d int) (err error) {
	defer err2.Handle(&err)

	return recur(d)
}

func recur(d int) (err error) {
	d--
	if d >= 0 {
		fmt.Println(10 / d)
		return recur(d)
	}
	return fmt.Errorf("root error")
}

func main() {
	// To see how automatic stack tracing works.
	//err2.SetErrorTracer(os.Stderr)
	//err2.SetPanicTracer(os.Stderr) // this is the default

	// to see how there is two predefined formatters and own can be
	// implemented.
	//err2.SetFormatter(formatter.Noop) // default is formatter.Decamel

	// even no handlers is given, errors are caught without specific handlers.
	defer err2.Catch()

	// If you don't want to use tracers or you just need proper error handler
	// here.
	//	defer err2.Catch(func(err error) {
	//		fmt.Println("ERROR:", err)
	//	})

	try.To(doMain())

	println("______===")
}

func doMain() (err error) {
	defer err2.Handle(&err)

	// You can select anyone of the try.To(CopyFile lines to play with and see
	// how err2 works. Especially interesting is automatic stack tracing.
	//
	// source file exist, but destination not in high probability
	//try.To(callRecur(1))
	try.To(CopyFile("main.go", "/notfound/path/file.bak"))
	//
	// both source and destination doesn't exist
	//try.To(CopyFile("/notfound/path/file.go", "/notfound/path/file.bak"))
	//
	// first argument is empty
	//try.To(CopyFile("main.go", ""))

	println("=== you cannot see this ===")
	return nil
}
