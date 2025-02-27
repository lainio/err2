package str

import (
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode"

	"github.com/lainio/err2/internal/x"
)

var (
	uncamel = regexp.MustCompile(`([A-Z]+)`)
	clean   = regexp.MustCompile(`[^\w]`)
)

// DecamelRegexp return the given string as space delimeted. Note! It's slow.
// Use [Decamel] instead. It's left here for learning purposes.
func DecamelRegexp(str string) string {
	str = clean.ReplaceAllString(str, " ")
	str = uncamel.ReplaceAllString(str, ` $1`)
	str = strings.Trim(str, " ")
	str = strings.ToLower(str)
	return str
}

func DecamelRmTryPrefix(s string) string {
	return strings.ReplaceAll(Decamel(s), "try ", "")
}

// Decamel return the given string as space delimeted. It's optimized to split
// and decamel function names returned from Go call stacks. For more information
// see its test cases.
func Decamel(s string) string {
	var (
		b           strings.Builder
		splittable  bool
		isUpper     bool
		prevSkipped bool
	)
	b.Grow(2 * len(s))

	for i, v := range s {
		skip := v == '(' || v == ')' || v == '*'
		if skip {
			if !prevSkipped && i != 0 { // first time write space
				b.WriteRune(' ')
			}
			prevSkipped = skip
			continue
		}
		toSpace := v == '.' || v == '_'
		if toSpace {
			if prevSkipped {
				continue
			} else if v == '.' {
				b.WriteRune(':')
			}
			v = ' '
			prevSkipped = true
		} else {
			isUpper = unicode.IsUpper(v)
			if isUpper {
				v = unicode.ToLower(v)
				if !prevSkipped && splittable {
					b.WriteRune(' ')
					prevSkipped = true
				}
			} else {
				prevSkipped = false
			}
		}
		b.WriteRune(v)
		splittable = !isUpper || unicode.IsNumber(v)
	}
	return b.String()
}

// FuncName is similar to runtime.Caller, but instead to return program counter
// or function name with full path, FuncName returns just function name,
// separated filename, and line number. If frame cannot be found ok is false.
//
// See more information from runtime.Caller. The skip tells how many stack
// frames are skipped. Note that FuncName calculates itself the skip frames.
func FuncName(skip int, long bool) (n, fname string, ln int, ok bool) {
	pc, file, ln, yes := runtime.Caller(skip + 1) // +1 skip ourself
	if yes {
		fn := runtime.FuncForPC(pc)
		fname = filepath.Base(file)
		ext := filepath.Ext(fname)
		trimmedFilename := strings.TrimSuffix(fname, ext) + "."
		n = strings.TrimPrefix(filepath.Base(fn.Name()), trimmedFilename)
	}
	fname = x.Whom(long, file, fname)
	return n, fname, ln, yes
}
