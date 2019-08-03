[![GoDoc](https://godoc.org/github.com/iWdGo/testingfiles?status.svg)](https://godoc.org/github.com/iWdGo/testingfiles)
[![Go Report Card](https://goreportcard.com/badge/github.com/iwdgo/testingfiles)](https://goreportcard.com/report/github.com/iwdgo/testingfiles)
[![codecov](https://codecov.io/gh/iWdGo/testingfiles/branch/master/graph/badge.svg)](https://codecov.io/gh/iWdGo/testingfiles)

# Using reference files for large output

Keeping reference on a file is useful becayse of the size of the output or for recording purposes.
Such a file is useful for testing.

Reference files are expected in a working directory.
Depending on the got source (`File`, `buffer`, `ReadCloser`), a comparison method is available.
If feasible, using directly the buffer or the response is more efficient than writing for file first.
When comparison fails and something is left in the buffer, a file is created with `got_` prefix.

## Working directory

The subdirectory avoids to have the data files mixed with source code.
The directory is not created but its existence is checked.
If the working directory (not the temp, nor the executing) is unavailable,
tests will panic.

In CI (Travis), the working directory is created.

## Testing of the module

Testing requires online access for one read. Otherwise, a reference file must be available.
No file is saved and the first test run will fail and produced the required file.
Got file can be renamed into a want file and the second run will be successful.

Benchmarking is fairly inconclusive and is more of a TODO.

```
go version go1.13beta1 windows/amd64
pkg: github.com/iwdgo/testingfiles
BenchmarkGetPageStringToFile-4                 1        1029407400 ns/op
BenchmarkGetPageBufferToFile-4                 1        1108362700 ns/op
BenchmarkGetPageBufferCompare-4                2         512716000 ns/op
BenchmarkGetPageReadCloserCompare-4            3         545020000 ns/op
```
