package sink

import (
	"bufio"
	"context"
	"dior/component"
	"dior/internal/lg"
	"dior/option"
	"os"
)

type fileSink struct {
	*component.Asynchronizer

	fileName string
	file     *os.File
	writer   *bufio.Writer
	splitter []byte
	bufSize  int
}

func init() {
	component.RegCmpCreator("file-sink", newFileSink)
}

func newFileSink(opts *option.Options) (component.Component, error) {
	return &fileSink{
		Asynchronizer: &component.Asynchronizer{},
		fileName:      opts.DstFile,
		splitter:      []byte("\n"),
		bufSize:       opts.DstBufSizeByte,
	}, nil
}

func (s *fileSink) Init(channel chan []byte) (err error) {
	s.Asynchronizer.Init(channel)
	s.Output = s.output

	s.file, err = os.Create(s.fileName)
	if err != nil {
		return
	}
	s.writer = bufio.NewWriterSize(s.file, s.bufSize) //创建新的 Writer 对象
	return
}

func (s *fileSink) Start(ctx context.Context) {
	s.Asynchronizer.Start(ctx)
}

func (s *fileSink) output(data []byte) {
	lg.DftLgr.Debug("FileSink.output data : %v", data)
	n, err := s.writer.Write(data)
	if err == nil {
		lg.DftLgr.Debug("FileSink.output write file %s byte count : %v", s.fileName, n)
	} else {
		lg.DftLgr.Error("FileSink.output write file %s error : %v", s.fileName, err)
	}
	n, err = s.writer.Write(s.splitter)
	if err == nil {
		//lg.DftLgr.Debug("FileSink.output write file %s byte count : %v", s.fileName, n)
	} else {
		lg.DftLgr.Error("FileSink.output write file %s error : %v", s.fileName, err)
	}
}

func (s *fileSink) Stop() {
	s.Asynchronizer.Stop()
	s.writer.Flush()
	s.file.Close()
	lg.DftLgr.Info("FileSink.Stop done.")
}
