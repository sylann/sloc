package main

import (
	"bufio"
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

	gst := newGlobalStats(paths)
	gst.inspectFiles()
	gst.DumpStatDetailsAsTsv()
	gst.GlobalStats()
}

type globalStats struct {
	files                 []*fileStats
	MaxLpfAll, MaxLpfCode int
	AvgLpfAll, AvgLpfCode float64
}

func newGlobalStats(paths []string) globalStats {
	size := len(paths)
	gst := globalStats{
		files: make([]*fileStats, size),
	}
	for i := 0; i < size; i++ {
		gst.files[i] = NewFileStats(paths[i])
	}
	return gst
}

// inspectFiles reads underlying files and stores aggregated statistics.
func (gst *globalStats) inspectFiles() {
	var (
		validFiles      int
		sumAll, sumCode int
		maxAll, maxCode int
	)
	for _, fst := range gst.files {
		err := fst.inspectFile()
		if err != nil {
			continue
		}
		validFiles++
		sumAll += fst.LinesAll
		sumCode += fst.LinesCode
		maxAll = max(maxAll, fst.LinesAll)
		maxCode = max(maxCode, fst.LinesCode)
	}
	gst.MaxLpfAll = maxAll
	gst.MaxLpfCode = maxCode
	gst.AvgLpfAll = float64(sumAll) / float64(validFiles)
	gst.AvgLpfCode = float64(sumCode) / float64(validFiles)
}

func (gst *globalStats) GlobalStats() {
	fmt.Printf("Files: %d\n", len(gst.files))
	fmt.Printf("Max LpF All:  %d\n", gst.MaxLpfAll)
	fmt.Printf("Max LpF Code: %d\n", gst.MaxLpfCode)
	fmt.Printf("Avg LpF All:  %.2f\n", gst.AvgLpfAll)
	fmt.Printf("Avg LpF Code: %.2f\n", gst.AvgLpfCode)
}

func (gst *globalStats) DumpStatDetailsAsTsv() {
	fmt.Println("Path\tError\tLinesAll\tLinesCode\tLinesEmpty\tLinesComment\tMaxBplAll\tMaxBplCode\tAvgBplAll\tAvgBplCode")
	for _, fst := range gst.files {
		fmt.Printf("%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%.2f\t%.2f\n",
			fst.Path,
			fst.Error(),
			fst.LinesAll,
			fst.LinesCode,
			fst.LinesEmpty,
			fst.LinesComment,
			fst.MaxBplAll,
			fst.MaxBplCode,
			fst.AvgBplAll,
			fst.AvgBplCode,
		)
	}
}

type fileStats struct {
	LinesAll, LinesCode, LinesEmpty, LinesComment int
	MaxBplAll, MaxBplCode                         int
	AvgBplAll, AvgBplCode                         float64
	Path                                          string
	err                                           error
}

func (fst *fileStats) Error() string {
	if fst.err == nil {
		return ""
	}
	return fst.err.Error()
}

func NewFileStats(path string) *fileStats {
	return &fileStats{Path: path}
}

// inspectFile reads the file and aggregates code syntax statistics of the content.
// If the path can't be read, or an underlying function fails, it sets fst.Error
// and returns the error.
func (fst *fileStats) inspectFile() error {
	f, err := os.Open(fst.Path)
	if err != nil {
		fst.err = err
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	return fst.inspectReader(reader)
}

func (fst *fileStats) inspectReader(reader *bufio.Reader) error {
	var (
		i                        int
		lbAll, lbCode, lbComment int
		sumAll, sumCode          int
		maxAll, maxCode          int
		inBlockComment           bool
		inLineComment            bool
		prevByte                 byte
	)

	const chunkSize = 1024
	for {
		buffer := make([]byte, chunkSize)
		n, err := reader.Read(buffer)
		if err == io.EOF {
			if n == 0 {
				goto calculateStats
			}
		} else if err != nil {
			return err
		}

		if n == 0 {
			goto calculateStats
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
				fst.LinesAll++
				if lbCode > 0 {
					fst.LinesCode++
				}
				// A line may contain both code and comments
				if lbComment > 0 {
					fst.LinesComment++
				}
				if lbCode == 0 && lbComment == 0 {
					fst.LinesEmpty++
				}
				log.Printf("Line %4d:  [%3d %3d %3d]\n", fst.LinesAll, lbCode, lbComment, lbAll)
				sumAll += lbAll
				maxAll = max(maxAll, lbAll)
				sumCode += lbCode
				maxCode = max(maxCode, lbCode)
				lbAll = 0
				lbCode = 0
				lbComment = 0
				inLineComment = false

			case '\r':
			// do nothing

			case ' ', '\t':
			// do nothing
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
calculateStats:
	fst.MaxBplAll = maxAll
	fst.MaxBplCode = maxCode
	fst.AvgBplAll = float64(sumAll) / float64(fst.LinesAll)
	fst.AvgBplCode = float64(sumCode) / float64(fst.LinesAll)
	return nil
}
