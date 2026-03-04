package sink

import (
	"context"
	"dior/component"
	"dior/internal/lg"
	"dior/option"
)

type nilSink struct {
	*component.Asynchronizer
}

func init() {
	component.RegCmpCreator("nil-sink", newNilSink)
}

func newNilSink(name string, opts *option.Options) (component.Component, error) {
	return &nilSink{
		Asynchronizer: component.NewAsynchronizer(name),
	}, nil
}

func (s *nilSink) Init(channel chan []byte) (err error) {
	s.Asynchronizer.Init(channel)
	s.Output = s.output
	return nil
}

func (s *nilSink) output(data []byte) error {
	lg.DftLgr.Debug("NilSink.output data : %v", data)
	return nil
}

func (s *nilSink) Start(ctx context.Context) {
	s.Asynchronizer.Start(ctx)
}

func (s *nilSink) Stop() {
	s.Asynchronizer.Stop()
	lg.DftLgr.Info("NilSink.Stop done.")
}
