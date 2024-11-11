package str_test

import (
	"testing"

	"github.com/lainio/err2/internal/except"
	"github.com/lainio/err2/internal/str"
)

const (
	camelStr = "BenchmarkRecursionWithOldErrorIfCheckAnd_Defer"
)

func BenchmarkDecamelRegexp(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = str.DecamelRegexp(camelStr)
	}
}

func BenchmarkDecamel(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = str.Decamel(camelStr)
	}
}

func BenchmarkDecamelRmTryPrefix(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = str.DecamelRmTryPrefix(camelStr)
	}
}

func TestCamel(t *testing.T) {
	t.Parallel()
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple", args{"CamelString"}, "camel string"},
		{"number", args{"CamelString2Testing"}, "camel string2 testing"},
		{"acronym", args{"ARMCamelString"}, "armcamel string"},
		{"acronym at end", args{"archIsARM"}, "arch is arm"},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := str.DecamelRegexp(tt.args.s)
			except.Equal(t, got, tt.want)
		})
	}
}

func TestDecamel(t *testing.T) {
	t.Parallel()
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
		{
			"our contant",
			args{camelStr},
			"benchmark recursion with old error if check and defer",
		},
		{"number", args{"CamelString2Testing"}, "camel string2 testing"},
		{"acronym", args{"ARMCamelString"}, "armcamel string"},
		{"acronym at end", args{"archIsARM"}, "arch is arm"},
		{
			"simple method",
			args{"(*DIDAgent).AssertWallet"},
			"didagent assert wallet",
		},
		{
			"package name and simple method",
			args{"ssi.(*DIDAgent).CreateWallet"},
			"ssi: didagent create wallet",
		},
		{
			"simple method and anonym",
			args{"(*DIDAgent).AssertWallet.Func1"},
			"didagent assert wallet: func1",
		},
		{
			"complex method and anonym",
			args{"(**DIDAgent).AssertWallet.Func1"},
			"didagent assert wallet: func1",
		},
		{
			"unnatural method and anonym",
			args{"(**DIDAgent)...AssertWallet...Func1"},
			"didagent assert wallet: func1",
		},
		{"from spf13 cobra", args{"bot.glob..func5"}, "bot: glob: func5"},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := str.Decamel(tt.args.s)
			except.Equal(t, got, tt.want)
		})
	}
}

func TestDecamelRmTryPrefix(t *testing.T) {
	t.Parallel()
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple", args{"CamelString"}, "camel string"},
		{"simple try", args{"TryCamelString"}, "camel string"},
		{"underscore", args{"CamelString_error"}, "camel string error"},
		{"underscore and try", args{"TryCamelString_error"}, "camel string error"},
		{
			"our contant",
			args{camelStr},
			"benchmark recursion with old error if check and defer",
		},
		{"number", args{"CamelString2Testing"}, "camel string2 testing"},
		{"acronym", args{"ARMCamelString"}, "armcamel string"},
		{"acronym and try at END so it left", args{"ARMCamelStringTry"}, "armcamel string try"},
		{"acronym and try", args{"TryARMCamelString"}, "armcamel string"},
		{"acronym at end", args{"archIsARM"}, "arch is arm"},
		{
			"simple method",
			args{"(*DIDAgent).AssertWallet"},
			"didagent assert wallet",
		},
		{
			"package name and simple method",
			args{"ssi.(*DIDAgent).CreateWallet"},
			"ssi: didagent create wallet",
		},
		{
			"package name and simple method and Function start try",
			args{"ssi.(*DIDAgent).TryCreateWallet"},
			"ssi: didagent create wallet",
		},
		{
			"simple method and anonym",
			args{"(*DIDAgent).AssertWallet.Func1"},
			"didagent assert wallet: func1",
		},
		{
			"complex method and anonym",
			args{"(**DIDAgent).AssertWallet.Func1"},
			"didagent assert wallet: func1",
		},
		{
			"complex method and anonym AND try",
			args{"(**DIDAgent).TryAssertWallet.Func1"},
			"didagent assert wallet: func1",
		},
		{
			"unnatural method and anonym",
			args{"(**DIDAgent)...AssertWallet...Func1"},
			"didagent assert wallet: func1",
		},
		{
			"unnatural method and anonym AND try",
			args{"(**DIDAgent)...TryAssertWallet...TryFunc1"},
			"didagent assert wallet: func1",
		},
		{"from spf13 cobra", args{"bot.glob..func5"}, "bot: glob: func5"},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := str.DecamelRmTryPrefix(tt.args.s)
			except.Equal(t, got, tt.want)
		})
	}
}
