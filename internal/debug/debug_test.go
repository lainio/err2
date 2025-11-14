package debug

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/lainio/err2/internal/expect"
)

func TestFullName(t *testing.T) {
	t.Parallel()
	type args struct {
		StackInfo
	}
	type ttest struct {
		name string
		args
		retval string
	}
	tests := []ttest{
		{"all empty", args{StackInfo{"", "", 0, nil, nil, false}}, ""},
		{
			"namespaces",
			args{StackInfo{"lainio/err2", "", 0, nil, nil, false}},
			"lainio/err2",
		},
		{
			"both",
			args{StackInfo{"lainio/err2", "try", 0, nil, nil, false}},
			"lainio/err2.try",
		},
		{
			"short both",
			args{StackInfo{"err2", "Handle", 0, nil, nil, false}},
			"err2.Handle",
		},
		{"func", args{StackInfo{"", "try", 0, nil, nil, false}}, "try"},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			expect.Equal(t, tt.retval, tt.fullName())
		})
	}
}

func TestIsAnchor(t *testing.T) {
	t.Parallel()
	type args struct {
		input string
		StackInfo
	}
	type ttest struct {
		name string
		args
		retval bool
	}
	tests := []ttest{
		{"panic func and short regexp", args{
			"github.com/lainio/err2.Return(0x14001c1ee20)",
			StackInfo{"", "panic(", 0, PackageRegexp, nil, false}}, true},
		{"func hit and regexp on", args{
			"github.com/lainioxx/err2_printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "printStackIf(", 0, noHitRegexp, nil, false}}, false},
		{"short regexp no match", args{
			"github.com/lainioxx/err2_printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "", 0, noHitRegexp, nil, false}}, false},
		{"short regexp", args{
			"github.com/lainio/err2/assert.That({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "", 0, PackageRegexp, nil, false}}, true},
		{"short", args{
			"github.com/lainio/err2.printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "", 0, nil, nil, false}}, true},
		{"short-but-false", args{
			"github.com/lainio/err2.printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"err2", "Handle", 0, nil, nil, false}}, false},
		{"medium", args{
			"github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})",
			StackInfo{"err2", "Returnw", 0, nil, nil, false}}, true},
		{"medium-but-false", args{
			"github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})",
			StackInfo{"err2", "Return(", 0, nil, nil, false}}, false},
		{"long", args{
			"github.com/lainio/err2.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{"err2", "Handle", 0, nil, nil, false}}, true},
		{"package name only", args{
			"github.com/lainio/err2/try.To1[...](...)",
			StackInfo{"lainio/err2", "", 0, nil, nil, false}}, true},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			expect.Equal(t, tt.retval, tt.isAnchor(tt.input))
		})
	}
}

func TestIsFuncAnchor(t *testing.T) {
	t.Parallel()
	type args struct {
		input string
		StackInfo
	}
	type ttest struct {
		name string
		args
		retval bool
	}
	tests := []ttest{
		{"func hit and regexp on", args{
			"github.com/lainioxx/err2_printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "printStackIf(", 0, noHitRegexp, nil, false}}, true},
		{"short regexp", args{
			"github.com/lainio/err2/assert.That({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{"", "", 0, PackageRegexp, nil, false}}, true},
		{"short", args{
			"github.com/lainio/err2.printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{Err2PackageID, "", 0, nil, nil, false}}, true},
		{"short-but-false", args{
			"github.com/lainio/err2.printStackIf({0x1545d2, 0x6}, 0x0, {0x12e3e0?, 0x188f50?})",
			StackInfo{Err2PackageID, "Handle", 0, nil, nil, false}}, false},
		{"medium", args{
			"github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})",
			StackInfo{Err2PackageID, "Returnw", 0, nil, nil, false}}, true},
		{"medium-but-false", args{
			"github.com/lainio/err2.Returnw(0x40000b3e60, {0x0, 0x0}, {0x0, 0x0, 0x0})",
			StackInfo{Err2PackageID, "Return(", 0, nil, nil, false}}, false},
		{"long", args{
			"github.com/lainio/err2.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{Err2PackageID, "Handle", 0, nil, nil, false}}, true},
		{"package name only", args{
			"github.com/lainio/err2/try.To1[...](...)",
			StackInfo{Err2PackageID, "", 0, nil, nil, false}}, true},

		// From bug (issue #30 and PR #31) tests, Handle name in own pkg
		{"user pkg containing Handle should not match", args{
			"mainHandle.First(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{Err2PackageID, "Handle", 0, nil, nil, false}}, false},
		{"user function containing Handle should not match", args{
			"main.FirstHandle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{Err2PackageID, "Handle", 0, nil, nil, false}}, false},
		{"user function containing Handle matches err2.Handle", args{
			"github.com/lainio/err2.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{Err2PackageID, "Handle", 0, nil, nil, false}}, true},
		{"user function named exactly Handle should not match", args{
			"main.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{Err2PackageID, "Handle", 0, nil, nil, false}}, false},
		{"user package function named Handle should not match", args{
			"mypackage.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{"err2", "Handle", 0, nil, nil, false}}, false},
		{"err2.Handle should match", args{
			"err2.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{"err2", "Handle", 0, nil, nil, false}}, true},
		{"lainio/err2.Handle should match", args{
			"github.com/lainio/err2.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{Err2PackageID, "Handle", 0, nil, nil, false}}, true},
		{"user package with err2 in path should not match", args{
			"github.com/mycompany/err2/mypackage.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{Err2PackageID, "Handle", 0, nil, nil, false}}, false},
		{"non-err2 versioned package should not match", args{
			"github.com/someone/otherpkg/v2.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{Err2PackageID, "Handle", 0, nil, nil, false}}, false},

		// See the PackageID!!
		{"versioned err2/v2.Handle should match", args{
			"github.com/lainio/err2/v2.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{Err2PackageID + "/v2", "Handle", 0, nil, nil, false}}, true},
		{"versioned err2/v10.Handle should match", args{
			"github.com/lainio/err2/v10.Handle(0x40000b3ed8, 0x40000b3ef8)",
			StackInfo{Err2PackageID + "/v10", "Handle", 0, nil, nil, false}}, true},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			expect.Equal(t, tt.isFuncAnchor(tt.input), tt.retval)
		})
	}
}

func TestFnLNro(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		output int
	}{
		{"ext package",
			"	/Users/harrilainio/go/pkg/mod/github.com/lainio/err2@v0.8.5/internal/handler/handler.go:69 +0xbc",
			69},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := fnLNro(tt.input)
			expect.Equal(t, output, tt.output)
		})
	}
}

func TestFnName(t *testing.T) {
	t.Parallel()
	type ttest struct {
		name   string
		input  string
		output string
	}
	tests := []ttest{
		{"panic", "panic({0x102ed30c0, 0x1035910f0})",
			"panic"},
		{
			"our namespace",
			"github.com/lainio/err2/internal/debug.FprintStack({0x102ff7e88, 0x14000010020}, {{0x0, 0x0}, {0x102c012b8, 0x6}, 0x1, 0x140000bcb40})",
			"debug.FprintStack",
		},
		{
			"our namespace and func1",
			"github.com/lainio/err2/internal/debug.FprintStack.func1({0x102ff7e88, 0x14000010020}, {{0x0, 0x0}, {0x102c012b8, 0x6}, 0x1, 0x140000bcb40})",
			"debug.FprintStack",
		},
		{
			"our double namespace",
			"github.com/lainio/err2/internal/handler.Info.callPanicHandler({{0x102ed30c0, 0x1035910f0}, {0x102ff7e88, 0x14000010020}, 0x0, 0x140018643e0, 0x0})",
			"handler.Info.callPanicHandler",
		},
		{
			"our handler process",
			"github.com/lainio/err2/internal/handler.Process.func1({{0x102ed30c0, 0x1035910f0}, {0x102ff7e88, 0x14000010020}, 0x0, 0x140018643e0, 0x0})",
			"handler.Process",
		},
		{
			"our handler process and more anonymous funcs",
			"github.com/lainio/err2/internal/handler.Process.func1.2({{0x102ed30c0, 0x1035910f0}, {0x102ff7e88, 0x14000010020}, 0x0, 0x140018643e0, 0x0})",
			"handler.Process",
		},
		{
			"method and package name",
			"github.com/findy-network/findy-agent/agent/ssi.(*DIDAgent).AssertWallet(...)",
			"ssi.(*DIDAgent).AssertWallet",
		},
		{
			"try.T simple",
			"main.TCopyFile.T.func(...)",
			"TCopyFile",
		},
		{
			"try.T simple A",
			"main.TCopyFile.T.func3(...)",
			"TCopyFile",
		},
		{
			"try.T1",
			"ssi.TCopyFile.T1[...].func3(...)",
			"ssi.TCopyFile",
		},
		{
			"try.T1",
			"main.TCopyFile.T1[...].func3(...)",
			"TCopyFile",
		},
		{
			"try.T2 in not main pkg",
			"github.com/findy-network/findy-agent/agent/ssi.TCopyFile.T2[...].func3(...)",
			"ssi.TCopyFile",
		},
		{
			"try.T3",
			"main.TCopyFile.T3[...].func3(...)",
			"TCopyFile",
		},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := fnName(tt.input)
			expect.Equal(t, output, tt.output)
		})
	}
}

func TestStackPrint_noLimits(t *testing.T) {
	t.Parallel()
	type ttest struct {
		name  string
		input string
	}
	tests := []ttest{
		{"short", input},
		{"medium", input1},
		{"long", input2},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := strings.NewReader(tt.input)
			w := new(bytes.Buffer)
			stackPrint(r, w, StackInfo{
				PackageName: "",
				FuncName:    "",
				Level:       0,
			})
			expect.Equal(t, tt.input, w.String())
		})
	}
}

func TestStackPrintForTest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		output string
		lvl    int
	}{
		//{"short", input, outputForTest, 0},
		{"short", input, outputForTestLvl2, 2},
		//{"real test trace", inputFromTest, outputFromTest, 4},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := strings.NewReader(tt.input)
			w := new(bytes.Buffer)
			printStackForTest(r, w, tt.lvl)
			a, b := len(tt.output), len(w.String())
			// print(tt.output)
			// println("------")
			// print(w.String())
			expect.Equal(t, a, b)
			expect.Equal(t, tt.output, w.String())
		})
	}
}

func TestCalcAnchor(t *testing.T) {
	t.Parallel()
	type args struct {
		input string
		StackInfo
	}
	type ttest struct {
		name string
		args
		anchor int
	}
	tests := []ttest{
		{
			"macOS from test using ALL regexp",
			args{
				inputFromMac,
				StackInfo{"", "StartPSM(", 1, nil, exludeRegexpsAll, false},
			},
			16,
		},
		{
			"macOS from test using regexp",
			args{
				inputFromMac,
				StackInfo{"", "panic(", 1, PackageRegexp, nil, false},
			},
			12,
		},
		{"short", args{input, StackInfo{"", "panic(", 0, nil, nil, false}}, 6},
		{
			"short error stack",
			args{
				inputByError,
				StackInfo{"", "panic(", 0, PackageRegexp, nil, false},
			},
			4,
		},
		{
			"short and nolimit",
			args{input, StackInfo{"", "", 0, nil, nil, false}},
			nilAnchor,
		},
		{
			"short and only LVL is 2",
			args{input, StackInfo{"", "", 2, nil, nil, false}},
			2,
		},
		{"medium", args{input1, StackInfo{"", "panic(", 0, nil, nil, false}}, 10},
		{
			"from test using panic",
			args{inputFromTest, StackInfo{"", "panic(", 0, nil, nil, false}},
			8,
		},
		{
			"from test",
			args{
				inputFromTest,
				StackInfo{"", "panic(", 0, PackageRegexp, nil, false},
			},
			14,
		},
		{
			"macOS from test using panic",
			args{inputFromMac, StackInfo{"", "panic(", 0, nil, nil, false}},
			12,
		},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := strings.NewReader(tt.input)
			anchor := calcAnchor(r, tt.StackInfo)
			expect.Equal(t, tt.anchor, anchor)
		})
	}
}

func TestStackPrint_limit(t *testing.T) {
	t.Parallel()
	type args struct {
		input string
		StackInfo
	}
	type ttest struct {
		name string
		args
		output string
	}
	tests := []ttest{
		{
			"find function with FRAME from test stack",
			args{inputFromTest,
				StackInfo{"", "", 8, nil, exludeRegexpsAll, true}},
			outputFromTestOnlyFunction,
		},
		{
			"find function with FRAME from mac stack",
			args{inputFromMac,
				StackInfo{"", "", 7, nil, exludeRegexpsAll, true}},
			outputFromMacOneFunction,
		},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			readStack := strings.NewReader(tt.input)
			writeStack := new(bytes.Buffer)
			stackPrint(readStack, writeStack, tt.StackInfo)
			ins := strings.Split(tt.input, "\n")
			outs := strings.Split(writeStack.String(), "\n")
			expect.Thatf(t, len(ins) > len(outs),
				"input length:%d should be greater:%d", len(ins), len(outs))
			wantResult, gotResult := tt.output, writeStack.String()
			expect.Equal(t, gotResult, wantResult)
		})
	}
}

func TestStackPrint_OneFunction(t *testing.T) {
	t.Parallel()
	type args struct {
		input string
		StackInfo
	}
	type ttest struct {
		name string
		args
		output string
	}
	tests := []ttest{
		{
			"real test trace",
			args{inputFromTest, StackInfo{"", "", 8, nil, exludeRegexps, false}},
			outputFromTest,
		},
		{
			"only level 4",
			args{input1, StackInfo{"", "", 4, nil, nil, false}},
			output1,
		},
		{
			"short",
			args{input, StackInfo{"err2", "Returnw(", 0, nil, nil, false}},
			output,
		},
		{
			"medium",
			args{input1, StackInfo{"err2", "Returnw(", 0, nil, nil, false}},
			output1,
		},
		{
			"medium level 2",
			args{input1, StackInfo{"err2", "Returnw(", 2, nil, nil, false}},
			output12,
		},
		{
			"medium level 0",
			args{input1, StackInfo{"err2", "Returnw(", 0, nil, nil, false}},
			output1,
		},
		{
			"medium panic",
			args{input1, StackInfo{"", "panic(", 0, nil, nil, false}},
			output1panic,
		},
		{
			"long",
			args{input2, StackInfo{"err2", "Handle(", 0, nil, nil, false}},
			output2,
		},
		{
			"long lvl 2",
			args{input2, StackInfo{"err2", "Handle(", 3, nil, nil, false}},
			output23,
		},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := strings.NewReader(tt.input)
			w := new(bytes.Buffer)
			stackPrint(r, w, tt.StackInfo)
			ins := strings.Split(tt.input, "\n")
			outs := strings.Split(w.String(), "\n")
			expect.Thatf(t, len(ins) > len(outs),
				"input length:%d should be greater:%d", len(ins), len(outs))
			b, a := tt.output, w.String()
			expect.Equal(t, a, b)
		})
	}
}

func TestFuncName(t *testing.T) {
	t.Parallel()
	type args struct {
		input string
		StackInfo
	}
	type ttest struct {
		name string
		args
		output   string
		outln    int
		outFrame int
	}
	tests := []ttest{
		{
			"basic",
			args{input2, StackInfo{"", "Handle", 1, nil, nil, false}},
			"err2.ReturnW",
			214,
			6,
		},
		{
			"basic lvl 3",
			args{input2, StackInfo{"", "Handle", 3, nil, nil, false}},
			"err2.ReturnW",
			214,
			6,
		},
		{
			"basic lvl 2",
			args{input2, StackInfo{"lainio/err2", "Handle", 1, nil, nil, false}},
			"err2.ReturnW",
			214,
			6,
		},
		{
			"method",
			args{inputFromTest, StackInfo{"", "Handle", 1, nil, nil, false}},
			"ssi.(*DIDAgent).AssertWallet",
			146,
			8,
		},
		{
			"pipeline",
			args{inputPipelineStack, StackInfo{"", "Handle", -1, nil, nil, false}},
			"CopyFile",
			29,
			9,
		},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := strings.NewReader(tt.input)
			name, ln, fr, found := funcName(r, StackInfo{
				PackageName: tt.PackageName,
				FuncName:    tt.FuncName,
				Level:       tt.Level,
			})
			expect.That(t, found)
			expect.Equal(t, tt.output, name)
			expect.Equal(t, ln, tt.outln)
			expect.Equal(t, fr, tt.outFrame)
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

	outputFromMacOneFunction = `goroutine 518 [running]:
github.com/findy-network/findy-agent/agent/cloud.(*Agent).PwPipe(0x1400024e2d0, {0x140017ad770, 0x24})
	/Users/harrilainio/go/src/github.com/findy-network/findy-agent/agent/cloud/agent.go:202 +0x324
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

	outputFromTest = `goroutine 31 [running]:
github.com/findy-network/findy-agent/agent/ssi.(*DIDAgent).AssertWallet(...)
        /home/god/go/src/github.com/findy-network/findy-agent/agent/ssi/agent.go:146
github.com/findy-network/findy-agent/agent/ssi.(*DIDAgent).myCreateDID(0x40003f92c0?, {0x0?, 0x0?})
        /home/god/go/src/github.com/findy-network/findy-agent/agent/ssi/agent.go:274 +0x78
github.com/findy-network/findy-agent/agent/ssi.(*DIDAgent).NewDID(0x40003f92c0?, 0x40000449a0?, {0x0?, 0x0?})
        /home/god/go/src/github.com/findy-network/findy-agent/agent/ssi/agent.go:230 +0x60
github.com/findy-network/findy-agent/agent/sec_test.TestPipe_packPeer(0x4000106d00?)
        /home/god/go/src/github.com/findy-network/findy-agent/agent/sec/pipe_test.go:355 +0x1b8
`

	outputFromTestOnlyFunction = `goroutine 31 [running]:
github.com/findy-network/findy-agent/agent/ssi.(*DIDAgent).AssertWallet(...)
        /home/god/go/src/github.com/findy-network/findy-agent/agent/ssi/agent.go:146
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

	// outputForTest is printStackForTest targeted result. Note that test0 and main
	// functions don't have package name `main` in them!! That's how func name is
	// calculated in our debug pkg.
	outputForTest = `    /home/god/go/src/github.com/lainio/err2/err2.go:107: err2.Handle
    /usr/local/go/src/runtime/panic.go:838: panic
    /home/god/go/src/github.com/lainio/err2/err2.go:214: err2.Returnw
    /usr/local/go/src/runtime/panic.go:838: panic
    /home/god/go/src/github.com/lainio/ic/main.go:18: test0
    /home/god/go/src/github.com/lainio/ic/main.go:74: main
`

	outputForTestLvl2 = `    /home/god/go/src/github.com/lainio/ic/main.go:74: main
    /home/god/go/src/github.com/lainio/ic/main.go:18: test0 STACK
    /usr/local/go/src/runtime/panic.go:838: panic STACK
    /home/god/go/src/github.com/lainio/err2/err2.go:214: err2.Returnw STACK
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

	inputPipelineStack = `goroutine 1 [running]:
runtime/debug.Stack()
        /usr/local/go/src/runtime/debug/stack.go:24 +0x64
github.com/lainio/err2/internal/debug.FuncName({{0x0, 0x0}, {0x12f04a, 0x6}, 0xffffffffffffffff, 0x0, {0x0, 0x0, 0x0}})
        /home/parallels/go/src/github.com/lainio/err2/internal/debug/debug.go:162 +0x44
github.com/lainio/err2/internal/handler.doBuildFormatStr(0x4000121b58?, 0x9bc5c?)
        /home/parallels/go/src/github.com/lainio/err2/internal/handler/handler.go:317 +0x7c
github.com/lainio/err2/internal/handler.buildFormatStr(...)
        /home/parallels/go/src/github.com/lainio/err2/internal/handler/handler.go:305
github.com/lainio/err2/internal/handler.PreProcess(0x4000121d88, 0x4000121ba0, {0x0, 0x0, 0x0})
        /home/parallels/go/src/github.com/lainio/err2/internal/handler/handler.go:280 +0xf8
github.com/lainio/err2.Handle(0x4000121d88, {0x0, 0x0, 0x0})
        /home/parallels/go/src/github.com/lainio/err2/err2.go:103 +0xd4
panic({0x115f20?, 0x4000036660?})
        /usr/local/go/src/runtime/panic.go:770 +0x124
github.com/lainio/err2/try.To(...)
        /home/parallels/go/src/github.com/lainio/err2/try/try.go:82
github.com/lainio/err2/try.To1[...](...)
        /home/parallels/go/src/github.com/lainio/err2/try/try.go:97
main.CopyFile({0x12f23c?, 0x1609c?}, {0x132cef, 0x17})
        /home/parallels/go/src/github.com/lainio/err2/samples/main-play.go:29 +0x254
main.doMain()
        /home/parallels/go/src/github.com/lainio/err2/samples/main-play.go:159 +0x68
main.doDoMain(...)
        /home/parallels/go/src/github.com/lainio/err2/samples/main-play.go:143
main.doPlayMain()
        /home/parallels/go/src/github.com/lainio/err2/samples/main-play.go:136 +0x68
main.main()
        /home/parallels/go/src/github.com/lainio/err2/samples/main.go:38 +0x15c
`
)
