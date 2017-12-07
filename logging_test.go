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
	levels := []LogLevel{LevelNone, LevelInfo, LevelDebug, LevelTrace}
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
