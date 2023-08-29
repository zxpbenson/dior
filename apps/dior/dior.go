package main

import (
	"dior/lg"
	"dior/option"
	"dior/pressor"
	"fmt"
	"os"
)

func main() {
	opts := option.NewOptions()

	flagSet := option.DiorFlagSet(opts)
	flagSet.Parse(os.Args[1:])

	json, _ := opts.Json()
	if err := opts.Validate(); err != nil {
		fmt.Printf("config : %s\nparam error : %v\n", json, err)
		os.Exit(1)
	}

	logger, err := lg.NewLogger(opts.LogPrefix, opts.LogLevel)
	if err != nil {
		fmt.Printf("logger create error : %v\n", err)
		os.Exit(1)
	}

	logger.Info("options : %s\n", json)

	pressWriter, err := pressor.NewPressor(opts)
	if err != nil {
		logger.Fatal("Processor create error : %v\n", err)
	}
	pressWriter.Start()
	logger.Info("dior done\n")
}
