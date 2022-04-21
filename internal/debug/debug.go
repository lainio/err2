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
	if si.PackageName != "" {
		return fmt.Sprintf("%s.%s", si.PackageName, si.FuncName)
	}
	return si.FuncName
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

// FprintStack prints to w the stack trace returned by runtime.Stack by starting
// from StackInfo.
func FprintStack(w io.Writer, si StackInfo) {
	stackBuf := bytes.NewBuffer(debug.Stack())
	stackPrint(stackBuf, w, si)
}

// stackPrint prints to standard error the stack trace returned by runtime.Stack
// by starting from stackLevel.
func stackPrint(r io.Reader, w io.Writer, si StackInfo) {
	scanner := bufio.NewScanner(r)

	// there is a caption line first, that's why we start from -1
	anchorLine := 0xffff
	for i := -1; scanner.Scan(); i++ {
		line := scanner.Text()

		// anchorLine can set when it's not yet set AND this is not the caption
		// line AND it matches to StackInfo criteria
		canSetAnchorLine := anchorLine == 0xffff && i > -1 && si.isAnchor(line)

		if canSetAnchorLine {
			anchorLine = i
		}

		// line can print when it is a caption OR there is no anchorLine
		// criteria OR the line (pair) is creater than anchorLine
		canPrint := i == -1 || 0 == si.Level+anchorLine || i >= 2*si.Level+anchorLine

		if canPrint {
			fmt.Fprintln(w, line)
		}
	}
}
