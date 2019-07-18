// Network connectivity is required to get the page.
// The first run of GetHTMLPage will fail and you can rename pagegot.html into pagewant.html

package testingfiles

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TODO Expand into difference detection
const (
	techName = "Google"
	myTech   = "MyTech"
	wantf    = "pagewant.html"
)

// Only one read on the network or filled with the existing want file
var wantb []byte

func TestMain(m *testing.M) {
	resp, err := http.Get("https://about.google/intl/en_be/")

	OutputDir("output")
	if err == nil {
		defer resp.Body.Close()
		if _, err = os.Stat(wantf); os.IsNotExist(err) {
			// File missing, create it
			log.Printf("creating %s file\n", wantf)
			err = ReadCloserToFile(wantf, resp.Body)
			if err != nil {
				log.Fatalf("create want file failed with %v", err)
			}
		} else {
			// TODO Update the file
			log.Fatalf("%v\n", err)
		}
		wantb, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
	} else {
		// No network mainly... Let us fill the buffer with the file
		// TODO Check permissions
		f, err := os.OpenFile(wantf, os.O_RDONLY, 777)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		fs, err := os.Stat(wantf)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		wantb = make([]byte, fs.Size())
		n, err := f.Read(wantb)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		if n != len(wantb) {
			log.Printf("page is trucated by %d\n", len(wantb)-n)
		}
	}
	os.Exit(m.Run())
}

// Replaces techname to get a different page
func bytesToBuffer() *bytes.Buffer {
	// It is assumed that replacement is case sensitive
	b := new(bytes.Buffer)
	_, _ = b.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	return b
}

// Buffer is used as a string and produces a file
// The check is using FileCompare to detect an error
// The error is used for the test and this method by the Benchmark
func GetPageStringToFile(name string) error {
	// got file is identical to want file - no page update
	StringToFile(name, wantb)
	i, _, _, _ := runtime.Caller(0)
	if funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), "."); len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	} else {
		return FileCompare(name, wantf) // second element is the func name
	}
}

// Test creation of a new file with an updated content. Error must be returned by comparison.
func TestPageStringToFile(t *testing.T) {
	if err := GetPageStringToFile("pagegot.html"); err != nil {
		t.Error(err)
	}
}

// Comparing a file to itelf must return nil
func TestFileCompare(t *testing.T) {
	if err := FileCompare(wantf, wantf); err != nil {
		t.Error(err)
	}
}

// Buffer to file, iso String. Then comparing files.
func GetPageBufferToFile(name string) error {
	// got file is rewritten with the updated values
	BufferToFile(name, bytesToBuffer())
	i, _, _, _ := runtime.Caller(0)
	if funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), "."); len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	} else {
		return FileCompare(name, "pagewant.html") // second element is the func name
	}
}

// Create a file from a buffer
func TestBufferToFile(t *testing.T) {
	b := new(bytes.Buffer)
	b.Write(wantb)
	BufferToFile("gotbuffer.html", b)
}

func TestBufferCompare(t *testing.T) {
	if err := BufferCompare(bytesToBuffer(), "pagewant.html"); err != nil {
		t.Error(err)
	}
}

// Create a file from a ReadCloser (r.Body)
func TestReadCloserToFile(t *testing.T) {
	b := new(bytes.Buffer)
	b.Write(wantb)
	err := ReadCloserToFile("gotbuffer.html", ioutil.NopCloser(b))
	if err != nil {
		t.Error(err)
	}
}

func TestReadCloserCompare(t *testing.T) {
	if err := ReadCloserCompare(ioutil.NopCloser(bytesToBuffer()), "pagewant.html"); err != nil {
		t.Error(err)
	}
}


// Benchmarks
// File operation is the most consuming. One file less means half the time.
// Buffer has a minor advantage over string.
func BenchmarkGetPageStringToFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageStringToFile("pagegot.html"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetPageBufferToFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageBufferToFile("pagegot.html"); err != nil {
			b.Fatal(err)
		}
	}
}

// No got file. Comparing buffer to want file. Got file created only if different
func GetPageBufferCompare() error {
	i, _, _, _ := runtime.Caller(0)
	if funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), "."); len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	} else {
		return BufferCompare(bytesToBuffer(), "pagegot.html")
	}
}

func BenchmarkGetPageBufferCompare(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageBufferCompare(); err != nil {
			b.Fatal(err)
		}
	}
}

// No got file. Comparing buffer to want file. Got file created only if different
func GetPageReadCloserCompare() error {
	i, _, _, _ := runtime.Caller(0)
	if funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), "."); len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	} else {
		return ReadCloserCompare(ioutil.NopCloser(bytesToBuffer()), "pagegot.html")
	}
}

func BenchmarkGetPageReadCloserCompare(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageReadCloserCompare(); err != nil {
			b.Fatal(err)
		}
	}
}
