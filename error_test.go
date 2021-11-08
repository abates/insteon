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

func TestError(t *testing.T) {
	err := &traceError{
		Cause: errors.New("Foo"),
		Frame: runtime.Frame{File: "/foo/bar/run.go", Line: 10, Function: "Woops"},
	}

	if err.Error() == "" {
		t.Error("Expected non-empty string")
	}
}

func TestTraceError(t *testing.T) {
	err := newTraceError(ErrBufferTooShort)
	if _, ok := err.(*traceError); !ok {
		t.Errorf("expected *Error got %T", err)
	}

	if !errors.Is(err, ErrBufferTooShort) {
		t.Errorf("expected TraceError to wrap ErrBufferTooShort")
	}
}
