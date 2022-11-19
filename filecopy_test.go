package err2_test

import (
	"fmt"
	"io"
	"os"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func CopyFile(src, dst string) (err error) {
	defer err2.Handle(&err, "copy file %s->%s", src, dst)

	// These try.To() checkers are as fast as `if err != nil {}`

	r := try.To1(os.Open(src))
	defer r.Close()

	w := try.To1(os.Create(dst))
	defer err2.Handle(&err, func() {
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
	// Output: copy file /notfound/path/file.go->/notfound/path/file.bak: open /notfound/path/file.go: no such file or directory
}
