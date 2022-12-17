package debug

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/lainio/err2/internal/helper"
)

type StackInfo struct {
	PackageName string
	FuncName    string
	Level       int

	*regexp.Regexp
}

var (
	// PackageRegexp is regexp search that help us find those lines that
	// includes function calls in our package and its sub packages. The
	// following lines help you figure out what kind of lines we are talking
	// about:
	//   github.com/lainio/err2/try.To1[...](...)
	//   github.com/lainio/err2/assert.Asserter.True(...)
	PackageRegexp = regexp.MustCompile(`lainio/err2[a-zA-Z0-9_/.\[\]]*\(`)

	// we want to check that this is not our package
	packageRegexp = regexp.MustCompile(`^github\.com/lainio/err2[a-zA-Z0-9_/.\[\]]*\(`)
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
	return si.isFuncAnchor(s)
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

// FuncName is similar to runtime.Caller, but instead to return program counter
// or function name with full path, funcName returns just function name,
// separated filename, and line number. If frame cannot be found ok is false.
//
// See more information from runtime.Caller. The StackInfo tells how many stack
// frames we should go back (Level), and other fields tell how to find the
// actual line where calculation should be started.
func FuncName(si StackInfo) (n string, ln int, ok bool) {
	stackBuf := bytes.NewBuffer(debug.Stack())
	return funcName(stackBuf, si)
}

// funcName see Funcname documentation.
func funcName(r io.Reader, si StackInfo) (n string, ln int, ok bool) {
	var buf bytes.Buffer
	stackBuf := io.TeeReader(r, &buf)
	anchorLine := calcAnchor(stackBuf, si)
	if anchorLine != nilAnchor {
		scanner := bufio.NewScanner(&buf)
		reachAnchor := false
		for i := -1; scanner.Scan(); i++ {
			line := scanner.Text()

			// if we have found the the actual line we need to process next
			// aka this line to get ln
			if ok {
				ln = fnLNro(line)
				return n, ln, ok
			}

			// we are interested the line before (2 x si.Level) the
			// anchorLine, AND we want to calc this only once
			reachAnchor = helper.Whom(reachAnchor, true,
				i == (anchorLine-2*si.Level))

			if reachAnchor && i%2 == 0 && notOurFunction(line) {
				n = fnName(line)
				ok = n != "panic"
			}
		}
	}
	return n, 0, false
}

// notOurFunction returns true if function in call stack line isn't from err2
// package.
func notOurFunction(line string) bool {
	return !packageRegexp.MatchString(line)
}

// fnName returns cleaned name of the function in the call stack line.
func fnName(line string) string {
	i := strings.LastIndex(line, "/")
	if i == -1 {
		i = 0
	} else {
		i++ // do not include '/'
	}

	j := strings.LastIndex(line[i:], "(")
	if j == -1 {
		j = len(line)
	} else {
		j += i
	}

	return strings.TrimSuffix(line[i:j], ".func1")
}

// fnLNro returns line number in the call stack line.
func fnLNro(line string) int {
	i := strings.LastIndex(line, "go:")
	if i == -1 {
		i = 0
	} else {
		i += 3 // do not include ':'
	}
	j := strings.LastIndex(line[i:], " ")
	if j == -1 {
		j = len(line)
	} else {
		j += i
	}
	nro, _ := strconv.Atoi(line[i:j])
	return nro
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
			// we can print a line when it's a caption OR the line (pair) is
			// greater than anchorLine
			canPrint = i == -1 || i >= 2*si.Level+anchorLine
		}

		if canPrint {
			fmt.Fprintln(w, line)
		}
	}
}

// calcAnchor calculates the optimal anchor line. Optimal is the shortest but
// including all the needed information.
func calcAnchor(r io.Reader, si StackInfo) int {
	var buf bytes.Buffer
	r = io.TeeReader(r, &buf)

	anchor := calc(r, func(s string) bool {
		return si.isAnchor(s)
	})

	needToCalcFnNameAnchor := si.FuncName != "" && si.Regexp != nil
	if needToCalcFnNameAnchor {
		fnNameAnchor := calc(&buf, func(s string) bool {
			return si.isFuncAnchor(s)
		})

		fnAnchorIsMoreOptimal := fnNameAnchor != nilAnchor &&
			fnNameAnchor > anchor
		if fnAnchorIsMoreOptimal {
			return fnNameAnchor
		}
	}
	return anchor
}

// calc calculates anchor line it takes criteria function as an argument.
func calc(r io.Reader, anchor func(s string) bool) int {
	scanner := bufio.NewScanner(r)

	// there is a caption line first, that's why we start from -1
	anchorLine := nilAnchor
	var i int
	for i = -1; scanner.Scan(); i++ {
		line := scanner.Text()

		// anchorLine can set when it's not the caption line AND it matches to
		// StackInfo criteria
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

const nilAnchor = 0xffff // reserve nilAnchor, remember we need -1 for algorithm
