package insteon

import (
	"fmt"
	"log"
	"os"
)

var Log = &Logger{level: LevelInfo, logger: log.New(os.Stderr, "", log.LstdFlags)}

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

const (
	LevelNone = iota
	LevelInfo
	LevelDebug
	LevelTrace
)

type Logger struct {
	level  LogLevel
	logger *log.Logger
}

func (s *Logger) Level(level LogLevel) {
	s.level = level
}

func (s *Logger) logf(level LogLevel, format string, v ...interface{}) {
	if s.level >= level {
		format = fmt.Sprintf("%5s %s", s.level, format)
		s.logger.Printf(format, v...)
	}
}

func (s *Logger) Infof(format string, v ...interface{}) {
	s.logf(LevelInfo, format, v...)
}

func (s *Logger) Debugf(format string, v ...interface{}) {
	s.logf(LevelDebug, format, v...)
}

func (s *Logger) Tracef(format string, v ...interface{}) {
	s.logf(LevelTrace, format, v...)
}
