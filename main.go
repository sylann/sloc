package main

import (
	"fmt"
	"log"
	"os"
)

type fileStats struct {
	Path         string
	Lines        int
	LinesCode    int
	LinesEmpty   int
	LinesComment int
	LineBytesAvg int
	LineBytesMax int
}

func main() {
	filepath := "./samples/simple.go"
	buffer, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalln("File does not exist")
	}

	fst := fileStats{Path: filepath}
	inspectFile(buffer, &fst)

	fmt.Printf("Result %#v\n", fst)
}

func inspectFile(buffer []byte, fst *fileStats) {
	var (
		lbAll          int
		lbCode         int
		lbComment      int
		inBlockComment bool
		inLineComment  bool
		bytesPerLine   []int = make([]int, 0)
		prevByte       byte
	)

	// HACK: This probably won't work reliably.
	// Probably better to write "consumers" that expect specific bytes,
	// Look again how golang does it!
	for i, b := range buffer {
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
			log.Printf("Line %4d:  [%3d %3d %3d]  %s\n", fst.Lines, lbCode, lbComment, lbAll, buffer[i-lbAll+1:i])
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
	// TODO: Add stats from bytesPerLine
}
