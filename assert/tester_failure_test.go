package assert

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/lainio/err2/internal/expect"
)

type reportingTester struct {
	*namedTester
}

func (t *reportingTester) FailNow() {
	fmt.Fprintln(os.Stderr, "failed tester:", t.name)
	t.TB.FailNow()
}

func TestTesterFailures(t *testing.T) {
	executable, err := os.Executable()
	expect.Thatf(t, err == nil, "test executable: %v", err)
	for _, mode := range []string{"closure", "legacy", "restored"} {
		t.Run(mode, func(t *testing.T) {
			for _, kind := range []string{"string", "error", "runtime", "other", "assertion", "sentinel", "failnow"} {
				t.Run(kind, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
					defer cancel()
					cmd := exec.CommandContext(ctx, executable, "-test.run=^TestTesterFailureProcess$", "-test.v", "-test.count=1")
					cmd.Env = append(os.Environ(), "ERR2_TESTER_CHILD="+mode+"/"+kind)
					out, err := cmd.CombinedOutput()
					output := string(out)
					var exitErr *exec.ExitError
					expect.Thatf(t, errors.As(err, &exitErr), "expected test failure: %v\n%s", err, output)
					expect.Equal(t, exitErr.ExitCode(), 1, "unexpected exit\n%s", output)
					owner, other := "inner", "outer"
					if mode != "closure" {
						owner, other = "outer", "inner"
					}
					markers := []string{"failed tester: " + owner, "all scopes cleared", "--- FAIL: TestTesterFailureProcess"}
					if mode != "legacy" {
						markers = append(markers, "outer restored")
					}
					for _, marker := range markers {
						expect.Thatf(t, strings.Contains(output, marker), "missing %q\n%s", marker, output)
					}
					expect.Thatf(t, !strings.Contains(output, "failed tester: "+other), "wrong failure owner\n%s", output)
					expect.Thatf(t, !strings.Contains(output, "continued after failure"), "FailNow did not exit\n%s", output)
					switch kind {
					case "string", "error", "runtime", "other":
						location := regexp.MustCompile(`panic-site=([^\r\n]+)`).FindStringSubmatch(output)
						expect.Equal(t, len(location), 2, "missing panic location\n%s", output)
						if kind == "runtime" {
							// The legacy runtime.Error offsets can label the final
							// message with a runtime frame. Preserve those offsets,
							// and verify the original source is present in the trace.
							expect.Thatf(t, strings.Contains(output, location[1]+": assert.testerFailureSite"), "missing runtime panic source\n%s", output)
						} else {
							expect.Thatf(t, strings.Contains(output, location[1]+": assertion catching: "), "wrong panic source line\n%s", output)
						}
						messages := map[string]string{
							"string":  "assertion catching: tester string panic",
							"error":   "assertion catching: tester error panic",
							"runtime": "assertion catching: runtime error: index out of range",
							"other":   "assertion catching: panic: 42",
						}
						expect.Thatf(t, strings.Contains(output, messages[kind]), "wrong panic message\n%s", output)
					case "assertion", "sentinel":
						expect.Thatf(t, strings.Contains(output, "tester assertion"), "missing assertion message\n%s", output)
					}
				})
			}
		})
	}
}

// This subprocess intentionally fails through the real testing harness. The
// wrappers make the selected TB observable without intercepting runtime.Goexit.
func TestTesterFailureProcess(t *testing.T) {
	mode := os.Getenv("ERR2_TESTER_CHILD")
	if mode == "" {
		t.Skip("only run as a subprocess")
	}
	parts := strings.Split(mode, "/")
	outer := &reportingTester{&namedTester{TB: t, name: "outer"}}
	inner := &reportingTester{&namedTester{TB: t, name: "inner"}}
	defer func() {
		if tester() != nil {
			t.Error("outer cleanup retained a tester")
			return
		}
		fmt.Fprintln(os.Stderr, "all scopes cleared")
	}()
	if parts[0] == "legacy" {
		// Legacy PopTester handles one registration; it does not restore nesting.
		PushTester(outer)
		defer PopTester()
		testerFailureSite(parts[1])
	} else {
		defer PushTester(outer)()
		func() {
			defer func() {
				if tester() != outer {
					t.Error("inner cleanup did not restore its context")
					return
				}
				fmt.Fprintln(os.Stderr, "outer restored")
			}()
			if parts[0] == "restored" {
				undo := PushTester(inner)
				undo()
			} else {
				defer PushTester(inner)()
			}
			testerFailureSite(parts[1])
		}()
	}
	t.Fatal("continued after failure")
}

func testerFailureSite(kind string) {
	switch kind {
	case "string":
		recordTesterPanicSite()
		panic("tester string panic")
	case "error":
		recordTesterPanicSite()
		panic(errors.New("tester error panic"))
	case "runtime":
		// following don't give you warning but throws:
		// data := []byte{0}; _ = data[len(data)]
		values := []int{0}
		recordTesterPanicSite()
		_ = values[len(values)]
	case "other":
		recordTesterPanicSite()
		panic(42)
	case "assertion":
		That(false, "tester assertion")
	case "sentinel":
		That(false, errors.New("tester assertion"))
	case "failnow":
		tester().FailNow()
	}
}

func recordTesterPanicSite() {
	_, file, line, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "panic-site=%s:%d\n", filepath.Base(file), line+1)
}
