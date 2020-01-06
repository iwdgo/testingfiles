[![GoDoc](https://godoc.org/github.com/iWdGo/testingfiles?status.svg)](https://godoc.org/github.com/iWdGo/testingfiles)
[![Go Report Card](https://goreportcard.com/badge/github.com/iwdgo/testingfiles)](https://goreportcard.com/report/github.com/iwdgo/testingfiles)
[![codecov](https://codecov.io/gh/iWdGo/testingfiles/branch/master/graph/badge.svg)](https://codecov.io/gh/iWdGo/testingfiles)

[![Build Status](https://travis-ci.com/iWdGo/testingfiles.svg?branch=master)](https://travis-ci.com/iWdGo/testingfiles)
[![Build Status](https://api.cirrus-ci.com/github/iWdGo/testingfiles.svg)](https://cirrus-ci.com/github/iWdGo/testingfiles)
[![Build status](https://ci.appveyor.com/api/projects/status/eimlas99romrrro0?svg=true)](https://ci.appveyor.com/project/iWdGo/testingfiles)
![Build status](https://github.com/iwdgo/testingfiles/workflows/Go/badge.svg)

# Using reference files for large output

Reference data for a test can be on file for various reasons: large data set, recording, post mortem,...
`got` source can be `File`, `Buffer`, `ReadCloser` and is compared to a `want` file.
Comparison is more efficient when the `got` file is not needed.
The `bytes.Buffer` of the `html.Response` can be compared directly to the `want` file.

When comparison fails, a file is created with `got_` prefix from the byte where the first difference
appeared. No further check on the file is done.

When running your tests for the first time, test will fail as not `want` file is usuablly available.
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

Reference files are expected in a working directory with default `output`.
Using a subdirectory avoids to have the data files mixed with source code.
The directory is not created but its existence is checked.
If the working directory is unavailable, tests will panic.
In CI scripts, the working directory is created before running the tests.

## Testing of the module

Testing can be online or offline.

Online is used to read a reference page. 
Offline requires to provide the reference file.

Testing usage is demonstrated in modules of [largeoutput](https://github.com/iwdgo/largeoutput) repository.

Benchmarking is fairly inconclusive and is more of a _TODO_.

```
go version go1.13beta1 windows/amd64
pkg: github.com/iwdgo/testingfiles
BenchmarkGetPageStringToFile-4                 1        1029407400 ns/op
BenchmarkGetPageBufferToFile-4                 1        1108362700 ns/op
BenchmarkGetPageBufferCompare-4                2         512716000 ns/op
BenchmarkGetPageReadCloserCompare-4            3         545020000 ns/op
```
