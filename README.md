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
- [ ]  count maximum number of characters per file
- [ ]  count average number of characters per file
- [ ]  count maximum number of characters per line
- [ ]  count average number of characters per line
- [ ]  inspect all files in directory
  - [ ]  in current directory by default
  - [ ]  in given directory (argument, default to current)
  - [ ]  reccursively
  - [ ]  inspect each file in a goroutine, aggregate results in parallel
  - [ ]  specific file type (argument)
  - [ ]  any file type, group results by file type
- [ ]  count runes instead of bytes?

