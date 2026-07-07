package main

import (
	"flag"
	"fmt"
	"forth/translator"
	"log"
	"os"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("source and target paths are required")
		flag.Usage()
		os.Exit(1)
	}
	sourcePath := args[0]
	targetPath := args[1]
	translator.Translate(sourcePath, targetPath)
	log.Printf("Translated %s to %s", sourcePath, targetPath)
}
