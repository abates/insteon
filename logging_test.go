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
	"bytes"
	"io"
	"log"
	"testing"
)

func TestSetLogLevel(t *testing.T) {
	tests := []struct {
		name   string
		input  LogLevel
		writer io.Writer
		want   []*log.Logger
	}{
		{"none", LevelNone, io.Discard, []*log.Logger{Log, LogDebug, LogTrace}},
		{"info", LevelInfo, bytes.NewBuffer(nil), []*log.Logger{Log}},
		{"debug", LevelDebug, bytes.NewBuffer(nil), []*log.Logger{Log, LogDebug}},
		{"trace", LevelTrace, bytes.NewBuffer(nil), []*log.Logger{Log, LogDebug, LogTrace}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			SetLogLevel(test.input, test.writer)
			for i, logger := range test.want {
				if logger.Writer() != test.writer {
					t.Errorf("Logger %d has an incorrect output writer", i)
				}
			}
		})
	}
}

func TestLogLevelSet(t *testing.T) {
	tests := []struct {
		input   string
		want    LogLevel
		wantErr bool
	}{
		{"none", LevelNone, false},
		{"info", LevelInfo, false},
		{"debug", LevelDebug, false},
		{"trace", LevelTrace, false},
		{"foo", LevelNone, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := LogLevel(0)
			gotErr := got.Set(test.input)
			if test.wantErr {
				if gotErr == nil {
					t.Errorf("want error got none")
				}
			} else if gotErr == nil {
				if got != test.want {
					t.Errorf("want LogLevel %v got %v", test.want, got)
				}
			} else {
				t.Errorf("Unexpected error: %v", gotErr)
			}

			if got.Get() != test.want {
				t.Errorf("want LogLevel %v got %v", test.want, got)
			}
		})
	}
}
