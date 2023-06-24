package try_test

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func ExampleOut1_copyFile() {
	copyFile := func(src, dst string) (err error) {
		defer err2.Handle(&err, "copy %s %s", src, dst)

		r := try.To1(os.Open(src))
		defer r.Close()

		w := try.To1(os.Create(dst))

		// If you prefer immediate error handling for some reason.
		try.Out1(io.Copy(w, r)).
			Handle(func(err error) error {
				w.Close()
				os.Remove(dst)
				return err
			})

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
	// Set error tracing to stdout that we can see it in Example output. In
	// normal cases that would be a Logging stream or stderr.
	err2.SetTracers(os.Stdout)

	countSomething := func(s string) int {
		return try.Out1(strconv.Atoi(s)).Logf("not number").Def1(100).Val1
	}
	num1 := countSomething("1")
	num2 := countSomething("WRONG")
	fmt.Printf("results: %d, %d", num1, num2)
	// not number: strconv.Atoi: parsing "WRONG": invalid syntax Output: results: 1, 100
}
