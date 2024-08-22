package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

type progOptions struct {
	debug   bool
	tsvPath string
}

func main() {
	conf := progOptions{}
	flag.BoolVar(&conf.debug, "debug", false, "Whether to print debug logs.")
	flag.StringVar(&conf.tsvPath, "tsv", "", "Path of a tsv file to write detailed results. No effect if not provided. Use '-' to print to stdout.")
	flag.Parse()
	paths := flag.Args()

	if len(paths) == 0 {
		fmt.Printf("USAGE: %s [-debug][-tsv TSVPATH] PATH [...]\n", os.Args[0])
		os.Exit(1)
	}

	if !conf.debug {
		log.SetOutput(io.Discard) // disable logging
	}

	gst := NewGlobalStats(paths)
	gst.InspectBatch()
	gst.PrintGlobalStats()

	switch conf.tsvPath {
	case "": // do nothing
	case "-":
		gst.DumpStatDetailsAsTsv(os.Stdout)
	default:
		f, err := os.Create(conf.tsvPath)
		if err != nil {
			fmt.Println("Could not write tsv file:", err.Error())
			os.Exit(2)
		}
		defer f.Close()

		gst.DumpStatDetailsAsTsv(f)
		fmt.Println("Detailed results written to", conf.tsvPath)
	}
}
