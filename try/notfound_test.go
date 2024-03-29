package try_test

import (
	"fmt"
	"os"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func FindObject(_ int) (val string, err error) {
	defer err2.Handle(&err)

	// both of the following lines can be used to transport err2.NotFound
	// you can try by outcommenting err2.Throwf
	//err2.Throwf("panic transport: %w", err2.ErrNotFound)
	return "", err2.ErrNotFound
}

func ExampleIsNotFound1() {
	// To see how automatic stack tracing works WITH panic transport please run
	// this example with:
	//   go test -v -run='^ExampleNotFound$'

	// pick up your poison: outcomment the nil line to see how error tracing
	// works.
	err2.SetErrorTracer(os.Stderr)
	err2.SetErrorTracer(nil)

	find := func(key int) string {
		defer err2.Catch(err2.Err(func(err error) {
			fmt.Println("ERROR:", err)
		}))
		notFound, value := try.IsNotFound1(FindObject(key))
		if notFound {
			return fmt.Sprintf("cannot find key (%d)", key)
		}
		return "value for key is:" + value
	}

	fmt.Println(find(1))
	// Output: cannot find key (1)
}
