package try_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/lainio/err2"
	"github.com/lainio/err2/internal/test"
	"github.com/lainio/err2/try"
)

func ExampleOut1_copyFile() {
	copyFile := func(src, dst string) (err error) {
		defer err2.Handle(&err, "copy %s %s", src, dst)

		r := try.To1(os.Open(src))
		defer r.Close()

		w := try.To1(os.Create(dst))

		// If you prefer immediate error handling for some reason.
		_ = try.Out1(io.Copy(w, r)).
			Handle(io.EOF, func(err error) error {
				fmt.Println("err == io.EOF")
				return nil // by returning nil we can reset the error
				// return err // fallthru to next check if err != nil
			}).
			Handle(func(err error) error {
				try.Out(w.Close()).Logf()
				try.Out(os.Remove(dst)).Logf()
				return err
			}).
			Val1

		w.Close()
		return nil
	}

	err := copyFile("/notfound/path/file.go", "/notfound/path/file.bak")
	if err != nil {
		fmt.Println(err)
	}

	// Output: copy /notfound/path/file.go /notfound/path/file.bak: open /notfound/path/file.go: no such file or directory
}

func ExampleResult1_Def1() {
	countSomething := func(s string) int {
		return try.Out1(strconv.Atoi(s)).Def1(100).Val1
	}
	num1 := countSomething("1")
	num2 := countSomething("not number, getting default (=100)")
	fmt.Printf("results: %d, %d", num1, num2)

	// Output: results: 1, 100
}

func ExampleResult1_Catch1() {
	countSomething := func(s string) int {
		return try.Out1(strconv.Atoi(s)).Catch1(100)
	}
	num1 := countSomething("1")
	num2 := countSomething("not number, getting default (=100)")
	fmt.Printf("results: %d, %d", num1, num2)

	// Output: results: 1, 100
}

func ExampleResult1_Logf() {
	// Set log tracing to stdout that we can see it in Example output. In
	// normal cases that would be a Logging stream or stderr.
	err2.SetLogTracer(os.Stdout)

	countSomething := func(s string) int {
		return try.Out1(strconv.Atoi(s)).Logf("not number").Catch1(100)
	}
	num1 := countSomething("1")
	num2 := countSomething("WRONG")
	fmt.Printf("results: %d, %d", num1, num2)
	err2.SetLogTracer(nil)

	// Output: not number: strconv.Atoi: parsing "WRONG": invalid syntax
	// results: 1, 100
}

func TestResult2_Logf(t *testing.T) {
	// Set log tracing to stdout that we can see it in Example output. In
	// normal cases that would be a Logging stream or stderr.
	err2.SetLogTracer(os.Stdout)

	convTwoStr := func(s1, s2 string) (_ int, _ int, err error) {
		defer err2.Handle(&err, nil)

		return try.To1(strconv.Atoi(s1)), try.To1(strconv.Atoi(s2)), nil
	}
	countSomething := func(s1, s2 string) (int, int) {
		v1, v2 := try.Out2(convTwoStr(s1, s2)).Logf("wrong number").Catch2(1, 2)
		return v1 + v2, v2
	}
	num1, num2 := countSomething("1", "err")
	fmt.Printf("results: %d, %d\n", num1, num2)
	test.RequireEqual(t, num1, 3)
	test.RequireEqual(t, num2, 2)
}

func ExampleResult1_Handle() {
	// try out f() |err| handle to show power of error handling language, EHL
	callRead := func(in io.Reader, b []byte) (eof bool, n int) {
		// we should use try.To1, but this is sample of try.Out.Handle
		n = try.Out1(in.Read(b)).
			Handle(io.EOF, func(err error) error {
				eof = true
				return nil
			}).       // our errors.Is == true, handler to get eof status
			Handle(). // rest of the errors just throw
			Val1      // get count of read bytes, 1st retval of io.Read
		return
	}
	// simple function to copy stream with io.Reader
	copyStream := func(src string) (s string, err error) {
		defer err2.Handle(&err)

		in := bytes.NewBufferString(src)
		tmp := make([]byte, 4)
		var out bytes.Buffer

		for eof, n := callRead(in, tmp); !eof; eof, n = callRead(in, tmp) {
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
