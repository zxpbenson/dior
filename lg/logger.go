package lg

import (
	"log"
	"os"
)

type Logger interface {
	logf(msgLevel LogLevel, f string, args ...interface{})
	append_logf(appender Appender, msgLevel LogLevel, f string, args ...interface{})
	Debug(f string, args ...interface{})
	Info(f string, args ...interface{})
	Warn(f string, args ...interface{})
	Error(f string, args ...interface{})
	Fatal(f string, args ...interface{})
}

func NewLogger(logPrefix string, cfgLevel string) (Logger, error) {
	logLevel, err := ParseLogLevel(cfgLevel)
	if err != nil {
		return nil, err
	}
	return &StdtLogger{
		logPrefix: logPrefix,
		cfgLevel:  logLevel,
		appender:  log.New(os.Stdout, logPrefix, log.Ldate|log.Ltime|log.Lmicroseconds),
	}, nil
}
