package insteon

import (
	"fmt"
	"log"
	"os"
)

// Log is the global log object. The default level is set to Info
var Log = &Logger{level: LevelInfo, logger: log.New(os.Stderr, "", log.LstdFlags)}

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

// Log levels are None, Info, Debug and Trace. Trace logging
// should only be used to display packets and messages as they
// are received or sent
const (
	LevelNone = iota
	LevelInfo
	LevelDebug
	LevelTrace
)

// Logger is a struct that keeps track of a log level and only
// prints messages of that level or lower
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// Level sets the Loggers log level
func (s *Logger) Level(level LogLevel) {
	s.level = level
}

func (s *Logger) logf(level LogLevel, format string, v ...interface{}) {
	if s.level >= level {
		format = fmt.Sprintf("%5s %s", level, format)
		s.logger.Printf(format, v...)
	}
}

// Infof will print a message at the Info level
func (s *Logger) Infof(format string, v ...interface{}) {
	s.logf(LevelInfo, format, v...)
}

// Debugf will print a message at the Debug level
func (s *Logger) Debugf(format string, v ...interface{}) {
	s.logf(LevelDebug, format, v...)
}

// Tracef will print a message at the Trace level
func (s *Logger) Tracef(format string, v ...interface{}) {
	s.logf(LevelTrace, format, v...)
}
