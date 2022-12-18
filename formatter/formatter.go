// Package formatter implements formatters and helper types.
package formatter

import (
	"github.com/lainio/err2/internal/str"
)

type Interface interface {
	Format(input string) string
}

type DoFmt func(i string) string

type Formatter struct {
	DoFmt
}

var (
	Decamel = &Formatter{DoFmt: str.Decamel}
	Noop    = &Formatter{DoFmt: func(i string) string { return i }}
)

func (f *Formatter) Format(input string) string {
	return f.DoFmt(input)
}
