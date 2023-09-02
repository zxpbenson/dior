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

func (this *StdtLogger) append_logf(appender Appender, msgLevel LogLevel, f string, args ...interface{}) {
	if this.cfgLevel > msgLevel {
		return
	}
	appender.Output(3, fmt.Sprintf(msgLevel.String()+": "+f, args...))
}

func (this *StdtLogger) logf(msgLevel LogLevel, f string, args ...interface{}) {
	if this.cfgLevel > msgLevel {
		return
	}
	this.appender.Output(3, fmt.Sprintf(msgLevel.String()+": "+f, args...))
}

func (this *StdtLogger) Debug(f string, args ...interface{}) {
	this.logf(DEBUG, f, args...)
}

func (this *StdtLogger) Info(f string, args ...interface{}) {
	this.logf(INFO, f, args...)
}

func (this *StdtLogger) Warn(f string, args ...interface{}) {
	this.logf(WARN, f, args...)
}

func (this *StdtLogger) Error(f string, args ...interface{}) {
	this.logf(ERROR, f, args...)
}

func (this *StdtLogger) Fatal(f string, args ...interface{}) {
	appender := log.New(os.Stderr, this.logPrefix, log.Ldate|log.Ltime|log.Lmicroseconds)
	this.append_logf(appender, FATAL, f, args...)
	os.Exit(1)
}

func (this *StdtLogger) Enable(msgLevel LogLevel) bool {
	return msgLevel >= this.cfgLevel
}
