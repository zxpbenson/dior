package option

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"strings"
)

type Options struct {
	// basic options
	app       string
	LogLevel  string `flag:"log-level"`
	LogPrefix string `flag:"log-prefix"`

	Src                    string   `flag:"src"` // nsq / kafka / press
	SrcTopic               string   `flag:"src-topic"`
	SrcLookupdTCPAddresses []string `flag:"src-lookupd-tcp-addresses"`
	SrcNSQDTCPAddresses    []string `flag:"src-nsqd-tcp-addresses"`
	SrcChannel             string   `flag:"src-channel"`
	SrcBootstrapServers    []string `flag:"src-bootstrap-servers"`
	SrcGroup               string   `flag:"src-group"`
	SrcSpeed               int64    `flag:"src-speed"`
	SrcFile                string   `flag:"src-file"`

	ChanSize int `flag:"chan-size"`

	Dst                    string   `flag:"dst"` // nsq / kafka
	DstLookupdTCPAddresses []string `flag:"dst-lookupd-tcp-address"`
	DstNSQDTCPAddresses    []string `flag:"dst-nsqd-tcp-address"`
	DstBootstrapServers    []string `flag:"dst-bootstrap-servers"`
	DstTopic               string   `flag:"dst-topic"`
	DstFile                string   `flag:"dst-file"`
	DstBufSizeByte         int      `flag:"dst-buf-size-byte"`
}

func (this *Options) Validate() error {
	if err := this.validateSrc(); err != nil {
		return err
	}

	if this.ChanSize < 0 {
		return errors.New("param [chan-size] : >= 0")
	}

	if err := this.validateDst(); err != nil {
		return err
	}
	return nil
}

func (this *Options) Json() ([]byte, error) {
	return json.Marshal(this)
}

func NewOptions(app string) *Options {
	logPrefix := "[" + app + "] "
	return &Options{
		app:       app,
		LogPrefix: logPrefix,
		LogLevel:  "info",

		SrcLookupdTCPAddresses: make([]string, 0),
		SrcNSQDTCPAddresses:    make([]string, 0),
		SrcBootstrapServers:    make([]string, 0),
		SrcSpeed:               10,

		ChanSize: 100,

		DstLookupdTCPAddresses: make([]string, 0),
		DstNSQDTCPAddresses:    make([]string, 0),
		DstBootstrapServers:    make([]string, 0),
		DstBufSizeByte:         4096,
	}
}

func FlagSet(opts *Options) *flag.FlagSet {
	flagSet := flag.NewFlagSet(opts.app, flag.ExitOnError)

	flagSet.StringVar(&opts.LogLevel, "log-level", opts.LogLevel, "set log verbosity: debug, info, warn, error, or fatal")
	flagSet.StringVar(&opts.LogPrefix, "log-prefix", opts.LogPrefix, "log message prefix")

	flagSet.StringVar(&opts.Src, "src", opts.Src, "source type, options : nsq, kafka, press")

	flagSet.StringVar(&opts.SrcTopic, "src-topic", opts.SrcTopic, "source topic of nsq or kafka")

	flagSet.Func("src-lookupd-tcp-addresses", "<addr>:<port>[,<addr>:<port>] to connect for source of nsq client", func(s string) error {
		if s == "" {
			return nil
		}
		opts.SrcLookupdTCPAddresses = strings.Split(s, ",")
		return nil
	})
	flagSet.Func("src-nsqd-tcp-addresses", "<addr>:<port>[,<addr>:<port>] to connect for source of nsq client", func(s string) error {
		if s == "" {
			return nil
		}
		opts.SrcNSQDTCPAddresses = strings.Split(s, ",")
		return nil
	})
	flagSet.StringVar(&opts.SrcChannel, "src-channel", opts.SrcChannel, "source channel of nsq consumer")

	flagSet.Func("src-bootstrap-servers", "<addr>:<port>[,<addr>:<port>] to connect for source of kafka client", func(s string) error {
		if s == "" {
			return nil
		}
		opts.SrcBootstrapServers = strings.Split(s, ",")
		return nil
	})
	flagSet.StringVar(&opts.SrcGroup, "src-group", opts.SrcGroup, "source group of kafka consumer")

	flagSet.Int64Var(&opts.SrcSpeed, "src-speed", opts.SrcSpeed, "speed of data writing per seconds, >= 0, 0 means as fast as possible，default 10")
	flagSet.StringVar(&opts.SrcFile, "src-file", opts.SrcFile, "path of file to read data for source")

	flagSet.IntVar(&opts.ChanSize, "chan-size", opts.ChanSize, "size of queue between source and sink, >= 0, 0 means non blocking queue")

	flagSet.StringVar(&opts.Dst, "dst", opts.Dst, "destination type, options : nsq, kafka")

	flagSet.StringVar(&opts.DstTopic, "dst-topic", opts.DstTopic, "destination topic of nsq or kafka")

	flagSet.Func("dst-lookupd-tcp-addresses", "<addr>:<port>[,<addr>:<port>] to connect for sink of nsq client", func(s string) error {
		if s == "" {
			return nil
		}
		opts.DstLookupdTCPAddresses = strings.Split(s, ",")
		return nil
	})
	flagSet.Func("dst-nsqd-tcp-addresses", "<addr>:<port>[,<addr>:<port>] to connect for sink of nsq client", func(s string) error {
		if s == "" {
			return nil
		}
		opts.DstNSQDTCPAddresses = strings.Split(s, ",")
		return nil
	})
	flagSet.Func("dst-bootstrap-servers", "<addr>:<port>[,<addr>:<port>] to connect for sink of kafka client", func(s string) error {
		if s == "" {
			return nil
		}
		opts.DstBootstrapServers = strings.Split(s, ",")
		return nil
	})
	flagSet.StringVar(&opts.DstFile, "dst-file", opts.SrcFile, "path of file to write data for sink")
	flagSet.IntVar(&opts.DstBufSizeByte, "dst-buf-size-byte", opts.DstBufSizeByte, "buf size of sink writer, >= 0, default 4096")
	return flagSet
}

func (this *Options) validateSrc() error {
	this.Src = strings.ToLower(this.Src)

	if this.Src != "nsq" && this.Src != "kafka" && this.Src != "press" {
		return errors.New("param [src] : nsq, kafka, press")
	}
	if this.Src == "kafka" {
		if err := this.validateSrcKafka(); err != nil {
			return err
		}
	}
	if this.Src == "nsq" {
		if err := this.validateSrcNSQ(); err != nil {
			return err
		}
	}
	if this.Src == "press" {
		if err := this.validateSrcPress(); err != nil {
			return err
		}
	}
	return nil
}

func (this *Options) validateSrcKafka() error {
	if len(this.SrcBootstrapServers) == 0 {
		return errors.New("param [src-bootstrap-servers] : required")
	}
	if this.SrcGroup == "" {
		return errors.New("param [src-group] : required")
	}
	if this.SrcTopic == "" {
		return errors.New("param [src-topic] : required")
	}
	return nil
}

func (this *Options) validateSrcNSQ() error {
	if len(this.SrcLookupdTCPAddresses) == 0 && len(this.SrcNSQDTCPAddresses) == 0 {
		return errors.New("param [src-lookupd-tcp-addresses] or [src-nsqd-tcp-addresses] : required")
	}
	if this.SrcChannel == "" {
		return errors.New("param [src-channel] : required")
	}
	if this.SrcTopic == "" {
		return errors.New("param [src-topic] : required")
	}
	return nil
}

func (this *Options) validateSrcPress() error {
	if this.SrcSpeed < 0 {
		return errors.New("param [src-speed] : >= 0")
	}
	if this.SrcFile == "" {
		return errors.New("param [src-file] : required")
	} else {
		fileInfo, err := os.Stat(this.SrcFile)
		if err != nil {
			if os.IsNotExist(err) {
				return errors.New("SrcFile not exists : " + this.SrcFile)
			}
			return errors.Join(err, errors.New("SrcFile access error : "+this.DstFile))
		}
		if fileInfo.IsDir() {
			return errors.New("SrcFile It`s a directory : " + this.SrcFile)
		}
		if fileInfo.Size() < 2 {
			return errors.New("SrcFile It`s empty : " + this.SrcFile)
		}
	}
	return nil
}

func (this *Options) validateDst() error {
	this.Dst = strings.ToLower(this.Dst)
	if this.Dst != "nsq" && this.Dst != "kafka" && this.Dst != "nil" && this.Dst != "file" {
		return errors.New("param [dst] : nsq, kafka, nil, file")
	}

	if this.Dst == "kafka" {
		if err := this.validateDstKafka(); err != nil {
			return err
		}
	}
	if this.Dst == "nsq" {
		if err := this.validateDstNSQ(); err != nil {
			return err
		}
	}
	if this.Dst == "nil" {
	}
	if this.Dst == "file" {
		if err := this.validateDstFile(); err != nil {
			return err
		}
	}
	return nil
}

func (this *Options) validateDstKafka() error {
	if len(this.DstBootstrapServers) == 0 {
		return errors.New("param [dst-bootstrap-servers] : required")
	}
	if this.DstTopic == "" {
		return errors.New("param [dst-topic] : required")
	}
	return nil
}

func (this *Options) validateDstNSQ() error {
	if len(this.DstLookupdTCPAddresses) == 0 && len(this.DstNSQDTCPAddresses) == 0 {
		return errors.New("param [dst-lookupd-tcp-addresses] or [dst-nsqd-tcp-addresses] : required")
	}
	if this.DstTopic == "" {
		return errors.New("param [dst-topic] : required")
	}
	return nil
}

func (this *Options) validateDstFile() error {
	if this.DstFile == "" {
		return errors.New("param [dst-file] : required")
	} else {
		//fileInfo, err := os.Stat(this.DstFile)
		_, err := os.Stat(this.DstFile)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return errors.Join(err, errors.New("DstFile access error : "+this.DstFile))
		}
		return errors.New("DstFile exists : " + this.DstFile)
		//if fileInfo.IsDir() {
		//	return errors.New("DstFile It`s a directory : " + this.DstFile)
		//}
		//if fileInfo.Size() > 1 {
		//	return errors.New("DstFile`s not empty : " + this.DstFile)
		//}
	}
	if this.DstBufSizeByte < 0 {
		return errors.New("param [dst-buf-size-byte] : >= 0")
	}
	return nil
}
