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
	// You can comment below line out an see what happens...
	defer err2.Returnf(&err, "copy file %s->%s", src, dst)
	// ... and you learn that call stacks are for every function level 'catch'
	// statement like defer err2.Returnf() is.

	assert.NotEmpty(src)
	assert.NotEmpty(dst)

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

func main() {
	// To see how automatic stack tracing works.
	err2.SetErrorTracer(os.Stderr)

	defer err2.Catch(func(err error) {
		fmt.Println("ERROR in copy file:", err)
	})

	// You can select anyone of the try.To(CopyFile lines to play with and see
	// how err2 works. Especially interesting is automatic stack tracing.
	//
	// source file exist, but destination not in high probability
	try.To(CopyFile("main.go", "/notfound/path/file.bak"))
	//
	// both source and destination doesn't exist
	//try.To(CopyFile("/notfound/path/file.go", "/notfound/path/file.bak"))
	//
	// first argument is empty
	//try.To(CopyFile("main.go", ""))
}
