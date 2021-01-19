package assert

import (
	"fmt"
	"reflect"
)

// True asserts that term is true. If not it panics with the given formatting
// string.
func True(term bool, format string, a ...interface{}) {
	if !term {
		myPanic(fmt.Sprintf(format, a...))
		//panic(fmt.Errorf(format, a...))
	}
}

// Len asserts that length of the object is equal. If not it panics with the
// given formatting string.
func Len(obj interface{}, length int, format string, a ...interface{}) {
	ok, l := getLen(obj)
	if !ok {
		myPanic("cannot get length")
	}

	if l != length {
		myPanic(format, a...)
	}
}

func getLen(x interface{}) (ok bool, length int) {
	v := reflect.ValueOf(x)
	defer func() {
		if e := recover(); e != nil {
			ok = false
		}
	}()
	return true, v.Len()
}

func myPanic(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
