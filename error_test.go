// Copyright 2018 Andrew Bates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package insteon

import (
	"errors"
	"runtime"
	"testing"
)

func TestBufError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc     string
		input    error
		expected string
	}{
		{"no cause", newBufError(nil, 10, 20), "need 10 bytes got 20"},
		{"with cause", newBufError(errors.New("Foo"), 10, 20), "Foo: need 10 bytes got 20"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if test.input.Error() != test.expected {
				t.Errorf("got %v expected %v", test.input.Error(), test.expected)
			}
		})
	}
}

func TestError(t *testing.T) {
	t.Parallel()
	err := &traceError{
		Cause: errors.New("Foo"),
		Frame: runtime.Frame{File: "/foo/bar/run.go", Line: 10, Function: "Woops"},
	}

	if err.Error() == "" {
		t.Error("Expected non-empty string")
	}
}

func TestIsError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc     string
		cause    error
		check    error
		expected bool
	}{
		{"traceError matches", &traceError{Cause: ErrBufferTooShort}, ErrBufferTooShort, true},
		{"traceError mismatch", &traceError{Cause: ErrReadTimeout}, ErrBufferTooShort, false},
		{"bufError match", &BufError{Cause: ErrBufferTooShort}, ErrBufferTooShort, true},
		{"bufError mismatch", &BufError{Cause: ErrReadTimeout}, ErrBufferTooShort, false},
		{"error match", ErrReadTimeout, ErrReadTimeout, true},
		{"error mismatch", ErrReadTimeout, ErrBufferTooShort, false},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if isError(test.cause, test.check) != test.expected {
				switch e := test.cause.(type) {
				case *traceError:
					t.Logf("%v == %v ? %v -- %v", e.Cause, test.check, e.Cause == test.check, isError(test.cause, test.check))
				case *BufError:
					t.Logf("%v == %v ? %v -- %v", e.Cause, test.check, e.Cause == test.check, isError(test.cause, test.check))
				}
				t.Errorf("got %v, expected %v ", isError(test.cause, test.check), test.expected)
			}
		})
	}
}

func TestTraceError(t *testing.T) {
	t.Parallel()
	err := newTraceError(ErrBufferTooShort)
	if _, ok := err.(*traceError); !ok {
		t.Errorf("expected *Error got %T", err)
	}
}
