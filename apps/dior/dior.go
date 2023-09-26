package main

import (
	"dior/component"
	"dior/lg"
	"dior/option"
	_ "dior/sink"
	_ "dior/source"
	"fmt"
	"os"
)

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
