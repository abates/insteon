package insteon

import (
	"fmt"
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
	return fmt.Sprintf("%v: need %d bytes got %d", be.Cause, be.Need, be.Got)
}

// Error is used only when something failed that needs to bubble up
// the location in code where the error occurred.
type Error struct {
	Cause error         // the underlying cause of the error
	Frame runtime.Frame // the runtime frame of the occurrance
}

// IsError will determine if `check` is wrapping an underlying error.
// If so, the underlying error is compared to `err`.
func IsError(check, err error) bool {
	switch e := err.(type) {
	case *Error:
		return e.Cause == check
	case *BufError:
		return e.Cause == check
	}
	return check == err
}

// Error indicates the underlying cause of the error as well as the file and line that the error occurred
func (e *Error) Error() string {
	return fmt.Sprintf("%s:%d in %q: %s", path.Base(e.Frame.File), e.Frame.Line, e.Frame.Function, e.Cause.Error())
}

// TraceError generates an Error and records the runtime stack frame
func TraceError(cause error) error {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc)
	frame, _ := frames.Next()

	return &Error{
		Cause: cause,
		Frame: frame,
	}
}
