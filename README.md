# SLOC Analyzer

A small program to produce various statistics related to the source lines of code
in a project or in a file.

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
- [ ]  add more stats
- [ ]  only dump global stats to stdout
  - [ ]  dump per file stats to a tsv file
  - [ ]  make tsv dump optional with a flag
- [ ]  inspect all files in directory
  - [ ]  in current directory by default
  - [ ]  in given directory (argument, default to current)
  - [ ]  reccursively
  - [ ]  inspect each file in a goroutine, aggregate results in parallel
  - [ ]  specific file type (argument)
  - [ ]  any file type, group results by file type
- [ ]  count runes instead of bytes?

