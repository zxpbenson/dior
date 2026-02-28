package main

import (
	"dior/component"
	"dior/internal/lg"
	_ "dior/internal/sink"
	_ "dior/internal/source"
	"dior/internal/version"
	"dior/option"
	"flag"
	"fmt"
	"os"
)

func main() {
	opts := option.NewOptions("dior")

	flagSet := option.FlagSet(opts)
	flagSet.Usage = func() {
		fmt.Println(usage())
		flagSet.PrintDefaults()
	}
	flagSet.Parse(os.Args[1:])

	if len(os.Args) == 1 {
		flagSet.Usage()
		os.Exit(0)
	}

	if flagSet.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Println(version.String("dior"))
		os.Exit(0)
	}

	jsonBytes, _ := opts.Json()
	if err := opts.Validate(); err != nil {
		fmt.Printf("config : %s\nparam error : %v\n", jsonBytes, err)
		os.Exit(1)
	}

	err := lg.InitDftLgr(opts.LogPrefix, opts.LogLevel)
	if err != nil {
		fmt.Printf("main.go goroutine logger create error : %v\n", err)
		os.Exit(1)
	}
	lg.DftLgr.Info("options : %s\n", jsonBytes)

	controller := component.NewController()
	if err := controller.AddComponents(opts); err != nil {
		fmt.Printf("Controller.AddComponents error : %v\n", err)
		os.Exit(1)
	}
	if err := controller.Init(); err != nil {
		fmt.Printf("Controller.Init error : %v\n", err)
		os.Exit(1)
	}
	controller.Start()

	lg.DftLgr.Info("Main goroutine done, bye.\n")
}

func usage() (usage string) {
	usage = "Supported press cases : " +
		"\n" +
		"\n  * press kafka" +
		"\n  * press nsq" +
		"\n" +
		"\nFor example, press Kafka : " +
		"\n" +
		"\n  ./build/dior \\" +
		"\n  --src press \\" +
		"\n  --src-file source.txt \\" +
		"\n  --src-speed 10 \\" +
		"\n  --dst kafka \\" +
		"\n  --dst-bootstrap-servers 127.0.0.1:9092 \\" +
		"\n  --dst-topic topic_to" +
		"\n" +
		"\nSupported transport cases : " +
		"\n" +
		"\n  * kafka to kafka" +
		"\n  * kafka to nsq" +
		"\n  * kafka to file" +
		"\n  * nsq to kafka" +
		"\n  * nsq to nsq" +
		"\n  * nsq to file" +
		"\n" +
		"\nFor example, kafka to nsq : " +
		"\n" +
		"\n  ./build/dior \\" +
		"\n  --src kafka \\" +
		"\n  --src-bootstrap-servers 127.0.0.1:9092 \\" +
		"\n  --src-topic topic_from \\" +
		"\n  --src-group benson \\" +
		"\n  --dst nsq \\" +
		"\n  --dst-nsqd-tcp-addresses 127.0.0.1:4150 \\" +
		"\n  --dst-topic topic_to\\" +
		"\n"

	return
}
