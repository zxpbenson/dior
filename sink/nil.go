package sink

import (
	"dior/component"
	"dior/lg"
	"dior/option"
)

type nilSink struct {
	*component.Asynchronizer
}

func init() {
	component.RegCmpCreator("nil-sink", newNilSink)
}

func newNilSink(opts *option.Options) (component.Component, error) {
	return &nilSink{
		Asynchronizer: &component.Asynchronizer{},
	}, nil
}

func (this *nilSink) Init() (err error) {
	this.Asynchronizer.Init()
	this.Output = this.output
	return nil
}

func (this *nilSink) output(data []byte) {
	lg.DftLgr.Debug("NilSink.output data : %v", data)
}
