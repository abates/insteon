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
		logger := &Logger{Logger: testLogger, Level: level}

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

func TestLog(t *testing.T) {
	tests := []struct {
		desc  string
		level LogLevel
		tf    func(*Logger)
		want  string
	}{
		// LevelInfo
		{"Infof", LevelInfo, func(l *Logger) { l.Infof("test") }, "INFO test"},
		{"Errorf", LevelInfo, func(l *Logger) { l.Errorf(nil, "test") }, ""},
		{"Errorf", LevelInfo, func(l *Logger) { l.Errorf(ErrReadTimeout, "test") }, "INFO test"},
		{"Debugf", LevelInfo, func(l *Logger) { l.Debugf("test") }, ""},
		{"Tracef", LevelInfo, func(l *Logger) { l.Tracef("test") }, ""},
		// LevelDebug
		{"Infof", LevelDebug, func(l *Logger) { l.Infof("test") }, "INFO test"},
		{"Errorf", LevelDebug, func(l *Logger) { l.Errorf(nil, "test") }, ""},
		{"Errorf", LevelDebug, func(l *Logger) { l.Errorf(ErrReadTimeout, "test") }, "INFO test"},
		{"Debugf", LevelDebug, func(l *Logger) { l.Debugf("test") }, "DEBUG test"},
		{"Tracef", LevelDebug, func(l *Logger) { l.Tracef("test") }, ""},
		// LevelTrace
		{"Infof", LevelTrace, func(l *Logger) { l.Infof("test") }, "INFO test"},
		{"Errorf", LevelTrace, func(l *Logger) { l.Errorf(nil, "test") }, ""},
		{"Errorf", LevelTrace, func(l *Logger) { l.Errorf(ErrReadTimeout, "test") }, "INFO test"},
		{"Debugf", LevelTrace, func(l *Logger) { l.Debugf("test") }, "DEBUG test"},
		{"Tracef", LevelTrace, func(l *Logger) { l.Tracef("test") }, "TRACE test"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			builder := &strings.Builder{}
			logger := &Logger{Level: test.level, Logger: log.New(builder, "", 0)}
			test.tf(logger)
			got := strings.TrimSpace(builder.String())
			if !strings.HasSuffix(got, test.want) {
				t.Errorf("want string %v got %v", test.want, got)
			}
		})
	}
}
