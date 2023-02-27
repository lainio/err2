package try_test

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/lainio/err2/internal/helper"
	"github.com/lainio/err2/try"
)

const dataFile = "./try.go"

func Benchmark_CopyBufferStd(b *testing.B) {
	all, err := os.ReadFile(dataFile)
	helper.Requiref(b, err == nil, "error: %v", err)
	helper.Require(b, all != nil)

	buf := make([]byte, 4)
	dst := bufio.NewWriter(bytes.NewBuffer(make([]byte, 0, len(all))))
	src := bytes.NewReader(all)
	for n := 0; n < b.N; n++ {
		try.To1(io.CopyBuffer(dst, src, buf))
	}
}

func Benchmark_CopyBufferOur(b *testing.B) {
	all, err := os.ReadFile(dataFile)
	helper.Requiref(b, err == nil, "error: %v", err)
	helper.Require(b, all != nil)

	tmp := make([]byte, 4)
	dst := bufio.NewWriter(bytes.NewBuffer(make([]byte, 0, len(all))))
	src := bytes.NewReader(all)
	for n := 0; n < b.N; n++ {
		for eof, n := try.IsEOF1(src.Read(tmp)); !eof; eof, n = try.IsEOF1(src.Read(tmp)) {
			try.To1(dst.Write(tmp[:n]))
		}
	}
}