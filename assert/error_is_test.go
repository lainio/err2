package assert_test

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/lainio/err2/assert"
)

func TestErrorIsAndErrorIsNot(t *testing.T) {
	sentinel := errors.New("sentinel")
	other := errors.New("other")
	wrapped := fmt.Errorf("wrapped: %w", sentinel)
	tests := []struct {
		name        string
		err, target error
		matches     bool
	}{
		{"both nil", nil, nil, true},
		{"nil error", nil, sentinel, false},
		{"nil target", sentinel, nil, false},
		{"same error", sentinel, sentinel, true},
		{"different errors", sentinel, other, false},
		{"same text", sentinel, errors.New("sentinel"), false},
		{"wrapped match", wrapped, sentinel, true},
		{"wrapped mismatch", wrapped, other, false},
		{"wrapped nil target", wrapped, nil, false},
		{"wrapped target", sentinel, wrapped, false},
	}
	helpers := []struct {
		name      string
		assert    func(error, error, ...any)
		wantMatch bool
		message   string
	}{
		{"ErrorIs", assert.ErrorIs, true, "assertion failure: not equal"},
		{"ErrorIsNot", assert.ErrorIsNot, false, "assertion failure: equal"},
	}

	for _, helper := range helpers {
		t.Run(helper.name, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					defer assert.PushAsserter(assert.Plain)()
					defer func() {
						r := recover()
						if tt.matches == helper.wantMatch {
							if r != nil {
								t.Fatalf("unexpected panic: %T: %v", r, r)
							}
							return
						}
						err, ok := r.(error)
						if !ok {
							t.Fatalf("got panic %T: %v, want assertion error", r, r)
						}
						if _, ok := err.(runtime.Error); ok {
							t.Fatalf("got runtime panic: %v", err)
						}
						want := fmt.Sprintf(helper.message+": got '%v', want '%v'",
							tt.err, tt.target)
						if err.Error() != want {
							t.Errorf("got %q, want %q", err, want)
						}
					}()
					helper.assert(tt.err, tt.target)
				})
			}
		})
	}
}

func TestErrorIsReporting(t *testing.T) {
	sentinel := errors.New("sentinel")
	reported := errors.New("reported")
	tests := []struct {
		name        string
		assert      func(error, error, ...any)
		err, target error
		message     string
	}{
		{"ErrorIs nil error", assert.ErrorIs, nil, sentinel,
			"assertion failure: not equal: got '<nil>', want 'sentinel'"},
		{"ErrorIs nil target", assert.ErrorIs, sentinel, nil,
			"assertion failure: not equal: got 'sentinel', want '<nil>'"},
		{"ErrorIsNot both nil", assert.ErrorIsNot, nil, nil,
			"assertion failure: equal: got '<nil>', want '<nil>'"},
	}
	modes := []struct {
		name     string
		asserter assert.Asserter
	}{
		{"Plain", assert.Plain},
		{"Production", assert.Production},
		{"Debug", assert.Debug},
	}
	arguments := []struct {
		name string
		args []any
	}{
		{"default", nil},
		{"formatted message", []any{"context %d", 42}},
		{"sentinel", []any{reported}},
	}

	for _, tt := range tests {
		for _, mode := range modes {
			for _, arg := range arguments {
				t.Run(tt.name+"/"+mode.name+"/"+arg.name, func(t *testing.T) {
					defer assert.PushAsserter(mode.asserter)()
					defer func() {
						r := recover()
						want := tt.message
						if len(arg.args) > 0 {
							extra := "context 42"
							if arg.name == "sentinel" {
								extra = reported.Error()
							}
							if mode.asserter == assert.Plain {
								want = extra
							} else {
								want += ": " + extra
							}
						}
						if mode.asserter == assert.Debug && arg.name != "sentinel" {
							if got, ok := r.(string); !ok || got != want {
								t.Errorf("got panic %T: %v, want string %q", r, r, want)
							}
							return
						}
						err, ok := r.(error)
						if !ok {
							t.Fatalf("got panic %T: %v, want assertion error", r, r)
						}
						if _, ok := err.(runtime.Error); ok {
							t.Fatalf("got runtime panic: %v", err)
						}
						if mode.asserter == assert.Production {
							if !strings.Contains(err.Error(), "error_is_test.go:") ||
								!strings.Contains(err.Error(), "TestErrorIsReporting.") ||
								!strings.HasSuffix(err.Error(), ": "+want) {
								t.Errorf("got %q, want caller info followed by %q", err, want)
							}
						} else if err.Error() != want {
							t.Errorf("got %q, want %q", err, want)
						}
						if arg.name == "sentinel" {
							if !errors.Is(err, reported) {
								t.Errorf("got %v, want sentinel in error chain", err)
							}
							if mode.asserter == assert.Plain && err != reported {
								t.Errorf("got %v, want original sentinel", err)
							}
						}
					}()
					tt.assert(tt.err, tt.target, arg.args...)
				})
			}
		}
	}
}
