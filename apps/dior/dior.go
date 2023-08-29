package main

import (
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
		fmt.Printf("config : %s\n, done error : %v\n", json, err)
		os.Exit(1)
	}

	fmt.Printf("options : %s\n", json)

	pressWriter, err := pressor.NewPressAble(opts)
	if err != nil {
		fmt.Printf("done error : %v\n", err)
		os.Exit(1)
	}
	pressWriter.Start()
	fmt.Printf("dior done\n")
}
