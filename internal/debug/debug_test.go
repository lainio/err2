package debug

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	input = `goroutine 1 [running]:
github.com/lainio/err2.Handle(0x40000b5ed8, 0x40000b5ef8)
	/home/god/go/src/github.com/lainio/err2/err2.go:107 +0x10c
panic({0x12e3e0, 0x188f20})
	/usr/local/go/src/runtime/panic.go:838 +0x20c
github.com/lainio/err2.Returnw(0x40000b5e60, {0x0, 0x0}, {0x0, 0x0, 0x0})
	/home/god/go/src/github.com/lainio/err2/err2.go:214 +0x21c
panic({0x12e3e0, 0x188f20})
	/usr/local/go/src/runtime/panic.go:838 +0x20c
main.test0()
	/home/god/go/src/github.com/lainio/ic/main.go:18 +0x64
main.main()
	/home/god/go/src/github.com/lainio/ic/main.go:74 +0x1d0
`
	output = `goroutine 1 [running]:
github.com/lainio/err2.Returnw(0x40000b5e60, {0x0, 0x0}, {0x0, 0x0, 0x0})
	/home/god/go/src/github.com/lainio/err2/err2.go:214 +0x21c
panic({0x12e3e0, 0x188f20})
	/usr/local/go/src/runtime/panic.go:838 +0x20c
main.test0()
	/home/god/go/src/github.com/lainio/ic/main.go:18 +0x64
main.main()
	/home/god/go/src/github.com/lainio/ic/main.go:74 +0x1d0
`

	input1 = `goroutine 1 [running]:
runtime/debug.Stack()
	/usr/local/go/src/runtime/debug/stack.go:24 +0x68
github.com/lainio/err2/internal/debug.FprintStack({0x1896a8, 0x40000a6010}, {{0x15427b, 0x4}, {0x1548e7, 0x7}, 0x0})
	/home/god/go/src/github.com/lainio/err2/internal/debug/debug.go:40 +0x3c
github.com/lainio/err2.printStackIf({0x1548e7, 0x7}, 0x0, {0x12e3e0?, 0x188f50?})
	/home/god/go/src/github.com/lainio/err2/err2.go:315 +0xdc
github.com/lainio/err2.checkStackTracePrinting({0x1548e7, 0x7}, {0x12e3e0?, 0x188f50})
	/home/god/go/src/github.com/lainio/err2/err2.go:334 +0xbc
github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})
	/home/god/go/src/github.com/lainio/err2/err2.go:201 +0x64
panic({0x12e3e0, 0x188f50})
	/usr/local/go/src/runtime/panic.go:838 +0x20c
main.test0()
	/home/god/go/src/github.com/lainio/ic/main.go:18 +0x64
main.main()
	/home/god/go/src/github.com/lainio/ic/main.go:74 +0x1d0
`
	output1 = `goroutine 1 [running]:
github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})
	/home/god/go/src/github.com/lainio/err2/err2.go:201 +0x64
panic({0x12e3e0, 0x188f50})
	/usr/local/go/src/runtime/panic.go:838 +0x20c
main.test0()
	/home/god/go/src/github.com/lainio/ic/main.go:18 +0x64
main.main()
	/home/god/go/src/github.com/lainio/ic/main.go:74 +0x1d0
`

	output1panic = `goroutine 1 [running]:
panic({0x12e3e0, 0x188f50})
	/usr/local/go/src/runtime/panic.go:838 +0x20c
main.test0()
	/home/god/go/src/github.com/lainio/ic/main.go:18 +0x64
main.main()
	/home/god/go/src/github.com/lainio/ic/main.go:74 +0x1d0
`

	output12 = `goroutine 1 [running]:
main.test0()
	/home/god/go/src/github.com/lainio/ic/main.go:18 +0x64
main.main()
	/home/god/go/src/github.com/lainio/ic/main.go:74 +0x1d0
`

	input2 = `goroutine 1 [running]:
runtime/debug.Stack()
	/usr/local/go/src/runtime/debug/stack.go:24 +0x68
github.com/lainio/err2/internal/debug.FprintStack({0x1896a8, 0x40000a6010}, {{0x15427b, 0x4}, {0x1545d2, 0x6}, 0x0})
	/home/god/go/src/github.com/lainio/err2/internal/debug/debug.go:40 +0x3c
github.com/lainio/err2.printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})
	/home/god/go/src/github.com/lainio/err2/err2.go:315 +0xdc
github.com/lainio/err2.checkStackTracePrinting({0x1545d2, 0x6}, {0x12e3e0?, 0x188f50})
	/home/god/go/src/github.com/lainio/err2/err2.go:334 +0xbc
github.com/lainio/err2.Handle(0x40000b3ed8, 0x40000b3ef8)
	/home/god/go/src/github.com/lainio/err2/err2.go:89 +0x54
panic({0x12e3e0, 0x188f50})
	/usr/local/go/src/runtime/panic.go:838 +0x20c
github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})
	/home/god/go/src/github.com/lainio/err2/err2.go:214 +0x21c
panic({0x12e3e0, 0x188f50})
	/usr/local/go/src/runtime/panic.go:838 +0x20c
main.test0()
	/home/god/go/src/github.com/lainio/ic/main.go:18 +0x64
main.main()
	/home/god/go/src/github.com/lainio/ic/main.go:74 +0x1d0
`

	output2 = `goroutine 1 [running]:
github.com/lainio/err2.Handle(0x40000b3ed8, 0x40000b3ef8)
	/home/god/go/src/github.com/lainio/err2/err2.go:89 +0x54
panic({0x12e3e0, 0x188f50})
	/usr/local/go/src/runtime/panic.go:838 +0x20c
github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})
	/home/god/go/src/github.com/lainio/err2/err2.go:214 +0x21c
panic({0x12e3e0, 0x188f50})
	/usr/local/go/src/runtime/panic.go:838 +0x20c
main.test0()
	/home/god/go/src/github.com/lainio/ic/main.go:18 +0x64
main.main()
	/home/god/go/src/github.com/lainio/ic/main.go:74 +0x1d0
`
	output23 = `goroutine 1 [running]:
panic({0x12e3e0, 0x188f50})
	/usr/local/go/src/runtime/panic.go:838 +0x20c
main.test0()
	/home/god/go/src/github.com/lainio/ic/main.go:18 +0x64
main.main()
	/home/god/go/src/github.com/lainio/ic/main.go:74 +0x1d0
`
)

func TestIsAnchor(t *testing.T) {
	type args struct {
		input string
		StackInfo
	}
	tests := []struct {
		name string
		args
		retval bool
	}{
		{"short", args{
			"github.com/lainio/err2.printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "", 0}}, true},
		{"short-but-false", args{
			"github.com/lainio/err2.printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"err2", "Handle", 0}}, false},
		{"medium", args{
			"github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})",
			StackInfo{"err2", "Returnw", 0}}, true},
		{"medium-but-false", args{
			"github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})",
			StackInfo{"err2", "Return(", 0}}, false},
		{"long", args{
			"github.com/lainio/err2.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{"err2", "Handle", 0}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.retval, tt.isAnchor(tt.input))
		})
	}
}

func TestStackPrint_noLimits(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"short", input},
		{"medium", input1},
		{"long", input2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			w := new(bytes.Buffer)
			stackPrint(r, w, StackInfo{
				PackageName: "",
				FuncName:    "",
				Level:       0,
			})
			require.EqualValues(t, tt.input, w.String())
		})
	}
}

func TestCalcAnchor(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		anchor int
	}{
		{"short", input, 6},
		{"medium", input1, 10},
		{"long", input2, 14},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			anchor := calcAnchor(r, StackInfo{
				PackageName: "",
				FuncName:    "panic(",
				Level:       0,
			})
			require.EqualValues(t, tt.anchor, anchor)
		})
	}
}

func TestStackPrint_limit(t *testing.T) {
	type args struct {
		input string
		StackInfo
	}
	tests := []struct {
		name string
		args
		output string
	}{
		{"short", args{input, StackInfo{"err2", "Returnw(", 0}}, output},
		{"medium", args{input1, StackInfo{"err2", "Returnw(", 0}}, output1},
		{"medium level 2", args{input1, StackInfo{"err2", "Returnw(", 2}}, output12},
		{"medium panic", args{input1, StackInfo{"", "panic(", 0}}, output1panic},
		{"long", args{input2, StackInfo{"err2", "Handle(", 0}}, output2},
		{"long lvl 2", args{input2, StackInfo{"err2", "Handle(", 3}}, output23},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			w := new(bytes.Buffer)
			stackPrint(r, w, StackInfo{
				PackageName: tt.PackageName,
				FuncName:    tt.FuncName,
				Level:       tt.Level,
			})
			ins := strings.Split(tt.input, "\n")
			outs := strings.Split(w.String(), "\n")
			require.Greater(t, len(ins), len(outs), tt.FuncName)
			println(len(outs), "/", len(ins))
			require.Equal(t, tt.output, w.String())
		})
	}

}
