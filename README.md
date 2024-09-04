# SLOC Analyzer

A small program to produce various statistics related to the source lines of code
in a project or in a file.

## Usage

Example run on the Golang source code base.
This requires having globs (`/**/`) enabled.

```sh
go run . -tsv dist/go-src.tsv /usr/local/go/src/**/*.go
```

## Build

```sh
go build -o dist/sloc -ldflags "-s"
```

## TODO

- [x]  inspect file character by character
- [x]  count lines
- [x]  count empty lines
- [x]  count comments
  - handle "/*...*/" -> " "
  - handle "//...\n" -> "\n"
- [x]  count sloc (source lines of code) (not empty, and not a comment)
- [x]  count maximum number of characters per file (global max)
- [x]  count average number of characters per file (global average)
- [x]  count maximum number of characters per line (file max)
- [x]  count average number of characters per line (file average)
- [x]  add more stats
- [x]  only dump global stats to stdout
  - [x]  dump per file stats to a tsv file
  - [x]  make tsv dump optional with a flag
- [ ]  improve global stats output
- [ ]  accept new options: root dir and file extensions (remove path args?)
       (globs won't work on all systems, and are already annoying in xargs)
  - in current directory by default
  - works reccursively (add option to enable/disable later?)
  - extensions mandatory or not
- [ ]  group results by extension

## Refactoring

- [ ]  find a way to improve how stats generation is coded

## Optimization

> Frankly for now, there does not seem to be a need for optimization as it already runs instantly on the full go code base.
> It even runs almost instantly when I forget to ignore blackholes such as node_modules!

- [ ]  test again whether chunked reading makes a meaningful difference
- [ ]  test what part of the write/print process costs the most
- [ ]  inspect each file in a goroutine, aggregate results in parallel
