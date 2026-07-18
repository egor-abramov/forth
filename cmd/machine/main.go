package main

import (
	"flag"
	"fmt"
	"forth/machine"
	"os"
)

func main() {
	var scalarMode bool
	flag.BoolVar(&scalarMode, "scalar", false, "")

	var trace bool
	flag.BoolVar(&trace, "trace", false, "")

	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("source and input paths are required")
		flag.Usage()
		os.Exit(1)
	}
	sourcePath := args[0]
	inputPath := args[1]
	machine.Simulate(sourcePath, inputPath, trace, scalarMode)
}
