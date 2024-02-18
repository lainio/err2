package try_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/lainio/err2/internal/test"
	"github.com/lainio/err2/try"
)

const dataFile = "./try.go"

func Benchmark_CopyBufferMy(b *testing.B) {
	all, err := os.ReadFile(dataFile)
	test.Requiref(b, err == nil, "error: %v", err)
	test.Require(b, all != nil)

	buf := make([]byte, 4)
	dst := bufio.NewWriter(bytes.NewBuffer(make([]byte, 0, len(all))))
	src := bytes.NewReader(all)
	for n := 0; n < b.N; n++ {
		try.To1(myCopyBuffer(dst, src, buf))
	}
}

func Benchmark_CopyBufferStd(b *testing.B) {
	all, err := os.ReadFile(dataFile)
	test.Requiref(b, err == nil, "error: %v", err)
	test.Require(b, all != nil)

	buf := make([]byte, 4)
	dst := bufio.NewWriter(bytes.NewBuffer(make([]byte, 0, len(all))))
	src := bytes.NewReader(all)
	for n := 0; n < b.N; n++ {
		try.To1(io.CopyBuffer(dst, src, buf))
	}
}

func Benchmark_CopyBufferOur(b *testing.B) {
	all, err := os.ReadFile(dataFile)
	test.Requiref(b, err == nil, "error: %v", err)
	test.Require(b, all != nil)

	tmp := make([]byte, 4)
	dst := bufio.NewWriter(bytes.NewBuffer(make([]byte, 0, len(all))))
	src := bytes.NewReader(all)
	for n := 0; n < b.N; n++ {
		for eof, n := try.IsEOF1(src.Read(tmp)); !eof; eof, n = try.IsEOF1(src.Read(tmp)) {
			try.To1(dst.Write(tmp[:n]))
		}
	}
}

// myCopyBuffer is copy/paste from Go std lib to remove noice and measure only a
// loop
func myCopyBuffer(dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
