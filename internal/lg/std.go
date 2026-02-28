package lg

import (
	"fmt"
	"log"
	"os"
)

type StdtLogger struct {
	logPrefix string
	cfgLevel  LogLevel
	appender  Appender
}

func (l *StdtLogger) append_logf(appender Appender, msgLevel LogLevel, f string, args ...interface{}) {
	if l.cfgLevel > msgLevel {
		return
	}
	appender.Output(3, fmt.Sprintf(msgLevel.String()+": "+f, args...))
}

func (l *StdtLogger) logf(msgLevel LogLevel, f string, args ...interface{}) {
	if l.cfgLevel > msgLevel {
		return
	}
	l.appender.Output(3, fmt.Sprintf(msgLevel.String()+": "+f, args...))
}

func (l *StdtLogger) Debug(f string, args ...interface{}) {
	l.logf(DEBUG, f, args...)
}

func (l *StdtLogger) Info(f string, args ...interface{}) {
	l.logf(INFO, f, args...)
}

func (l *StdtLogger) Warn(f string, args ...interface{}) {
	l.logf(WARN, f, args...)
}

func (l *StdtLogger) Error(f string, args ...interface{}) {
	l.logf(ERROR, f, args...)
}

func (l *StdtLogger) Fatal(f string, args ...interface{}) {
	appender := log.New(os.Stderr, l.logPrefix, log.Ldate|log.Ltime|log.Lmicroseconds)
	l.append_logf(appender, FATAL, f, args...)
	os.Exit(1)
}

func (l *StdtLogger) Enable(msgLevel LogLevel) bool {
	return msgLevel >= l.cfgLevel
}
