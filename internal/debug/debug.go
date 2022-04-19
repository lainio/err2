package debug

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime/debug"
)

// PrintStack prints to standard error the stack trace returned by runtime.Stack
// by starting from stackLevel.
func PrintStack(stackLevel int) {
	FprintStack(os.Stderr, stackLevel)
}

// FprintStack prints to standard error the stack trace returned by runtime.Stack
// by starting from stackLevel.
func FprintStack(w io.Writer, stackLevel int) {
	stackBuf := bytes.NewBuffer(debug.Stack())
	scanner := bufio.NewScanner(stackBuf)

	// there is a caption line first, that's why we start from -1
	for i := -1; scanner.Scan(); i++ {
		if i == -1 || i/2 >= stackLevel {
			fmt.Fprintln(w, scanner.Text())
		}
	}
}
