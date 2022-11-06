package str

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	re = regexp.MustCompile(`([A-Z]+)`)
)

// CamelRegexp return the given string as space delimeted. Note! it's slow. Use
// Decamel instead.
func CamelRegexp(str string) string {
	str = re.ReplaceAllString(str, ` $1`)
	str = strings.Trim(str, " ")
	return str
}

// Decamel return the given string as space delimeted.
func Decamel(s string) string {
	var b strings.Builder
	splittable, roundUpper, _ := false, false, false
	for _, v := range s {
		roundUpper = unicode.IsUpper(v)
		if roundUpper {
			v = unicode.ToLower(v)
			if splittable {
				b.WriteByte(' ')
			}
		}
		b.WriteRune(v)
		splittable = !roundUpper || unicode.IsNumber(v)
	}
	return b.String()
}
