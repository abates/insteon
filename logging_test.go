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

	for i, test := range tests {
		if test.str != test.level.String() {
			t.Errorf("tests[%d] expected %q got %q", i, test.str, test.level.String())
		}
	}
}

func TestLogging(t *testing.T) {
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
			t.Errorf("Expected %q got %q", expected, buffer.String())
		}
	}
}
