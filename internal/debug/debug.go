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

	"github.com/lainio/err2/internal/x"
)

// StackInfo has two parts. The first part is for anchor line, i.e., line in the call
// stack that we want to include, and where output starts. The second part is
// ExlRegexps that are used to filter out lines from final output.
type StackInfo struct {
	PackageName string
	FuncName    string
	Level       int

	*regexp.Regexp

	// these are used to filter out specific lines from output
	ExlRegexp []*regexp.Regexp

	PrintFirstOnly bool
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
	packageRegexp = regexp.MustCompile(
		`^github\.com/lainio/err2[a-zA-Z0-9_/\.\[\]\@]*\(`,
	)

	// testing package exluding regexps:
	testingPkgRegexp  = regexp.MustCompile(`^testing\.`)
	testingFileRegexp = regexp.MustCompile(`^.*\/src\/testing\/testing\.go`)

	exludeRegexps    = []*regexp.Regexp{testingPkgRegexp, testingFileRegexp}
	exludeRegexpsAll = []*regexp.Regexp{
		testingPkgRegexp,
		testingFileRegexp,
		packageRegexp,
	}
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
	fullName := si.fullName()

	// If only package name is provided (no function name), match the package name as-is
	if si.FuncName == "" {
		return strings.Contains(s, fullName)
	}

	// Build the needle to search for
	var needle string
	if strings.HasSuffix(fullName, "(") {
		// FuncName already includes "(", use as-is
		needle = fullName
	} else {
		// Add "(" to match function calls
		needle = fullName + "("
	}

	// If there's no package name (only function name like "Handle"), we need to ensure
	// we match at a function boundary (preceded by .) to prevent matching user functions
	// like "FirstHandle" when searching for "Handle".
	// This is critical because "FirstHandle(" contains "Handle(" as a substring.
	if si.PackageName == "" && !strings.HasSuffix(si.FuncName, "(") {
		searchFrom := 0
		for {
			idx := strings.Index(s[searchFrom:], needle)
			if idx == -1 {
				return false
			}
			// Adjust idx to absolute position in the original string
			absIdx := searchFrom + idx

			// Only allow a '.' boundary and require "err2" immediately before it.
			// Note: Go stack lines never start with a bare function name, so we don't check absIdx == 0.
			if absIdx > 0 && isErr2BeforeDot(s, absIdx) {
				return true
			}

			// Continue search after this match
			searchFrom = absIdx + 1
		}
	}

	return strings.Contains(s, needle)
}

// isErr2BeforeDot returns true if the token immediately before the dot is "err2",
// or a major version token "vN" whose previous path segment is "err2".
// absIdx points to the start of the function name; absIdx-1 must be '.'.
// Examples:
//   - "github.com/lainio/err2.Handle(...)" -> true (package token is "err2")
//   - "github.com/lainio/err2/v2.Handle(...)" -> true (versioned module)
//   - "err2.Handle(...)" -> true (package token is "err2")
//   - "main.Handle(...)" -> false (package token is "main")
//   - "mypackage.Handle(...)" -> false (package token is "mypackage")
func isErr2BeforeDot(s string, absIdx int) bool {
	dot := absIdx - 1
	if dot < 0 || s[dot] != '.' {
		return false
	}

	// Find the last '/' before the dot to isolate the package token.
	start := strings.LastIndex(s[:dot], "/")
	if start == -1 {
		start = 0
	} else {
		start++ // exclude '/'
	}
	pkg := s[start:dot]
	if pkg == "err2" {
		return true
	}

	// Support versioned imports like ".../err2/v2.Handle".
	// If pkg == "vN" and the previous path token == "err2", accept.
	if len(pkg) > 1 && pkg[0] == 'v' {
		allDigits := true
		for i := 1; i < len(pkg); i++ {
			if pkg[i] < '0' || pkg[i] > '9' {
				allDigits = false
				break
			}
		}
		if allDigits && start > 0 {
			prev := strings.LastIndex(s[:start-1], "/")
			if prev != -1 && s[prev+1:start-1] == "err2" {
				return true
			}
		}
	}

	return false
}

func (si StackInfo) needToCalcFnNameAnchor() bool {
	return si.FuncName != "" && si.Regexp != nil
}

// isLvlOnly return true if all fields are nil and Level != 0 that should be
// used then.
func (si StackInfo) isLvlOnly() bool {
	return si.Level != 0 && si.Regexp == nil && si.PackageName == "" &&
		si.FuncName == ""
}

func (si StackInfo) canPrint(s string, anchorLine, i int) (ok bool) {
	if si.isLvlOnly() {
		// we don't need it now because only Lvl is used to decide what's
		// printed from call stack.
		anchorLine = 0
	}
	if si.PrintFirstOnly {
		ok = i >= 2*si.Level+anchorLine && i < 2*si.Level+anchorLine+2
	} else {
		ok = i >= 2*si.Level+anchorLine
	}

	if si.ExlRegexp == nil {
		return ok
	}

	// if any of the ExlRegexp match we don't print
	for _, reg := range si.ExlRegexp {
		if reg.MatchString(s) {
			return false
		}
	}
	return ok
}

// PrintStackForTest prints to io.Writer the stack trace returned by
// runtime.Stack and processed to proper format to be shown in test output by
// starting from stackLevel.
func PrintStackForTest(w io.Writer, stackLevel int) {
	stack := debug.Stack()
	//println(string(stack))
	stackBuf := bytes.NewBuffer(stack)
	printStackForTest(stackBuf, w, stackLevel)
}

// printStackForTest prints to io.Writer the stack trace returned by
// runtime.Stack and processed to proper format to be shown in test output by
// starting from stackLevel.
func printStackForTest(r io.Reader, w io.Writer, stackLevel int) {
	build := make([]string, 0, 24)
	buf := new(bytes.Buffer)
	stackPrint(r, buf, StackInfo{Level: stackLevel, ExlRegexp: exludeRegexps})
	scanner := bufio.NewScanner(buf)
	funcName := ""
	for i := -1; scanner.Scan(); i++ {
		line := scanner.Text()
		if i == -1 {
			continue
		}
		if i%2 == 0 {
			funcName = fnName(line)
		} else {
			line = strings.TrimPrefix(line, "\t")
			s := strings.Split(line, " ")
			out := fmt.Sprintf("    %s: %s", s[0], funcName)
			build = append(build, out)
		}
	}
	buildReverse := x.SReverse(build)
	for i, line := range buildReverse {
		fmt.Fprint(w, line+x.Whom(i > 0, " STACK\n", "\n"))
	}
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
func FuncName(si StackInfo) (n string, ln int, frame int, ok bool) {
	stack := debug.Stack()
	//println(string(stack))
	stackBuf := bytes.NewBuffer(stack)
	return funcName(stackBuf, si)
}

// funcName see Funcname documentation.
func funcName(r io.Reader,
	si StackInfo,
) (
	n string,
	ln int,
	frame int,
	ok bool,
) {
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
				return n, ln, i / 2, ok
			}

			// we are interested the line before (2 x si.Level) the
			// anchorLine, AND we want to calc this only once
			reachAnchor = x.Whom(reachAnchor, true,
				i == (anchorLine-2*si.Level))

			foundIt := reachAnchor && i%2 == 0 && notOurFunction(line)
			if foundIt {
				n = fnName(line)
				ok = n != "panic"
			}
		}
	}
	return n, 0, -1, false
}

// notOurFunction returns true if function in call stack line isn't from err2
// package.
func notOurFunction(line string) bool {
	return !packageRegexp.MatchString(line)
}

// fnName returns cleaned name of the function in the call stack line.
func fnName(line string) string {
	// remove main pkg name from func names because it ruins error msgs.
	line = strings.TrimPrefix(line, "main.")

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

	tryFnNames := []string{".T.", ".T1[...]", ".T2[...]", ".T3[...]"}
	line = cleanFuncNames(line[i:j], tryFnNames, true)

	fnNames := []string{".func", "func"}
	line = cleanFuncNames(line, fnNames, false)
	return line
}

func cleanFuncNames(line string, names []string, concat bool) string {
	for _, name := range names {
		b, e, found := strings.Cut(line, name)
		if found {
			line = b
			if concat {
				line += e
			}
		}
	}
	return line
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
	anchorLine := calcAnchor(r, si) // the line we want to start show stack

	scanner := bufio.NewScanner(&buf)
	for i := -1; scanner.Scan(); i++ {
		line := scanner.Text()

		// we can print a line if we didn't find anything, i.e. anchorLine is
		// nilAnchor, which means that our start is not limited by the anchor
		canPrint := anchorLine == nilAnchor
		// if it's not nilAnchor we need to check it more carefully
		if !canPrint {
			// we can print a line when it's a caption OR the line (pair) is
			// greater than anchorLine
			canPrint = i == -1 || si.canPrint(line, anchorLine, i)
		}

		if canPrint {
			fmt.Fprintln(w, line)
		}
	}
}

// calcAnchor calculates the optimal anchor line. Optimal is the shortest but
// including all the needed information.
func calcAnchor(r io.Reader, si StackInfo) int {
	if si.isLvlOnly() {
		// these are buffers, there's no error, but because we use TeeReader
		// we need to read all before return, otherwise caller gets nothing.
		_, _ = io.ReadAll(r)
		return si.Level
	}
	var buf bytes.Buffer
	r = io.TeeReader(r, &buf)

	anchor := calc(r, func(s string) bool {
		return si.isAnchor(s)
	})

	if si.needToCalcFnNameAnchor() {
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
