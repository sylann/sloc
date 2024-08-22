package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// GlobalStats represents the aggregated code statistics of a batch of files.
//
// Vocabulary:
// - All: All lines
// - Code: Only lines containing at least one code symbol
// - Comment: Only lines containing at least one comment
// - Empty: Only lines containing only whitespace (no code and no comment)
// - Lpf: "Lines per file"
// - Avg: "Average"
type GlobalStats struct {
	files                                             []*FileStats
	MaxLpfAll, MaxLpfCode, MaxLpfComment, MaxLpfEmpty int
	AvgLpfAll, AvgLpfCode, AvgLpfComment, AvgLpfEmpty float64
}

// NewGlobalStats creates a proper new FileStats instance.
func NewGlobalStats(paths []string) GlobalStats {
	size := len(paths)
	gst := GlobalStats{
		files: make([]*FileStats, size),
	}
	for i := 0; i < size; i++ {
		gst.files[i] = NewFileStats(paths[i])
	}
	return gst
}

// InspectBatch reads underlying files and stores aggregated
// statistics for the whole batch.
func (gst *GlobalStats) InspectBatch() {
	var (
		validFiles                            int
		sumAll, sumCode, sumComment, sumEmpty int
		maxAll, maxCode, maxComment, maxEmpty int
	)
	for _, fst := range gst.files {
		err := fst.InspectFile()
		if err != nil {
			continue
		}
		validFiles++
		sumAll += fst.LinesAll
		sumCode += fst.LinesCode
		sumComment += fst.LinesComment
		sumEmpty += fst.LinesEmpty
		maxAll = max(maxAll, fst.LinesAll)
		maxCode = max(maxCode, fst.LinesCode)
		maxComment = max(maxComment, fst.LinesComment)
		maxEmpty = max(maxEmpty, fst.LinesEmpty)
	}
	gst.MaxLpfAll = maxAll
	gst.MaxLpfCode = maxCode
	gst.MaxLpfComment = maxComment
	gst.MaxLpfEmpty = maxEmpty
	gst.AvgLpfAll = float64(sumAll) / float64(validFiles)
	gst.AvgLpfCode = float64(sumCode) / float64(validFiles)
	gst.AvgLpfComment = float64(sumComment) / float64(validFiles)
	gst.AvgLpfEmpty = float64(sumEmpty) / float64(validFiles)
}

// PrintGlobalStats prints out global statistics about previously aggregated data.
func (gst *GlobalStats) PrintGlobalStats() {
	fmt.Printf("Files: %d\n", len(gst.files))
	fmt.Printf("Max LpF All:     %d\n", gst.MaxLpfAll)
	fmt.Printf("Max LpF Code:    %d\n", gst.MaxLpfCode)
	fmt.Printf("Max LpF Comment: %d\n", gst.MaxLpfComment)
	fmt.Printf("Max LpF Empty:   %d\n", gst.MaxLpfEmpty)
	fmt.Printf("Avg LpF All:     %.2f\n", gst.AvgLpfAll)
	fmt.Printf("Avg LpF Code:    %.2f\n", gst.AvgLpfCode)
	fmt.Printf("Avg LpF Comment: %.2f\n", gst.AvgLpfComment)
	fmt.Printf("Avg LpF Empty:   %.2f\n", gst.AvgLpfEmpty)
}

// DumpStatDetailsAsTsv prints out all aggregated file data in a detailed TSV table format.
func (gst *GlobalStats) DumpStatDetailsAsTsv(w io.Writer) {
	fmt.Fprintln(w, "Path\tError"+
		"\tLinesAll\tLinesCode\tLinesComment\tLinesEmpty"+
		"\tMaxBplAll\tMaxBplCode\tMaxBplComment"+
		"\tAvgBplAll\tAvgBplCode\tAvgBplComment")
	for _, fst := range gst.files {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%.2f\t%.2f\t%.2f\n",
			fst.Path,
			fst.Error(),
			fst.LinesAll,
			fst.LinesCode,
			fst.LinesComment,
			fst.LinesEmpty,
			fst.MaxBplAll,
			fst.MaxBplCode,
			fst.MaxBplComment,
			fst.AvgBplAll,
			fst.AvgBplCode,
			fst.AvgBplComment,
		)
	}
}

// FileStats represents the aggregated code statistics of a single file.
//
// Vocabulary:
// - All: All lines
// - Code: Only lines containing at least one code symbol
// - Comment: Only lines containing at least one comment
// - Empty: Only lines containing only whitespace (no code and no comment)
// - Bpl: "Bytes per line"
// - Avg: "Average"
type FileStats struct {
	LinesAll, LinesCode, LinesComment, LinesEmpty int
	MaxBplAll, MaxBplCode, MaxBplComment          int
	AvgBplAll, AvgBplCode, AvgBplComment          float64
	Path                                          string
	err                                           error
}

// Error returns the string representation of any error encountered
// after calling `InspectFile`. Returns the empty string if there is no error.
func (fst *FileStats) Error() string {
	if fst.err == nil {
		return ""
	}
	return fst.err.Error()
}

// NewFileStats creates a proper new FileStats instance.
func NewFileStats(path string) *FileStats {
	return &FileStats{Path: path}
}

// InspectFile reads the file and aggregates code syntax statistics of the content.
// If the path can't be read, or an underlying function fails, it sets fst.Error
// and returns the error.
func (fst *FileStats) InspectFile() error {
	f, err := os.Open(fst.Path)
	if err != nil {
		fst.err = err
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	return fst.inspectReader(reader)
}

// inspectReader is the underlying logic used by InspectFile.
func (fst *FileStats) inspectReader(reader *bufio.Reader) error {
	var (
		i                           int
		lbAll, lbCode, lbComment    int
		sumAll, sumCode, sumComment int
		maxAll, maxCode, maxComment int
		inBlockComment              bool
		inLineComment               bool
		prevByte                    byte
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
				sumAll += lbAll
				sumCode += lbCode
				sumComment += lbComment
				maxAll = max(maxAll, lbAll)
				maxCode = max(maxCode, lbCode)
				maxComment = max(maxComment, lbComment)
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
	fst.MaxBplComment = maxComment
	fst.AvgBplAll = float64(sumAll) / float64(fst.LinesAll)
	fst.AvgBplCode = float64(sumCode) / float64(fst.LinesAll)
	fst.AvgBplComment = float64(sumComment) / float64(fst.LinesAll)
	return nil
}
