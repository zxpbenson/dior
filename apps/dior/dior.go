package main

import (
	"dior/component"
	"dior/lg"
	"dior/option"
	"dior/sink"
	"dior/source"
	"fmt"
	"os"
)

func init() {
	source.TriggerInit()
	sink.TriggerInit()
}

func main() {
	opts := option.NewOptions("dior")

	flagSet := option.FlagSet(opts)
	flagSet.Parse(os.Args[1:])

	json, _ := opts.Json()
	if err := opts.Validate(); err != nil {
		fmt.Printf("config : %s\nparam error : %v\n", json, err)
		os.Exit(1)
	}

	err := lg.InitDftLgr(opts.LogPrefix, opts.LogLevel)
	if err != nil {
		fmt.Printf("main goroutine logger create error : %v\n", err)
		os.Exit(1)
	}

	lg.DftLgr.Info("options : %s\n", json)

	controller := component.NewController()
	controller.AddComponents(opts)
	controller.Init()
	controller.Start()

	lg.DftLgr.Info("main goroutine dior done\n")
}
