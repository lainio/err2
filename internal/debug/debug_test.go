package debug

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/lainio/err2/internal/test"
)

func TestFullName(t *testing.T) {
	type args struct {
		StackInfo
	}
	tests := []struct {
		name string
		args
		retval string
	}{
		{"all empty", args{StackInfo{"", "", 0, nil}}, ""},
		{"namespaces", args{StackInfo{"lainio/err2", "", 0, nil}}, "lainio/err2"},
		{"both", args{StackInfo{"lainio/err2", "try", 0, nil}}, "lainio/err2.try"},
		{"short both", args{StackInfo{"err2", "Handle", 0, nil}}, "err2.Handle"},
		{"func", args{StackInfo{"", "try", 0, nil}}, "try"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.Requiref(t, tt.retval == tt.fullName(), "must be equal: %s",
				tt.retval)
		})
	}
}

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
		{"panic func and short regexp", args{
			"github.com/lainio/err2.Return(0x14001c1ee20)",
			StackInfo{"", "panic(", 0, PackageRegexp}}, true},
		{"func hit and regexp on", args{
			"github.com/lainioxx/err2_printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "printStackIf(", 0, noHitRegexp}}, false},
		{"short regexp no match", args{
			"github.com/lainioxx/err2_printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "", 0, noHitRegexp}}, false},
		{"short regexp", args{
			"github.com/lainio/err2/assert.That({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "", 0, PackageRegexp}}, true},
		{"short", args{
			"github.com/lainio/err2.printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "", 0, nil}}, true},
		{"short-but-false", args{
			"github.com/lainio/err2.printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"err2", "Handle", 0, nil}}, false},
		{"medium", args{
			"github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})",
			StackInfo{"err2", "Returnw", 0, nil}}, true},
		{"medium-but-false", args{
			"github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})",
			StackInfo{"err2", "Return(", 0, nil}}, false},
		{"long", args{
			"github.com/lainio/err2.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{"err2", "Handle", 0, nil}}, true},
		{"package name only", args{
			"github.com/lainio/err2/try.To1[...](...)",
			StackInfo{"lainio/err2", "", 0, nil}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.Require(t, tt.retval == tt.isAnchor(tt.input), "equal")
		})
	}
}

func TestIsFuncAnchor(t *testing.T) {
	type args struct {
		input string
		StackInfo
	}
	tests := []struct {
		name string
		args
		retval bool
	}{
		{"func hit and regexp on", args{
			"github.com/lainioxx/err2_printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "printStackIf(", 0, noHitRegexp}}, true},
		{"short regexp", args{
			"github.com/lainio/err2/assert.That({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "", 0, PackageRegexp}}, true},
		{"short", args{
			"github.com/lainio/err2.printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "", 0, nil}}, true},
		{"short-but-false", args{
			"github.com/lainio/err2.printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"err2", "Handle", 0, nil}}, false},
		{"medium", args{
			"github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})",
			StackInfo{"err2", "Returnw", 0, nil}}, true},
		{"medium-but-false", args{
			"github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})",
			StackInfo{"err2", "Return(", 0, nil}}, false},
		{"long", args{
			"github.com/lainio/err2.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{"err2", "Handle", 0, nil}}, true},
		{"package name only", args{
			"github.com/lainio/err2/try.To1[...](...)",
			StackInfo{"lainio/err2", "", 0, nil}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.Require(t, tt.retval == tt.isFuncAnchor(tt.input), "equal")
		})
	}
}

func TestFnLNro(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output int
	}{
		{"ext package",
			"	/Users/harrilainio/go/pkg/mod/github.com/lainio/err2@v0.8.5/internal/handler/handler.go:69 +0xbc",
			69},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := fnLNro(tt.input)
			test.Require(t, output == tt.output, output)
		})
	}
}

func TestFnName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"panic", "panic({0x102ed30c0, 0x1035910f0})",
			"panic"},
		{"our namespace", "github.com/lainio/err2/internal/debug.FprintStack({0x102ff7e88, 0x14000010020}, {{0x0, 0x0}, {0x102c012b8, 0x6}, 0x1, 0x140000bcb40})",
			"debug.FprintStack"},
		{"our namespace and func1", "github.com/lainio/err2/internal/debug.FprintStack.func1({0x102ff7e88, 0x14000010020}, {{0x0, 0x0}, {0x102c012b8, 0x6}, 0x1, 0x140000bcb40})",
			"debug.FprintStack"},
		{"our double namespace", "github.com/lainio/err2/internal/handler.Info.callPanicHandler({{0x102ed30c0, 0x1035910f0}, {0x102ff7e88, 0x14000010020}, 0x0, 0x140018643e0, 0x0})",
			"handler.Info.callPanicHandler"},
		{"our handler process", "github.com/lainio/err2/internal/handler.Process.func1({{0x102ed30c0, 0x1035910f0}, {0x102ff7e88, 0x14000010020}, 0x0, 0x140018643e0, 0x0})",
			"handler.Process"},
		{"our handler process and more anonymous funcs", "github.com/lainio/err2/internal/handler.Process.func1.2({{0x102ed30c0, 0x1035910f0}, {0x102ff7e88, 0x14000010020}, 0x0, 0x140018643e0, 0x0})",
			"handler.Process"},
		{"method and package name", "github.com/findy-network/findy-agent/agent/ssi.(*DIDAgent).AssertWallet(...)",
			"ssi.(*DIDAgent).AssertWallet"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := fnName(tt.input)
			test.Require(t, output == tt.output, output)
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
			test.Require(t, tt.input == w.String(), "")
		})
	}
}

func TestCalcAnchor(t *testing.T) {
	type args struct {
		input string
		StackInfo
	}
	tests := []struct {
		name string
		args
		anchor int
	}{
		{"macOS from test using regexp", args{inputFromMac, StackInfo{"", "panic(", 1, PackageRegexp}}, 12},
		{"short", args{input, StackInfo{"", "panic(", 0, nil}}, 6},
		{"short error stack", args{inputByError, StackInfo{"", "panic(", 0, PackageRegexp}}, 4},
		{"short and nolimit", args{input, StackInfo{"", "", 0, nil}}, nilAnchor},
		{"medium", args{input1, StackInfo{"", "panic(", 0, nil}}, 10},
		{"from test using panic", args{inputFromTest, StackInfo{"", "panic(", 0, nil}}, 8},
		{"from test", args{inputFromTest, StackInfo{"", "panic(", 0, PackageRegexp}}, 14},
		{"macOS from test using panic", args{inputFromMac, StackInfo{"", "panic(", 0, nil}}, 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			anchor := calcAnchor(r, tt.StackInfo)
			test.Require(t, tt.anchor == anchor, "equal")
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
		{"short", args{input, StackInfo{"err2", "Returnw(", 0, nil}}, output},
		{"medium", args{input1, StackInfo{"err2", "Returnw(", 0, nil}}, output1},
		{"medium level 2", args{input1, StackInfo{"err2", "Returnw(", 2, nil}}, output12},
		{"medium panic", args{input1, StackInfo{"", "panic(", 0, nil}}, output1panic},
		{"long", args{input2, StackInfo{"err2", "Handle(", 0, nil}}, output2},
		{"long lvl 2", args{input2, StackInfo{"err2", "Handle(", 3, nil}}, output23},
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
			test.Require(t, len(ins) > len(outs), "input length should be greater")
			test.Require(t, tt.output == w.String(), "not equal")
		})
	}
}

func TestFuncName(t *testing.T) {
	type args struct {
		input string
		StackInfo
	}
	tests := []struct {
		name string
		args
		output string
		outln  int
	}{
		{"basic", args{input2, StackInfo{"", "Handle", 1, nil}}, "err2.ReturnW", 214},
		{"basic lvl 3", args{input2, StackInfo{"", "Handle", 3, nil}}, "err2.ReturnW", 214},
		{"basic lvl 2", args{input2, StackInfo{"lainio/err2", "Handle", 1, nil}}, "err2.ReturnW", 214},
		{"method", args{inputFromTest, StackInfo{"", "Handle", 1, nil}}, "ssi.(*DIDAgent).AssertWallet", 146},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			name, ln, ok := funcName(r, StackInfo{
				PackageName: tt.PackageName,
				FuncName:    tt.FuncName,
				Level:       tt.Level,
			})
			test.Require(t, ok, "not found")
			test.Requiref(t, tt.output == name, "not equal %v", name)
			test.Requiref(t, ln == tt.outln, "ln must be equal %d == %d", ln, tt.outln)
		})
	}
}

// Testing variables

var (
	// noHitRegexp is a testing regexp to cause no hits.
	noHitRegexp = regexp.MustCompile(`^$`)

	inputFromMac = `goroutine 518 [running]:
runtime/debug.Stack()
	/opt/homebrew/Cellar/go/1.18/libexec/src/runtime/debug/stack.go:24 +0x64
github.com/lainio/err2/internal/debug.FprintStack({0x102ff7e88, 0x14000010020}, {{0x0, 0x0}, {0x102c012b8, 0x6}, 0x1, 0x140000bcb40})
	/Users/harrilainio/go/pkg/mod/github.com/lainio/err2@v0.8.5/internal/debug/debug.go:58 +0x40
github.com/lainio/err2/internal/handler.printStack({0x102ff7e88, 0x14000010020}, {{0x0, 0x0}, {0x102c012b8, 0x6}, 0x1, 0x140000bcb40}, {0x102ed30c0, 0x1035910f0})
	/Users/harrilainio/go/pkg/mod/github.com/lainio/err2@v0.8.5/internal/handler/handler.go:69 +0xbc
github.com/lainio/err2/internal/handler.Info.callPanicHandler({{0x102ed30c0, 0x1035910f0}, {0x102ff7e88, 0x14000010020}, 0x0, 0x140018643e0, 0x0})
	/Users/harrilainio/go/pkg/mod/github.com/lainio/err2@v0.8.5/internal/handler/handler.go:45 +0x94
github.com/lainio/err2/internal/handler.Process({{0x102ed30c0, 0x1035910f0}, {0x102ff7e88, 0x14000010020}, 0x0, 0x140018643e0, 0x0})
	/Users/harrilainio/go/pkg/mod/github.com/lainio/err2@v0.8.5/internal/handler/handler.go:59 +0xa4
github.com/lainio/err2.Return(0x14001c1ee20)
	/Users/harrilainio/go/pkg/mod/github.com/lainio/err2@v0.8.5/err2.go:171 +0xec
panic({0x102ed30c0, 0x1035910f0})
	/opt/homebrew/Cellar/go/1.18/libexec/src/runtime/panic.go:844 +0x26c
github.com/findy-network/findy-agent/agent/cloud.(*Agent).PwPipe(0x1400024e2d0, {0x140017ad770, 0x24})
	/Users/harrilainio/go/src/github.com/findy-network/findy-agent/agent/cloud/agent.go:202 +0x324
github.com/findy-network/findy-agent/agent/prot.StartPSM({{0x102c36e41, 0x37}, {0x102c3c16e, 0x40}, {0x103006588, 0x140003325a0}, {0x103003ad0, 0x14001c168c0}, 0x102fee8b0})
	/Users/harrilainio/go/src/github.com/findy-network/findy-agent/agent/prot/processor.go:75 +0x244
github.com/findy-network/findy-agent/protocol/trustping.startTrustPing({0x103006588, 0x140003325a0}, {0x103003ad0, 0x14001c168c0})
	/Users/harrilainio/go/src/github.com/findy-network/findy-agent/protocol/trustping/trust_ping_protocol.go:58 +0xec
created by github.com/findy-network/findy-agent/agent/prot.FindAndStartTask
	/Users/harrilainio/go/src/github.com/findy-network/findy-agent/agent/prot/processor.go:337 +0x21c
`

	inputFromTest = `goroutine 31 [running]:
testing.tRunner.func1.2({0xa8e0e0, 0x40001937d0})
        /usr/local/go/src/testing/testing.go:1389 +0x1c8
testing.tRunner.func1()
        /usr/local/go/src/testing/testing.go:1392 +0x380
panic({0xa8e0e0, 0x40001937d0})
        /usr/local/go/src/runtime/panic.go:838 +0x20c
github.com/lainio/err2.Handle(0xd14818)
        /home/god/go/src/github.com/lainio/err2/err2.go:133 +0xac
panic({0xa8e0e0, 0x40001937d0})
        /usr/local/go/src/runtime/panic.go:838 +0x20c
github.com/lainio/err2/assert.Asserter.reportPanic(...)
        /home/god/go/src/github.com/lainio/err2/assert/asserter.go:165
github.com/lainio/err2/assert.Asserter.reportAssertionFault(0x0, {0xba0fe9?, 0x0?}, {0x0?, 0x0, 0x0})
        /home/god/go/src/github.com/lainio/err2/assert/asserter.go:147 +0x21c
github.com/lainio/err2/assert.Asserter.True(...)
        /home/god/go/src/github.com/lainio/err2/assert/asserter.go:49
github.com/findy-network/findy-agent/agent/ssi.(*DIDAgent).AssertWallet(...)
        /home/god/go/src/github.com/findy-network/findy-agent/agent/ssi/agent.go:146
github.com/findy-network/findy-agent/agent/ssi.(*DIDAgent).myCreateDID(0x40003f92c0?, {0x0?, 0x0?})
        /home/god/go/src/github.com/findy-network/findy-agent/agent/ssi/agent.go:274 +0x78
github.com/findy-network/findy-agent/agent/ssi.(*DIDAgent).NewDID(0x40003f92c0?, 0x40000449a0?, {0x0?, 0x0?})
        /home/god/go/src/github.com/findy-network/findy-agent/agent/ssi/agent.go:230 +0x60
github.com/findy-network/findy-agent/agent/sec_test.TestPipe_packPeer(0x4000106d00?)
        /home/god/go/src/github.com/findy-network/findy-agent/agent/sec/pipe_test.go:355 +0x1b8
testing.tRunner(0x4000106d00, 0xd14820)
        /usr/local/go/src/testing/testing.go:1439 +0x110
`

	inputByError = `goroutine 1 [running]:
panic({0x137b20, 0x400007ac60})
	/usr/local/go/src/runtime/panic.go:838 +0x20c
github.com/lainio/err2/try.To(...)
	/home/god/go/src/github.com/lainio/err2/try/try.go:50
github.com/lainio/err2/try.To1[...](...)
	/home/god/go/src/github.com/lainio/err2/try/try.go:58
main.test1()
	/home/god/go/src/github.com/lainio/ic/main.go:29 +0x110
main.main()
	/home/god/go/src/github.com/lainio/ic/main.go:73 +0x1b0
`

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
github.com/harri/err2.ReturnW(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})
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
github.com/harri/err2.ReturnW(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})
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
