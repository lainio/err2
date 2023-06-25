package try_test

import (
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
			Handle(func(err error) error {
				w.Close()
				os.Remove(dst)
				return err
			}).Val1

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

func ExampleResult1_Logf() {
	// Set log tracing to stdout that we can see it in Example output. In
	// normal cases that would be a Logging stream or stderr.
	err2.SetLogTracer(os.Stdout)

	countSomething := func(s string) int {
		return try.Out1(strconv.Atoi(s)).Logf("not number").Def1(100).Val1
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
		r := try.Out2(convTwoStr(s1, s2)).Logf("wrong number").Def2(1, 2)
		v1, v2 := r.Val1, r.Val2
		return v1 + v2, v2
	}
	num1, num2 := countSomething("1", "err")
	fmt.Printf("results: %d, %d\n", num1, num2)
	test.Requiref(t, num1 == 3, "wrong number: got %d, want: 3", num1)
	test.Requiref(t, num2 == 2, "wrong number: got %d, want: 2", num2)
}
