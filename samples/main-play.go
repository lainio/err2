// Package main includes samples of err2. It works as a playground for users of
// the err2 package to test how different APIs work. We suggest you take your
// favorite editor and start to play with the main.go file. The comments on it
// guide you.
//
// We have only a few examples built over the [CopyFile] and [CallRecur] functions,
// but with them you can try all the important APIs from err2, try, and assert.
// Just follow the comments and try suggested things :-)
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
)

// CopyFile copies the source file to the given destination. If any error occurs it
// returns an error value describing the reason.
func CopyFile(src, dst string) (err error) {
	defer err2.Handle(&err)

	r := try.To1(os.Open(src))
	defer r.Close()

	w := try.To1(os.Create(dst))
	defer err2.Handle(&err, func(err error) error {
		try.Out(os.Remove(dst)).Logf()
		return err
	})
	defer w.Close()

	try.To1(io.Copy(w, r))
	return nil
}

func ClassicCopyFile(src, dst string) error {
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("copy %s %s: %v", src, dst, err)
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("copy %s %s: %v", src, dst, err)
	}

	if _, err := io.Copy(w, r); err != nil {
		w.Close()
		os.Remove(dst)
		return fmt.Errorf("copy %s %s: %v", src, dst, err)
	}

	if err := w.Close(); err != nil {
		os.Remove(dst)
		return fmt.Errorf("copy %s %s: %v", src, dst, err)
	}
	return nil
}

// TryCopyFile copies the source file to the given destination. If any error occurs it
// returns an error value describing the reason.
func TryCopyFile(src, dst string) {
	// You can out-comment above handler line(s) to see what happens.

	// You'll learn that call stacks are for every function level 'catch'
	// statement like defer err2.Handle() is.

	assert.NotEmpty(src)
	assert.NotEmpty(dst)

	r := try.To1(os.Open(src))
	defer r.Close()

	w:= try.To1(os.Create(dst))
	defer err2.Handle(&err, func(err error) error {
		try.Out(os.Remove(dst)).Logf("cleaning error")
		return err
	})
	defer w.Close()
	try.To1(io.Copy(w, r))
}

func CallRecur(d int) (ret int, err error) {
	defer err2.Handle(&err)

	return doRecur(d)
}

func doRecur(d int) (ret int, err error) {
	d--
	if d >= 0 {
		// Keep below to show how asserts work
		//assert.NotZero(d)
		// Comment out the above assert statement to simulate runtime-error
		ret = 10 / d
		fmt.Println(ret)
		//return doRecur(d)
	}
	return ret, fmt.Errorf("root error")
}

func doPlayMain() {
	// Keep here that you can play without changing imports
	assert.That(true)

	// To see how automatic stack tracing works.
	//err2.SetErrorTracer(os.Stderr)

	//err2.SetPanicTracer(os.Stderr) // this is the default

	// to see how there are two predefined formatters and own can be
	// implemented.
	//err2.SetFormatter(formatter.Noop) // default is formatter.Decamel

	// errors are caught without specific handlers.
	defer err2.Catch(err2.Stderr)

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

	fmt.Println("___ happy ending ===")
}

func doDoMain() {
	try.To(doMain())
}

func doMain() (err error) {
	// Example of Handle/Catch API where we can have multiple handlers.
	// Note that this is a silly sample where logging is done trice and noops
	// are used without a purpose. All of this is that you get an idea how you
	// could use the error handlers and chain them together.

	//defer err2.Handle(&err, err2.Noop, err2.Log, err2.Log)
	//defer err2.Handle(&err, nil, err2.Noop, err2.Log)
	//defer err2.Handle(&err, nil, err2.Log)
	defer err2.Handle(&err)

	// You can select any one of the try.To(CopyFile lines to play with and see
	// how err2 works. Especially interesting is automatic stack tracing.
	//
	// source file exists, but the destination is not in high probability
	//TryCopyFile("main.go", "/notfound/path/file.bak")

	// Both source and destination don't exist
	//TryCopyFile("/notfound/path/file.go", "/notfound/path/file.bak")

	// to play with real args:
	TryCopyFile(flag.Arg(0), flag.Arg(1))

	if len(flag.Args()) > 0 {
		// Next fn demonstrates how error and panic traces work, comment out all
		// above CopyFile calls to play with:
		argument := try.To1(strconv.Atoi(flag.Arg(0)))
		ret := try.To1(CallRecur(argument))
		fmt.Println("ret val:", ret)
	} else {
		// 2nd argument is empty to assert
		TryCopyFile("main.go", "")
	}

	fmt.Println("=== you cannot see this ===")
	return nil
}
