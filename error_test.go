package insteon

import (
	"fmt"
	"runtime"
	"testing"
)

func TestBufError(t *testing.T) {
	tests := []struct {
		cause        error
		need         int
		got          int
		expectedFmts []string
		expectedArgs []interface{}
	}{
		{nil, 10, 0, []string{"%sneed %d bytes got %d"}, []interface{}{"", 10, 0}},
		{fmt.Errorf("Foo"), 10, 0, []string{"%v: ", "%sneed %d bytes got %d"}, []interface{}{fmt.Errorf("Foo"), "", 10, 0}},
	}

	for i, test := range tests {
		err := newBufError(test.cause, test.need, test.got)
		if msg, fail := testStr(err.Error, test.expectedFmts, test.expectedArgs); fail {
			t.Errorf("tests[%d]: %s", i, msg)
		}
	}
}

func TestError(t *testing.T) {
	err := &Error{
		Cause: fmt.Errorf("Foo"),
		Frame: runtime.Frame{File: "/foo/bar/run.go", Line: 10, Function: "Woops"},
	}

	if msg, failed := testStr(err.Error, []string{"%s:%d in %q: %s"}, []interface{}{"run.go", 10, "Woops", "Foo"}); failed {
		t.Errorf("%s", msg)
	}
}
