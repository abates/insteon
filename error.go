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
	"path"
	"runtime"
)

// BufError has information to indicate a buffer error
type BufError struct {
	Cause error // the underlying error
	Need  int   // the number of bytes required
	Got   int   // the length of the supplied buffer
}

func newBufError(cause error, need int, got int) *BufError {
	return &BufError{Cause: cause, Need: need, Got: got}
}

// Error will indicate what caused a buffer error to occur
func (be *BufError) Error() string {
	cause := ""
	if be.Cause != nil {
		cause = sprintf("%v: ", be.Cause)
	}
	return sprintf("%sneed %d bytes got %d", cause, be.Need, be.Got)
}

// traceError is used only when something failed that needs to bubble up
// the location in code where the error occurred.
type traceError struct {
	Cause error         // the underlying cause of the error
	Frame runtime.Frame // the runtime frame of the occurrence
}

// isError will determine if `check` is wrapping an underlying error.
// If so, the underlying error is compared to `err`.
func isError(check, err error) bool {
	switch e := check.(type) {
	case *traceError:
		check = e.Cause
	case *BufError:
		check = e.Cause
	}
	return check == err
}

// Error indicates the underlying cause of the error as well as the file and line that the error occurred
func (e *traceError) Error() string {
	return sprintf("%s:%d in %q: %s", path.Base(e.Frame.File), e.Frame.Line, e.Frame.Function, e.Cause.Error())
}

// newTraceError generates an Error and records the runtime stack frame
func newTraceError(cause error) error {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc)
	frame, _ := frames.Next()

	return &traceError{
		Cause: cause,
		Frame: frame,
	}
}
