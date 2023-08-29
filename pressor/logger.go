package pressor

import (
	"dior/lg"
)

type Logger lg.Logger

const (
	LOG_DEBUG = lg.DEBUG
	LOG_INFO  = lg.INFO
	LOG_WARN  = lg.WARN
	LOG_ERROR = lg.ERROR
	LOG_FATAL = lg.FATAL
)

func (this *Pressor) logf(level lg.LogLevel, f string, args ...interface{}) {
	opts := this.opts
	lg.Logf(opts.Logger, opts.LogLevel, level, f, args...)
}
