[![GoDoc](https://godoc.org/github.com/iWdGo/testingfiles?status.svg)](https://godoc.org/github.com/iWdGo/testingfiles)
[![Go Report Card](https://goreportcard.com/badge/github.com/iwdgo/testingfiles)](https://goreportcard.com/report/github.com/iwdgo/testingfiles)

# Using reference files for large output

A reference file (want) is useful or needed because of the size of the output or for recording purposes.
It is useful for testing.

Reference files are in a working directory.
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

## Testing of the module

Testing requires online access for one read. Otherwise, a reference file must be available.
No file is saved and the first test run will fail and produced the required file.
Got file can be renamed into a want file and the second run will be successful.

Benchmarking is fairly unconclusive.

```
go version go1.13beta1 windows/amd64
pkg: github.com/iwdgo/testingfiles
BenchmarkGetPageStringToFile-4                 1        1029407400 ns/op
BenchmarkGetPageBufferToFile-4                 1        1108362700 ns/op
BenchmarkGetPageBufferCompare-4                2         512716000 ns/op
BenchmarkGetPageReadCloserCompare-4            3         545020000 ns/op
```
