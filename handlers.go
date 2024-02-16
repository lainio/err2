package err2

import (
	"fmt"
	"os"

	"github.com/lainio/err2/internal/handler"
)

// Stderr is a built-in helper to use with Handle and Catch. It prints the
// error to stderr and it resets the current error value. It's a handy Catch
// handler in main function.
//
// You can use it like this:
//
//	func main() {
//		defer err2.Catch(err2.Stderr)
func Stderr(err error) error {
	if err == nil {
		return nil
	}
	fmt.Fprintln(os.Stderr, err.Error())
	return nil
}

// Stdout is a built-in helper to use with Handle and Catch. It prints the
// error to stdout and it resets the current error value. It's a handy Catch
// handler in main function.
//
// You can use it like this:
//
//	func main() {
//		defer err2.Catch(err2.Stdout)
func Stdout(err error) error {
	if err == nil {
		return nil
	}
	fmt.Fprintln(os.Stdout, err.Error())
	return nil
}

// Noop is a built-in helper to use with Handle and Catch. It keeps the current
// error value the same. You can use it like this:
//
//	defer err2.Handle(&err, err2.Noop)
func Noop(err error) error { return err }

// Reset is a built-in helper to use with Handle and Catch. It sets the current
// error value to nil. You can use it like this to reset the error:
//
//	defer err2.Handle(&err, err2.Reset)
func Reset(error) error { return nil }

// Err is a built-in helper to use with Handle and Catch. It offers simplifier
// for error handling function for cases where you don't need to change the
// current error value. For instance, if you want to just write error to stdout,
// and don't want to use SetLogTracer and keep it to write to your logs.
//
//	defer err2.Catch(err2.Err(func(err error) {
//		fmt.Println("ERROR:", err)
//	}))
//
// Note, that since Err helper we have other helpers like Stdout that allows
// previous block be written as simple as:
//
//	defer err2.Catch(err2.Stdout)
func Err(f func(err error)) Handler {
	return func(err error) error {
		f(err)
		return err
	}
}

const lvl = 10

// Log prints error string to the current log that is set by SetLogTracer.
func Log(err error) error {
	if err == nil {
		return nil
	}
	_ = handler.LogOutput(lvl, err.Error())
	return err
}
