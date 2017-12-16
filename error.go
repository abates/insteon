package insteon

import (
	"bytes"
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

type AggregateError struct {
	Errors []error
}

func NewAggregateError() *AggregateError {
	return &AggregateError{}
}

func (ae *AggregateError) Len() int {
	return len(ae.Errors)
}

func (ae *AggregateError) Append(err error) {
	if err != nil {
		ae.Errors = append(ae.Errors, err)
	}
}

func (ae *AggregateError) Error() string {
	var buf bytes.Buffer
	for _, err := range ae.Errors {
		buf.WriteString(fmt.Sprintf("%s\n", err.Error()))
	}
	return buf.String()
}
