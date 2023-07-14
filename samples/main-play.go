// Package main includes samples of err2. It works as a playground for users of
// the err2 package to test how different APIs work. We suggest you take your
// favorite editor and start to play with the main.go file. The comments on it
// guide you.
//
// We have only a few examples built over the CopyFile and callRecur functions,
// but with them you can try all the important APIs from err2, try, and assert.
// Just follow the comments and try suggested things :-)
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
)

// CopyFile copies the source file to the given destination. If any error occurs it
// returns an error value describing the reason.
func CopyFile(src, dst string) (err error) {
	defer err2.Handle(&err) // automatic error message: see err2.Formatter
	// You can out-comment above handler line(s) to see what happens.

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

	return doRecur(d)
}

func doRecur(d int) (err error) {
	d--
	if d >= 0 {
		// Keep below to show how asserts work
		assert.NotZero(d)
		// Comment out the above assert statement to simulate runtime-error
		fmt.Println(10 / d)
		return doRecur(d)
	}
	return fmt.Errorf("root error")
}

func doPlayMain() {
	// Keep here that you can play without changing imports
	assert.That(true)

	// If asserts are treated as panics instead of errors, you get the stack trace.
	// you can try that by taking the next line out of the comment:
	assert.SetDefault(assert.Development)

	// same thing but one line assert msg
	//assert.SetDefault(assert.Production)

	// To see how automatic stack tracing works.
	//err2.SetErrorTracer(os.Stderr)

	//err2.SetPanicTracer(os.Stderr) // this is the default

	// to see how there are two predefined formatters and own can be
	// implemented.
	//err2.SetFormatter(formatter.Noop) // default is formatter.Decamel

	// errors are caught without specific handlers.
	defer err2.Catch("CATCH")

	// If you don't want to use tracers or you just need a proper error handler
	// here.
	//	defer err2.Catch(func(err error) {
	//		fmt.Println("ERROR:", err)
	//	})

	// by calling one of these you can test how automatic logging in above
	// catch works correctly: the last source of error check is shown in line
	// count
	doDoMain()
	//try.To(doMain())

	println("______===")
}

func doDoMain() {
	try.To(doMain())
}

func doMain() (err error) {
	defer err2.Handle(&err)

	// You can select any one of the try.To(CopyFile lines to play with and see
	// how err2 works. Especially interesting is automatic stack tracing.
	//
	// source file exists, but the destination is not in high probability
	//try.To(CopyFile("main.go", "/notfound/path/file.bak"))

	// Both source and destination don't exist
	//try.To(CopyFile("/notfound/path/file.go", "/notfound/path/file.bak"))

	// 2nd argument is empty
	//try.To(CopyFile("main.go", ""))

	// Next fn demonstrates how error and panic traces work, comment out all
	// above CopyFile calls to play with:
	try.To(callRecur(1))

	println("=== you cannot see this ===")
	return nil
}
