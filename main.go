package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

type progOptions struct {
	debug bool
}

func main() {
	conf := progOptions{}
	flag.BoolVar(&conf.debug, "debug", false, "Whether to print debug logs.")
	flag.Parse()
	paths := flag.Args()

	if len(paths) == 0 {
		fmt.Printf("USAGE: %s PATH [...]\n", os.Args[0])
		os.Exit(1)
	}

	if !conf.debug {
		log.SetOutput(io.Discard) // disable logging
	}

	gst := NewGlobalStats(paths)
	gst.InspectBatch()
	gst.DumpStatDetailsAsTsv()
	gst.PrintGlobalStats()
}
