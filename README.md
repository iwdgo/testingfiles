[![GoDoc](https://godoc.org/github.com/iWdGo/testingfiles?status.svg)](https://godoc.org/github.com/iWdGo/testingfiles)
[![Go Report Card](https://goreportcard.com/badge/github.com/iwdgo/testingfiles)](https://goreportcard.com/report/github.com/iwdgo/testingfiles)

# Testing when large output is expected

This module provides a few func's for testing using files.
A file is useful or needed because of the size of the output for instance.

The suggested work method is to have reference files in a directory which becomes the working directory using OutputDir.
Then the content is put in a file and files are compared.

## Working directory

The subdirectory avoids to have the data files mixed with source code.
The directory is not created but its existence is checked.
If the working directory (not the temp, nor the executing) is unavailable,
tests will panic.

## Buffer comparison

To avoid overhead, the buffer can also be compared to a file.
A file is produced with the reminder from the difference found.

## Testing

Testing is done through repository `github.com/iwdgo/largeoutput`  