[![Go Reference](https://pkg.go.dev/badge/github.com/iwdgo/testingfiles.svg)](https://pkg.go.dev/github.com/iwdgo/testingfiles)
[![Go Report Card](https://goreportcard.com/badge/github.com/iwdgo/testingfiles)](https://goreportcard.com/report/github.com/iwdgo/testingfiles)
[![codecov](https://codecov.io/gh/iwdgo/testingfiles/branch/master/graph/badge.svg)](https://codecov.io/gh/iwdgo/testingfiles)

[![Build Status](https://app.travis-ci.com/iwdgo/testingfiles.svg?branch=master)](https://app.travis-ci.com/iwdgo/testingfiles)
[![Build Status](https://api.cirrus-ci.com/github/iwdgo/testingfiles.svg)](https://cirrus-ci.com/github/iwdgo/testingfiles)
[![Build status](https://ci.appveyor.com/api/projects/status/eimlas99romrrro0?svg=true)](https://ci.appveyor.com/project/iWdGo/testingfiles)
![Build status](https://github.com/iwdgo/testingfiles/workflows/Go/badge.svg)

# Using reference files for large output

A `want` reference file is compared to data from a `got` source.
Comparison is provided for `File`, `Buffer` or `ReadCloser` where a file is the least efficient.

When comparison fails, a file is created with `got_` prefix from the byte where the first difference
appeared. No further check on the file is done.

When running tests for the first time, they might fail as no `want` file is usually available.
The produced `got` file can be renamed into a `want` file to have a second successful run.

### Offline test

```

func TestOffline(t *testing.T) {
	if err := tearDownOffline(handler(...), t.Name()); err != nil {
    		t.Error(err)
    }
}

func tearDownOffline(b *bytes.Buffer, s string) (err error) {
	if b == nil {
		return errors.New("bytes.Buffer is nil")
	}
	testingfiles.OutputDir("output")
	
	if err := testingfiles.BufferCompare(b, s); err != nil {
        return err
    }
    return nil
}

```

### Online test

```

func TestOnline(t *testing.T) {
	resp, err := http.Get(getAppUrl("").String())
	if err != nil {
		t.Fatal(err)
	}
	tearDown(t, resp)
}

func tearDown(t *testing.T, resp *http.Response) {
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.StatusCode != 200 {
		t.Fatalf("request failed with error %d for %s", resp.StatusCode, s)
	}
	testingfiles.OutputDir("output")
	
	if err := testingfiles.ReadCloserCompare(resp.Body, t.Bame()); err != nil {
        t.Error(err)
    }
}

```

## Working directory

Reference files are expected to reside in a working directory which defaults to `./output`.
Using a subdirectory avoids having the data files mixed with source code.
The directory is not created but its existence is checked.
If the working directory is unavailable, tests will panic.
In CI scripts, the working directory is created before running the tests.

## Testing of the module

Testing can be online or offline.

Online is used to read a reference page. 
Offline requires to provide the reference file.

Testing usage is demonstrated in modules of [largeoutput](https://github.com/iwdgo/largeoutput) repository.

Benchmarking between string and bytes.Buffer is inconclusive inline with documented behavior.

```
go version go1.13beta1 windows/amd64
pkg: github.com/iwdgo/testingfiles
BenchmarkGetPageStringToFile-4                 1        1029407400 ns/op
BenchmarkGetPageBufferToFile-4                 1        1108362700 ns/op
BenchmarkGetPageBufferCompare-4                2         512716000 ns/op
BenchmarkGetPageReadCloserCompare-4            3         545020000 ns/op
```
