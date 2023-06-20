package try_test

import (
	"fmt"
	"io"
	"os"

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
		_ = try.Out1(io.Copy(w, r)).Handle(func(err error) error {
			w.Close()
			os.Remove(dst)
			return err
		}).Val1 // and if success, you can use the Copy's retval

		w.Close()
		return nil
	}

	err := copyFile("/notfound/path/file.go", "/notfound/path/file.bak")
	if err != nil {
		fmt.Println(err)
	}
	// Output: copy /notfound/path/file.go /notfound/path/file.bak: open /notfound/path/file.go: no such file or directory
}
