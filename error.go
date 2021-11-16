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
	"fmt"
	"path"
	"runtime"
)

var (
	// ErrBufferTooShort indicates a buffer underrun when unmarshalling data
	ErrBufferTooShort = errors.New("Buffer is too short")

	// ErrReadTimeout indicates the timeout period expired while waiting for
	// a specific message
	ErrReadTimeout = errors.New("Read Timeout")

	// ErrAckTimeout indicates the timeout period expired while waiting for
	// ack message for a previously sent Direct message
	ErrAckTimeout = errors.New("Ack Timeout")

	// ErrWriteTimeout indicates the timeout period expired while waiting to
	// write a message
	ErrWriteTimeout = errors.New("Write Timeout")

	// ErrNotSupported indicates some feature (namely an updateable All-Link database) is
	// not supported by the underlying Insteon device
	ErrNotSupported = errors.New("Feature is not supported by the device")

	// ErrNotImplemented indicates that a device function has not yet been implemented
	ErrNotImplemented = errors.New("Command is not yet implemented")

	// ErrNotLinkable indicates a linking function was requested on a non-linkable device
	ErrNotLinkable = errors.New("Device is not remotely linkable")

	// ErrUnknownCommand is returned by the device (as a Nak) in response to an unknown command byte
	ErrUnknownCommand = errors.New("Unknown command")

	// ErrUnknown is returned by a connection when a NAK occurred but the error code
	// is not known
	ErrUnknown = errors.New("Device returned unknown error")

	// ErrPreNak is returned by I2Cs devices (this error condition is not documented)
	ErrPreNak = errors.New("Database search took too long")

	// ErrAddrFormat is returned when unmarshalling an address from text and the
	// text is in an unsupported format
	ErrAddrFormat = errors.New("address format is xx.xx.xx (digits in hex)")

	// ErrInvalidMemAddress indicates a link record memory address is invalid
	ErrInvalidMemAddress = errors.New("Invalid memory address")

	// ErrVersion is returned when an engine version value is not known
	ErrVersion = errors.New("Unknown Insteon Engine Version")
)

// traceError is used only when something failed that needs to bubble up
// the location in code where the error occurred.
type traceError struct {
	Cause error         // the underlying cause of the error
	Frame runtime.Frame // the runtime frame of the occurrence
}

func (e *traceError) Unwrap() error {
	return e.Cause
}

// Error indicates the underlying cause of the error as well as the file and line that the error occurred
func (e *traceError) Error() string {
	return fmt.Sprintf("%s:%d in %q: %s", path.Base(e.Frame.File), e.Frame.Line, e.Frame.Function, e.Cause.Error())
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
