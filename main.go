package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

type progOptions struct {
	paths   []string
	debug   bool
}

type fileStats struct {
	Path         string
	Error        error
	Lines        int
	LinesCode    int
	LinesEmpty   int
	LinesComment int
	LineBytesAvg int
	LineBytesMax int
}

func (f *fileStats) String() string {
	return f.Path +
		"\n\t- Lines: " + strconv.Itoa(f.Lines) +
		"\n\t- LinesCode: " + strconv.Itoa(f.LinesCode) +
		"\n\t- LinesEmpty: " + strconv.Itoa(f.LinesEmpty) +
		"\n\t- LinesComment: " + strconv.Itoa(f.LinesComment) +
		"\n\t- LineBytesAvg: " + strconv.Itoa(f.LineBytesAvg) +
		"\n\t- LineBytesMax: " + strconv.Itoa(f.LineBytesMax)
}

func main() {
	conf := progOptions{}
	flag.BoolVar(&conf.debug, "debug", false, "Whether to print debug logs.")
	flag.Parse()
	conf.paths = flag.Args()

	if len(conf.paths) == 0 {
		fmt.Printf("USAGE: %s PATH [...]\n", os.Args[0])
		os.Exit(1)
	}

	if !conf.debug {
		log.SetOutput(io.Discard) // disable logging
	}

	results := make([]fileStats, len(conf.paths))

	for i, fst := range results {
		inspectFile(conf.paths[i], &fst)
	}
}

func inspectFile(fp string, fst *fileStats) {
	fst.Path = fp

	f, err := os.Open(fp)
	if err != nil {
		fst.Error = err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	inspectReader(reader, fst)

	fmt.Println(fst.String())
}

func inspectReader(reader *bufio.Reader, fst *fileStats) error {
	var (
		i              int
		lbAll          int
		lbCode         int
		lbComment      int
		inBlockComment bool
		inLineComment  bool
		bytesPerLine   []int = make([]int, 0)
		prevByte       byte
	)

	const chunkSize = 1024
	for {
		buffer := make([]byte, chunkSize)
		n, err := reader.Read(buffer)
		if err == io.EOF {
			if n == 0 {
				return nil
			}
		} else if err != nil {
			return err
		}

		if n == 0 {
			return nil
		} else if n < chunkSize {
			buffer = buffer[:n]
		}

		// HACK: This probably won't work reliably.
		// Probably better to write "consumers" that expect specific bytes,
		// Look again how golang does it!
		for _, b := range buffer {
			i++
			lbAll++

			switch b {
			case '/':
				switch prevByte {
				case '/':
					if !inBlockComment && !inLineComment {
						inLineComment = true
					}
				case '*':
					if inBlockComment {
						inBlockComment = false
					}
				}

			case '*':
				switch prevByte {
				case '/':
					if !inBlockComment && !inLineComment {
						inBlockComment = true
					}
				}

			case '\n':
				fst.Lines++
				if lbCode > 0 {
					fst.LinesCode++
				}
				// A line may contain both code and comments
				if lbComment > 0 {
					fst.LinesComment++
				} else {
					fst.LinesEmpty++
				}
				log.Printf("Line %4d:  [%3d %3d %3d]\n", fst.Lines, lbCode, lbComment, lbAll)
				bytesPerLine = append(bytesPerLine, lbAll)
				lbAll = 0
				lbCode = 0
				lbComment = 0
				inLineComment = false

			case '\r':
			// do nothing

			case ' ', '\t':
			// do mothing
			// HACK: find another way to consider lines with only whitespace as empty

			default:
				// Any character other than whitespace and comment punctuation
				// contributes to code or comment
				if inBlockComment || inLineComment {
					lbComment++
				} else {
					lbCode++
				}
			}
			prevByte = b
		}
	}
	// TODO: Add stats from bytesPerLine
}
