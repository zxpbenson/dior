package sink

import (
	"bufio"
	"context"
	"dior/component"
	"dior/internal/lg"
	"dior/option"
	"os"
	"sync/atomic"
)

type fileSink struct {
	*component.Asynchronizer

	fileName       string
	file           *os.File
	writer         *bufio.Writer
	splitter       []byte
	bufSize        int
	processedCount atomic.Int64 // 跟踪写入条数
}

func init() {
	component.RegCmpCreator("file-sink", newFileSink)
}

func newFileSink(name string, opts *option.Options) (component.Component, error) {
	return &fileSink{
		Asynchronizer: component.NewAsynchronizer(name),
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

func (s *fileSink) output(data []byte) {
	if lg.DftLgr.Enable(lg.DEBUG) {
		lg.DftLgr.Debug("FileSink.output data: %v", string(data))
	}

	// 写入数据
	n, err := s.writer.Write(data)
	if err != nil {
		lg.DftLgr.Error("FileSink.output write data error: %v", err)
		return
	}
	lg.DftLgr.Debug("FileSink.output write file %s byte count: %v", s.fileName, n)

	// 写入分隔符
	n, err = s.writer.Write(s.splitter)
	if err != nil {
		lg.DftLgr.Error("FileSink.output write separator error: %v", err)
		return
	}

	// 每写入100条数据刷新一次缓冲区，避免内存占用过高
	if s.processedCount.Add(1)%100 == 0 {
		if err := s.writer.Flush(); err != nil {
			lg.DftLgr.Error("FileSink.output flush error: %v", err)
		}
	}
}

func (s *fileSink) Start(ctx context.Context) {
	s.Asynchronizer.Start(ctx)
}

func (s *fileSink) Stop() {
	// 先刷新缓冲区
	if err := s.writer.Flush(); err != nil {
		lg.DftLgr.Error("FileSink.Stop flush error: %v", err)
	}

	// 关闭文件
	if err := s.file.Close(); err != nil {
		lg.DftLgr.Error("FileSink.Stop close error: %v", err)
	}

	// 停止异步处理
	s.Asynchronizer.Stop()

	lg.DftLgr.Info("FileSink.Stop done.")
}
