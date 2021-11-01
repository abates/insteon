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
	"io"
	"log"
	"os"
)

var (
	Log      = log.New(os.Stderr, "", log.LstdFlags)
	LogDebug = log.New(io.Discard, "DEBUG ", log.LstdFlags)
	LogTrace = log.New(io.Discard, "TRACE ", log.LstdFlags|log.Llongfile)
)

// LogLevel indicates verbosity of logging
type LogLevel int

func (ll LogLevel) String() string {
	switch ll {
	case LevelNone:
		return "NONE"
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	case LevelTrace:
		return "TRACE"
	}
	return ""
}

// Set sets the LogLevel to the matching input string.  This function
// satisfies the Set function for the flag.Value interface
func (ll *LogLevel) Set(s string) (err error) {
	switch s {
	case "none":
	case "info":
		(*ll) = LevelInfo
	case "debug":
		(*ll) = LevelDebug
	case "trace":
		(*ll) = LevelTrace
	default:
		err = errors.New("valid values {none|info|debug|trace}")
	}
	return err
}

// Get returns the underlying LogLevel value in order to satisfy the
// flag.Value interface
func (ll *LogLevel) Get() interface{} {
	return LogLevel(*ll)
}

const (
	// LevelNone - log nothing to Stderr
	LevelNone = LogLevel(iota)

	// LevelInfo - log normal information (warnings) to Stderr
	LevelInfo

	// LevelDebug - log debug information (used for development and troubleshooting)
	LevelDebug

	// LevelTrace - log everything, including I/O
	LevelTrace
)

func SetLogLevel(ll LogLevel, writer io.Writer) {
	Log.SetOutput(io.Discard)
	LogDebug.SetOutput(io.Discard)
	LogTrace.SetOutput(io.Discard)

	if LevelInfo <= ll {
		Log.SetOutput(writer)
	}

	if LevelDebug <= ll {
		LogDebug.SetOutput(writer)
	}

	if LevelTrace <= ll {
		LogTrace.SetOutput(writer)
	}
}
