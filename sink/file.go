package sink

import (
	"bufio"
	"dior/component"
	"dior/lg"
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

func (this *fileSink) Init() (err error) {
	this.Asynchronizer.Init()
	this.Output = this.output

	this.file, err = os.Create(this.fileName)
	if err != nil {
		return
	}
	this.writer = bufio.NewWriterSize(this.file, this.bufSize) //创建新的 Writer 对象
	return
}

func (this *fileSink) output(data []byte) {
	lg.DftLgr.Debug("FileSink.output data : %v", data)
	n, err := this.writer.Write(data)
	if err == nil {
		lg.DftLgr.Debug("FileSink.output write file %s byte count : %v", this.fileName, n)
	} else {
		lg.DftLgr.Error("FileSink.output write file %s error : %v", this.fileName, err)
	}
	n, err = this.writer.Write(this.splitter)
	if err == nil {
		//lg.DftLgr.Debug("FileSink.output write file %s byte count : %v", this.fileName, n)
	} else {
		lg.DftLgr.Error("FileSink.output write file %s error : %v", this.fileName, err)
	}
}

func (this *fileSink) Stop() {
	this.Asynchronizer.Stop()
	this.writer.Flush()
	this.file.Close()
	lg.DftLgr.Info("FileSink.Stop done.")
}
