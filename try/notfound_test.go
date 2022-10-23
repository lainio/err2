package try_test

import (
	"fmt"
	"os"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func FindObject(key int) (val string, err error) {
	//defer err2.Return(&err)
	defer err2.Returnw(&err, "")

	// both of the following lines can be used to transport err2.NotFound
	// you can try by outcommenting err2.Throwf
	//err2.Throwf("panic transport: %w", err2.NotFound)
	return "", err2.NotFound
}

func ExampleIsNotFound1() {
	// To see how automatic stack tracing works WITH panic transport please run
	// this example with:
	//   go test -v -run='^ExampleNotFound$'
	err2.SetErrorTracer(os.Stderr)
	err2.SetErrorTracer(nil)
	find := func(key int) string {
		defer err2.Catch(func(err error) {
			fmt.Println("ERROR:", err)
		})
		found, value := try.IsNotFound1(FindObject(key))
		if found {
			return fmt.Sprintf("cannot find key (%d)", key)
		}
		return "value for key is:" + value
	}

	fmt.Println(find(1))
	// Output: cannot find key (1)
}
