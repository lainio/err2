package str_test

import (
	"testing"

	"github.com/lainio/err2/internal/helper"
	"github.com/lainio/err2/internal/str"
)

const (
	camelStr = "BenchmarkRecursionWithOldErrorIfCheckAnd_Defer"
)

func BenchmarkCamelRegexp(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = str.CamelRegexp(camelStr)
	}
}

func BenchmarkDecamel(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = str.Decamel(camelStr)
	}
}

func TestDecamel(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple", args{"CamelString"}, "camel string"},
		{"underscore", args{"CamelString_error"}, "camel string error"},
		{"our contant", args{camelStr}, "benchmark recursion with old error if check and defer"},
		{"number", args{"CamelString2Testing"}, "camel string2 testing"},
		{"acronym", args{"ARMCamelString"}, "armcamel string"},
		{"acronym at end", args{"archIsARM"}, "arch is arm"},
		{"simple method", args{"(*DIDAgent).AssertWallet"}, "didagent assert wallet"},
		{"package name and simple method", args{"ssi.(*DIDAgent).AssertWallet"}, "ssi didagent assert wallet"},
		{"simple method and anonym", args{"(*DIDAgent).AssertWallet.Func1"}, "didagent assert wallet func1"},
		{"complex method and anonym", args{"(**DIDAgent).AssertWallet.Func1"}, "didagent assert wallet func1"},
		{"unnatural method and anonym", args{"(**DIDAgent)...AssertWallet...Func1"}, "didagent assert wallet func1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := str.Decamel(tt.args.s)
			helper.Requiref(t, got == tt.want, "got: %v, want: %v", got, tt.want)
		})
	}
}