package debug

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
)

type StackInfo struct {
	PackageName string
	FuncName    string
	Level       int

	*regexp.Regexp
}

var (
	PackageRegexp = regexp.MustCompile(`lainio/err2/\w*\.`)
)

func (si StackInfo) fullName() string {
	dot := ""
	if si.PackageName != "" && si.FuncName != "" {
		dot = "."
	}
	return fmt.Sprintf("%s%s%s", si.PackageName, dot, si.FuncName)
}

func (si StackInfo) isAnchor(s string) bool {
	// Regexp matching is high priority. That's why it's the first one.
	if si.Regexp != nil {
		return si.Regexp.MatchString(s)
	}
	if si.PackageName == "" && si.FuncName == "" {
		return true // cannot calculate anchor, calling algorithm set it zero
	}
	return strings.Contains(s, si.fullName())
}

func (si StackInfo) isFuncAnchor(s string) bool {
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

// FprintStack prints the stack trace returned by runtime.Stack to the writer.
// The StackInfo tells what it prints from the stack.
func FprintStack(w io.Writer, si StackInfo) {
	stackBuf := bytes.NewBuffer(debug.Stack())
	stackPrint(stackBuf, w, si)
}

// stackPrint prints the stack trace read from reader and to the writer. The
// StackInfo tells what it prints from the stack.
func stackPrint(r io.Reader, w io.Writer, si StackInfo) {
	var buf bytes.Buffer
	r = io.TeeReader(r, &buf)
	anchorLine := calcAnchor(r, si)

	scanner := bufio.NewScanner(&buf)
	for i := -1; scanner.Scan(); i++ {
		line := scanner.Text()

		// we can print a line if we didn't find anything, i.e. anchorLine is
		// nilAnchor
		canPrint := anchorLine == nilAnchor
		// if it's not nilAnchor we need to check it more carefully
		if !canPrint {
			// we can print a line when it is a caption OR there is no
			// anchorLine criteria OR the line (pair) is creater than
			// anchorLine
			canPrint = i == -1 /*|| 0 == si.Level+anchorLine*/ || i >= 2*si.Level+anchorLine
		}

		if canPrint {
			fmt.Fprintln(w, line)
		}
	}
}

func calcAnchor(r io.Reader, si StackInfo) int {
	var buf bytes.Buffer
	r = io.TeeReader(r, &buf)

	anchor := calc(r, si, func(s string) bool {
		return si.isAnchor(s)
	})

	needToCalcFnNameAnchor := si.FuncName != "" && si.Regexp != nil
	if needToCalcFnNameAnchor {
		fnNameAnchor := calc(&buf, si, func(s string) bool {
			return si.isFuncAnchor(s)
		})
		if fnNameAnchor != nilAnchor && fnNameAnchor > anchor {
			return fnNameAnchor
		}
	}
	return anchor
}

func calc(r io.Reader, si StackInfo, anchor func(s string) bool) int {
	scanner := bufio.NewScanner(r)

	// there is a caption line first, that's why we start from -1
	anchorLine := nilAnchor
	var i int
	for i = -1; scanner.Scan(); i++ {
		line := scanner.Text()

		// anchorLine can set when it is not the caption
		// line AND it matches to StackInfo criteria
		canSetAnchorLine := i > -1 && anchor(line)

		if canSetAnchorLine {
			anchorLine = i
		}
	}
	if i-1 == anchorLine {
		return nilAnchor
	}
	return anchorLine
}

const nilAnchor = 0xffff
