package debug

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
)

type StackInfo struct {
	PackageName string
	FuncName    string
	Level       int
}

func (si StackInfo) fullName() string {
	return fmt.Sprintf("%s.%s", si.PackageName, si.FuncName)
}

func (si StackInfo) isAnchor(s string) bool {
	if si.PackageName == "" && si.FuncName == "" {
		return true // cannot calculate anchor, calling algorithm set it zero
	}
	return strings.Contains(s, si.fullName())
}

// PrintStack prints to standard error the stack trace returned by runtime.Stack
// by starting from stackLevel.
func PrintStack(stackLevel int) {
	FprintStack(os.Stderr, StackInfo{Level: stackLevel})
}

// FprintStack prints to standard error the stack trace returned by runtime.Stack
// by starting from stackLevel.
func FprintStack(w io.Writer, si StackInfo) {
	stackBuf := bytes.NewBuffer(debug.Stack())
	scanner := bufio.NewScanner(stackBuf)

	// there is a caption line first, that's why we start from -1
	anchorLine := 0xffff
	for i := -1; scanner.Scan(); i++ {
		line := scanner.Text()
		if si.isAnchor(line) {
			anchorLine = i
		}
		if i == -1 || i/2 >= si.Level+anchorLine {
			fmt.Fprintln(w, line)
		}
	}
}
