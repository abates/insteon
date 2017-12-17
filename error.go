package insteon

import (
	"fmt"
	"path"
	"runtime"
)

type BufError struct {
	Cause error
	Need  int
	Got   int
}

func newBufError(cause error, need int, got int) *BufError {
	return &BufError{Cause: cause, Need: need, Got: got}
}

func (be *BufError) Error() string {
	return fmt.Sprintf("%v: need %d bytes got %d", be.Cause, be.Need, be.Got)
}

type Error struct {
	Cause error
	Frame runtime.Frame
}

func IsError(check, err error) bool {
	switch e := err.(type) {
	case *Error:
		return e.Cause == check
	case *BufError:
		return e.Cause == check
	}
	return check == err
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%d in %q: %s", path.Base(e.Frame.File), e.Frame.Line, e.Frame.Function, e.Cause.Error())
}

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
