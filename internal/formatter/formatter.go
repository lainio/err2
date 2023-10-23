// Package formatter imlements thread safe storage for Formatter interface.
package formatter

import (
	"sync/atomic"

	format "github.com/lainio/err2/formatter"
)

var (
	formatter atomic.Value
)

func SetFormatter(fmter format.Interface) {
	formatter.Store(fmter)
}

func Formatter() format.Interface {
	fmter, isInterface := formatter.Load().(format.Interface)
	if isInterface {
		return fmter
	}
	return nil
}
