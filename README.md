[![GoDoc](https://godoc.org/github.com/iWdGo/testingfiles?status.svg)](https://godoc.org/github.com/iWdGo/testingfiles)
[![Go Report Card](https://goreportcard.com/badge/github.com/iwdgo/testingfiles)](https://goreportcard.com/report/github.com/iwdgo/testingfiles)

# Testing when large output is expected

This module eases the use of reference files for testing.
A reference file (want) is useful or needed because of the size of the output or for recording purposes.

Supported work method is to place reference files in a working directory.
Content is put in a file and files are compared.
Compring a buffer (*bytes.Buffer or io.ReadCloser) is more efficient when feasible.

## Working directory

The subdirectory avoids to have the data files mixed with source code.
The directory is not created but its existence is checked.
If the working directory (not the temp, nor the executing) is unavailable,
tests will panic.

## Buffer comparison

To avoid overhead, the buffer can also be compared to a file.
If comparison fails, a got file is produced with the reminder from the difference found.

## Testing

Testing requires online access. No file is saved and the first test run will fail.
Got file can be renamed into a want file.  