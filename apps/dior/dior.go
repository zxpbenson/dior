package main

import (
	"dior/component"
	"dior/lg"
	"dior/option"
	_ "dior/sink"
	_ "dior/source"
	"dior/version"
	"flag"
	"fmt"
	"os"
)

func main() {
	opts := option.NewOptions("dior")

	flagSet := option.FlagSet(opts)
	flagSet.Parse(os.Args[1:])

	if flagSet.Lookup("help").Value.(flag.Getter).Get().(bool) || len(os.Args) == 1 {
		fmt.Println(usage())
		flagSet.Usage()
		os.Exit(0)
	}

	if flagSet.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Println(version.String("dior"))
		os.Exit(0)
	}

	json, _ := opts.Json()
	if err := opts.Validate(); err != nil {
		fmt.Printf("config : %s\nparam error : %v\n", json, err)
		os.Exit(1)
	}

	err := lg.InitDftLgr(opts.LogPrefix, opts.LogLevel)
	if err != nil {
		fmt.Printf("main.go goroutine logger create error : %v\n", err)
		os.Exit(1)
	}
	lg.DftLgr.Info("options : %s\n", json)

	controller := component.NewController()
	controller.AddComponents(opts)
	controller.Init()
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
		"\n  * press nsq" +
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
