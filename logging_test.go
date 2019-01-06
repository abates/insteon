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
	"fmt"
	"log"
	"strings"
	"testing"
)

func TestLogLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		level LogLevel
		str   string
	}{
		{LevelNone, "NONE"},
		{LevelInfo, "INFO"},
		{LevelDebug, "DEBUG"},
		{LevelTrace, "TRACE"},
		{LogLevel(-1), ""},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("level %q", test.str), func(t *testing.T) {
			t.Parallel()
			if test.str != test.level.String() {
				t.Errorf("got %q, want %q", test.level.String(), test.str)
			}
		})
	}
}

func TestLogging(t *testing.T) {
	t.Parallel()
	levels := []LogLevel{LevelNone, LevelInfo, LevelDebug}
	for _, level := range levels {
		messages := []string{}
		for _, l := range levels {
			if l == LevelNone {
				continue
			}
			messages = append(messages, fmt.Sprintf("%5s message", l))
		}

		buffer := bytes.NewBuffer([]byte{})
		testLogger := log.New(buffer, "", 0)
		logger := &Logger{logger: testLogger}
		logger.Level(level)

		logger.Infof("message")
		logger.Debugf("message")
		logger.Tracef("message")

		expected := strings.Join(messages[0:int(level)], "\n")
		if expected != "" {
			expected += "\n"
		}
		if expected != buffer.String() {
			t.Errorf("got %q, want %q", buffer.String(), expected)
		}
	}
}
