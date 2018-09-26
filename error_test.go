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
	err := &TraceError{
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
		{&TraceError{Cause: ErrBufferTooShort}, ErrBufferTooShort, true},
		{&TraceError{Cause: ErrReadTimeout}, ErrBufferTooShort, false},
		{&BufError{Cause: ErrBufferTooShort}, ErrBufferTooShort, true},
		{&BufError{Cause: ErrReadTimeout}, ErrBufferTooShort, false},
		{ErrReadTimeout, ErrReadTimeout, true},
		{ErrReadTimeout, ErrBufferTooShort, false},
	}

	for i, test := range tests {
		if IsError(test.cause, test.check) != test.expected {
			switch e := test.cause.(type) {
			case *TraceError:
				t.Logf("tests[%d] %v == %v ? %v -- %v", i, e.Cause, test.check, e.Cause == test.check, IsError(test.cause, test.check))
			case *BufError:
				t.Logf("tests[%d] %v == %v ? %v -- %v", i, e.Cause, test.check, e.Cause == test.check, IsError(test.cause, test.check))
			}
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, IsError(test.cause, test.check))
		}
	}
}

func TestTraceError(t *testing.T) {
	err := NewTraceError(ErrBufferTooShort)
	if _, ok := err.(*TraceError); !ok {
		t.Errorf("expected *Error got %T", err)
	}
}
