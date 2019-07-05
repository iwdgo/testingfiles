// Network connectivity is required to get the page.
// The first run of GetHTMLPage will fail and you can rename pagegot.html into pagewant.html

package testingfiles

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

/*
File operation is the most consuming. One file less means half the time.
Buffer has a minor advantage over string.

goos: windows
goarch: amd64
BenchmarkGetHTMLPageString-4                           1        1059377800 ns/op
BenchmarkGetHTMLPageBuffer-4                           2         920314550 ns/op
BenchmarkGetHTMLPageBufferNoGotFile-4                  2         513047750 ns/op
*/
// TODO Rename testing
const (
	techName = "Google"
	myTech   = "MyTech"
)

// TODO Return []byte, *bytes.Buffer, ...
func getTechHomePage() []byte {
	resp, err := http.Get("https://about.google/intl/en_be/")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return html
}

func getHTMLPage() []byte {
	// It is assumed that replacement is case sensitive
	return bytes.Replace(getTechHomePage(), []byte(techName), []byte(myTech), -1)
}

//  Benchmark is about the file comparison
func GetPageStringToFile() error {
	OutputDir("output")
	pfileName := "pagegot.html"
	StringToFile(pfileName, getHTMLPage())
	i, _, _, _ := runtime.Caller(0)
	if funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), "."); len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	} else {
		return FileCompare(pfileName, "pagewant.html") // second element is the func name
	}
}

/* Buffer to file, iso String. Then comparing files. No real gain */
func GetPageBufferToFile() error {
	OutputDir("output")
	pfileName := "pagegot.html"
	BufferToFile(pfileName, bytes.NewBuffer(getHTMLPage()))
	i, _, _, _ := runtime.Caller(0)
	if funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), "."); len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	} else {
		return FileCompare(pfileName, "pagewant.html") // second element is the func name
	}
}

// No got file. Comparing buffer to want file. Got file created only if different
func GetPageBufferCompareNoGotFile() error {
	OutputDir("output") // for want file
	i, _, _, _ := runtime.Caller(0)
	if funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), "."); len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	} else {
		return BufferCompare(bytes.NewBuffer(getHTMLPage()), "pagewant.html")
	}
}

// Tests
func TestGetHTMLPageString(t *testing.T) {
	if err := GetPageStringToFile(); err != nil {
		t.Error(err)
	}
}

func TestGetHTMLPageBuffer(t *testing.T) {
	if err := GetPageBufferToFile(); err != nil {
		t.Error(err)
	}
}

func TestGetPageBufferCompareNoGotFile(t *testing.T) {
	if err := GetPageBufferCompareNoGotFile(); err != nil {
		t.Error(err)
	}
}

// Benchmarks
func BenchmarkGetPageStringToFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageStringToFile(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetPageBufferToFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageBufferToFile(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetPageBufferCompareNoGotFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageBufferCompareNoGotFile(); err != nil {
			b.Fatal(err)
		}
	}
}
