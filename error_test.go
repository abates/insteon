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
	"fmt"
	"runtime"
	"testing"
)

func TestBufError(t *testing.T) {
	tests := []struct {
		input    error
		expected string
	}{
		{newBufError(nil, 10, 20), "need 10 bytes got 20"},
		{newBufError(fmt.Errorf("Foo"), 10, 20), "Foo: need 10 bytes got 20"},
	}

	for i, test := range tests {
		if test.input.Error() != test.expected {
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, test.input.Error())
		}
	}
}

func TestError(t *testing.T) {
	err := &traceError{
		Cause: fmt.Errorf("Foo"),
		Frame: runtime.Frame{File: "/foo/bar/run.go", Line: 10, Function: "Woops"},
	}

	if err.Error() == "" {
		t.Errorf("Expected non-empty string")
	}
}

func TestIsError(t *testing.T) {
	tests := []struct {
		cause    error
		check    error
		expected bool
	}{
		{&traceError{Cause: ErrBufferTooShort}, ErrBufferTooShort, true},
		{&traceError{Cause: ErrReadTimeout}, ErrBufferTooShort, false},
		{&BufError{Cause: ErrBufferTooShort}, ErrBufferTooShort, true},
		{&BufError{Cause: ErrReadTimeout}, ErrBufferTooShort, false},
		{ErrReadTimeout, ErrReadTimeout, true},
		{ErrReadTimeout, ErrBufferTooShort, false},
	}

	for i, test := range tests {
		if isError(test.cause, test.check) != test.expected {
			switch e := test.cause.(type) {
			case *traceError:
				t.Logf("tests[%d] %v == %v ? %v -- %v", i, e.Cause, test.check, e.Cause == test.check, isError(test.cause, test.check))
			case *BufError:
				t.Logf("tests[%d] %v == %v ? %v -- %v", i, e.Cause, test.check, e.Cause == test.check, isError(test.cause, test.check))
			}
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, isError(test.cause, test.check))
		}
	}
}

func TestTraceError(t *testing.T) {
	err := newTraceError(ErrBufferTooShort)
	if _, ok := err.(*traceError); !ok {
		t.Errorf("expected *Error got %T", err)
	}
}
