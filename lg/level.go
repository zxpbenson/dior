package lg

import (
	"fmt"
	"strings"
)

const (
	DEBUG = LogLevel(1)
	INFO  = LogLevel(2)
	WARN  = LogLevel(3)
	ERROR = LogLevel(4)
	FATAL = LogLevel(5)
)

type LogLevel int

func (l *LogLevel) Get() interface{} { return *l }

func (l *LogLevel) Set(s string) error {
	lvl, err := ParseLogLevel(s)
	if err != nil {
		return err
	}
	*l = lvl
	return nil
}

func (l *LogLevel) String() string {
	switch *l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	}
	return "invalid"
}

func ParseLogLevel(levelstr string) (LogLevel, error) {
	switch strings.ToLower(levelstr) {
	case "debug":
		return DEBUG, nil
	case "info":
		return INFO, nil
	case "warn":
		return WARN, nil
	case "error":
		return ERROR, nil
	case "fatal":
		return FATAL, nil
	}
	return 0, fmt.Errorf("invalid log level '%s' (debug, info, warn, error, fatal)", levelstr)
}
