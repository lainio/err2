//go:build !windows

package err2_test

import (
	"fmt"
	"io"
	"os"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func CopyFile(src, dst string) (err error) {
	// Automatic error annotation from current function name.
	defer err2.Handle(&err)

	// NOTE. These try.To() checkers are as fast as `if err != nil {}`
	r := try.To1(os.Open(src))
	defer r.Close() // deferred resource cleanup is perfect match with err2

	w := try.To1(os.Create(dst))
	defer err2.Handle(&err, func() {
		// If error happens during Copy we clean not completed file here
		// Look how well it suits with other cleanups like Close calls.
		os.Remove(dst)
	})
	defer w.Close()
	try.To1(io.Copy(w, r))
	return nil
}

func Example() {
	// To see how automatic stack tracing works please run this example with:
	//   go test -v -run='^Example$'
	err2.SetErrorTracer(os.Stderr)

	err := CopyFile("/notfound/path/file.go", "/notfound/path/file.bak")
	if err != nil {
		fmt.Println(err)
	}
	// in real word example 'run example' is 'copy file' it comes automatically
	// from function name that calls `err2.Handle` in deferred.

	// Output: testing: run example: open /notfound/path/file.go: no such file or directory
}
