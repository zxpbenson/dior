package option

import (
	"encoding/json"
	"flag"
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

func (this *Options) Json() ([]byte, error) {
	return json.Marshal(this)
}

func NewOptions(app string) *Options {
	logPrefix := "[" + app + "] "
	opts := &Options{
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
	opts.loadEnv()
	return opts
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
