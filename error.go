package insteon

import (
	"fmt"
	"path"
	"runtime"
)

type Error struct {
	Cause error
	Frame runtime.Frame
}

func IsError(check, err error) bool {
	if e, ok := err.(*Error); ok {
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
