package debug

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime/debug"
)

// PrintStack prints to standard error the stack trace returned by runtime.Stack
// by starting from stackLevel.
func PrintStack(stackLevel int) {
	stackBuf := bytes.NewBuffer(debug.Stack())
	scanner := bufio.NewScanner(stackBuf)

	for i := 0; scanner.Scan(); i++ {
		if i/2 > stackLevel {
			fmt.Fprintln(os.Stderr, scanner.Text())
		}
	}
}
