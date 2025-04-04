//go:build !windows

package try_test

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

var (
	errForTesting = fmt.Errorf("error for %s", "testing")
)

func ExampleIs_errorHappens() {
	copyStream := func(src string) (s string, err error) {
		defer err2.Handle(&err, "copy stream (%s)", src)

		err = errForTesting
		try.Is(err, io.EOF)
		return src, nil
	}

	str, err := copyStream("testing string")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str)
	// Output: copy stream (testing string): error for testing
}

func ExampleIs_errorHappensNot() {
	copyStream := func(src string) (s string, err error) {
		defer err2.Handle(&err, "copy stream %s", src)

		err = fmt.Errorf("something: %w", errForTesting)
		if try.Is(err, errForTesting) {
			return "wrapping works", nil
		}

		return src, nil
	}

	str, err := copyStream("testing string")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str)
	// Output: wrapping works
}

//nolint:unparam
func ExampleOut_errorHappensNot() {
	var is bool
	var errFn = func(error) error {
		is = true
		return nil
	}
	copyStream := func(src string) (s string, err error) {
		defer err2.Handle(&err, "copy stream %s", src)

		err = fmt.Errorf("something: %w", errForTesting)

		try.Out(err).Handle(errForTesting, errFn)
		if is {
			return "wrapping works", nil
		}

		return src, nil
	}

	str, err := copyStream("testing string")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str)
	// Output: wrapping works
}

func ExampleIsEOF1() {
	copyStream := func(src string) (s string, err error) {
		defer err2.Handle(&err)

		in := bytes.NewBufferString(src)
		tmp := make([]byte, 4)
		var out bytes.Buffer
		for eof, n := try.IsEOF1(in.Read(tmp)); !eof; eof, n = try.IsEOF1(in.Read(tmp)) {
			out.Write(tmp[:n])
		}

		return out.String(), nil
	}

	str, err := copyStream("testing string")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str)
	// Output: testing string
}

func Example_copyFile() {
	copyFile := func(src, dst string) (err error) {
		defer err2.Handle(&err, "copy")

		// These try package helpers are as fast as Check() calls which is as
		// fast as `if err != nil {}`

		r := try.T1(os.Open(src))("source file")
		defer r.Close()

		w := try.T1(os.Create(dst))("target file")
		defer err2.Handle(&err, err2.Err(func(error) {
			os.Remove(dst)
		}))
		defer w.Close()
		try.To1(io.Copy(w, r))
		return nil
	}

	err := copyFile("/notfound/path/file.go", "/notfound/path/file.bak")
	if err != nil {
		fmt.Println(err)
	}
	// Output: copy: source file: open /notfound/path/file.go: no such file or directory
}
