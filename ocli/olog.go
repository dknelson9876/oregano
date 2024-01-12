package ocli

import "log"

type LogLevel int

const (
	Disable LogLevel = iota
	Debug
	DebugDetail
	DebugMinuteDetail
)

type OLogger struct {
	Level LogLevel
}

func NewOLogger(level LogLevel) *OLogger {
	return &OLogger{Level: level}
}

func (l *OLogger) Print(level LogLevel, v ...any) {
	if level <= l.Level {
		log.Print(v...)
	}
}

func (l *OLogger) Println(level LogLevel, v ...any) {
	if level <= l.Level {
		log.Println(v...)
	}
}

func (l *OLogger) Printf(level LogLevel, format string, v ...any) {
	if level <= l.Level {
		log.Printf(format, v...)
	}
}