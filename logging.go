package insteon

import (
	"log"
	"os"
)

var (
	// Log is the global log object. The default level is set to Info
	Log = &Logger{level: LevelInfo, logger: log.New(os.Stderr, "", log.LstdFlags)}
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

const (
	// LevelNone - log nothing to Stderr
	LevelNone = iota

	// LevelInfo - log normal information (warnings) to Stderr
	LevelInfo

	// LevelDebug - log debug information (used for development and troubleshooting)
	LevelDebug

	// LevelTrace - log everything, including I/O
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
		format = sprintf("%5s %s", level, format)
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
