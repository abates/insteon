package insteon

import (
	"fmt"
	"log"
	"os"
)

var Log = Logger(&bitbucketLogger{})
var StderrLogger = &stderrLogger{level: LevelInfo, logger: log.New(os.Stderr, "", log.LstdFlags)}

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

type Logger interface {
	Level(LogLevel)
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Tracef(format string, v ...interface{})
}

type bitbucketLogger struct{}

func (*bitbucketLogger) Level(LogLevel)                {}
func (*bitbucketLogger) Infof(string, ...interface{})  {}
func (*bitbucketLogger) Debugf(string, ...interface{}) {}
func (*bitbucketLogger) Tracef(string, ...interface{}) {}

type stderrLogger struct {
	level  LogLevel
	logger *log.Logger
}

func (s *stderrLogger) Level(level LogLevel) {
	s.level = level
}

func (s *stderrLogger) logf(level LogLevel, format string, v ...interface{}) {
	if s.level >= level {
		format = fmt.Sprintf("%5s %s", s.level, format)
		s.logger.Printf(format, v...)
	}
}

func (s *stderrLogger) Infof(format string, v ...interface{}) {
	s.logf(LevelInfo, format, v...)
}

func (s *stderrLogger) Debugf(format string, v ...interface{}) {
	s.logf(LevelDebug, format, v...)
}

func (s *stderrLogger) Tracef(format string, v ...interface{}) {
	s.logf(LevelTrace, format, v...)
}
