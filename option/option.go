package option

import (
	"dior/lg"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"strings"
)

type Logger lg.Logger

type Options struct {
	// basic options
	LogLevel  lg.LogLevel `flag:"log-level"`
	LogPrefix string      `flag:"log-prefix"`
	Logger    Logger

	Dest                   string   `flag:"dest"` // nsq / writer
	NSQLookupdTCPAddresses []string `flag:"lookupd-tcp-address"`
	NSQDTCPAddresses       string   `flag:"nsqd-tcp-address"`
	KafkaBootstrapServer   []string `flag:"bootstrap-server"`
	Topic                  string   `flag:"topic"`
	Speed                  int64    `flag:"speed"`
	DataFile               string   `flag:"data-file"`
}

func (this *Options) Validate() error {
	this.Dest = strings.ToLower(this.Dest)
	if this.Dest != "nsq" && this.Dest != "kafka" {
		return errors.New("param [dest] : nsq, kafka")
	}
	if this.Topic == "" {
		return errors.New("param [topic] : required")
	}
	if this.DataFile == "" {
		return errors.New("param [data-file] : required")
	} else {
		fileInfo, err := os.Stat(this.DataFile)
		if err != nil {
			if os.IsNotExist(err) {
				return errors.New("File not exists : " + this.DataFile)
			}
			return err
		}
		if fileInfo.IsDir() {
			return errors.New("It`s a directory : " + this.DataFile)
		}
		if fileInfo.Size() == 0 {
			return errors.New("It`s empty : " + this.DataFile)
		}
	}
	if this.Speed < 0 {
		return errors.New("param [speed] : >= 0")
	}
	if this.Dest == "nsq" && this.NSQDTCPAddresses == "" && len(this.NSQLookupdTCPAddresses) == 0 {
		return errors.New("param [lookupd-tcp-address] or [nsqd-tcp-address] : required if dest is nsq")
	}
	if this.Dest == "writer" && len(this.KafkaBootstrapServer) == 0 {
		return errors.New("param [bootstrap-server] : required if dest is writer")
	}
	return nil
}

func (this *Options) Json() ([]byte, error) {
	return json.Marshal(this)
}

func NewOptions() *Options {
	return &Options{
		LogPrefix: "[dior] ",
		LogLevel:  lg.INFO,

		NSQLookupdTCPAddresses: make([]string, 0),
		KafkaBootstrapServer:   make([]string, 0),
		Speed:                  10,
	}
}

func DiorFlagSet(opts *Options) *flag.FlagSet {
	flagSet := flag.NewFlagSet("dior", flag.ExitOnError)

	flagSet.IntVar((*int)(&opts.LogLevel), "log-level", int(opts.LogLevel), "set log verbosity: debug, info, warn, error, or fatal")
	flagSet.StringVar(&opts.LogPrefix, "log-prefix", opts.LogPrefix, "log message prefix")

	flagSet.StringVar(&opts.Dest, "dest", opts.Dest, "target type, options : nsq, writer")

	flagSet.Func("lookupd-tcp-address", "<addr>:<port>[,<addr>:<port>] to connect for nsq clients", func(s string) error {
		if s == "" {
			return nil
		}
		opts.NSQLookupdTCPAddresses = strings.Split(s, ",")
		return nil
	})

	flagSet.StringVar(&opts.NSQDTCPAddresses, "nsqd-tcp-address", opts.NSQDTCPAddresses, "<addr>:<port> to connect to for nsq clients")

	flagSet.Func("bootstrap-server", "<addr>:<port>[,<addr>:<port>] to connect for writer clients", func(s string) error {
		if s == "" {
			return nil
		}
		opts.KafkaBootstrapServer = strings.Split(s, ",")
		return nil
	})

	flagSet.StringVar(&opts.Topic, "topic", opts.Topic, "target topic of nsq or writer")
	flagSet.Int64Var(&opts.Speed, "speed", opts.Speed, "speed of data writing per seconds, >= 0")
	flagSet.StringVar(&opts.DataFile, "data-file", opts.DataFile, "path to data file")

	return flagSet
}
