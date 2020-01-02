// Network connectivity is required to get the page.
// Tests are using one page using one get or an available file in output directory
// The page is updated by replacing one word. It is available to test as []byte and a file.
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

const (
	techName = "Google"
	myTech   = "MyTech"
	wantf    = "originalpage.html"
	updatedf = "updatedpage.html"
	wd       = "output"
	// Syscall package does not know other OS and messages are unreachable
	// On Travis, syscall messages seem unusable
	ERROR_FILE_NOT_FOUND = "The system cannot find the file specified"

	EEXIST = "no such file or directory"
)

// Only one read on the network or filled with the existing want file
var wantb []byte

func TestMain(m *testing.M) {
	resp, err := http.Get("https://about.google/intl/en_be/")

	OutputDir(wd)
	if err == nil {
		defer resp.Body.Close()
		if _, err = os.Stat(wantf); os.IsNotExist(err) {
			// File missing, create it
			log.Printf("creating %s file\n", wantf)
			err = ReadCloserToFile(wantf, resp.Body)
			if err != nil {
				log.Fatalf("create want file failed with %v", err)
			}
		}
		// File updates will occur in the tests
		wantb, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
	} else {
		// No network mainly... Let us fill the buffer with the file
		perm := os.FileMode(0444)
		if runtime.GOOS == "windows" {
			perm = 0400
		}
		f, err := os.OpenFile(wantf, os.O_RDONLY, perm)
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
	createTestFiles()
	e := m.Run()
	removeTestFiles()
	os.Exit(e)
}

// Buffer is used as a string and produces a file
// The check is using FileCompare to detect an error
// The error is used for the test and this method by the Benchmark
func GetPageStringToFile(name string) error {
	// got file is identical to want file - no page update
	StringToFile(name, wantb)
	return FileCompare(name, wantf) // second element is the func name
}

// Test creation of a new file with an updated content. Error must be returned by comparison.
func TestPageStringToFile(t *testing.T) {
	var err error
	if err = os.Remove(t.Name()); err != nil && !os.IsNotExist(err) {
		log.Println(err)
	}
	// First run fails when file is created.
	err = GetPageStringToFile(t.Name())
	travis := len(os.Getenv("TRAVIS_GO_VERSION")) != 0
	if err != nil && !travis && !strings.Contains(fmt.Sprintf("%v", err), "want file is larger by") {
		t.Error(err)
	}
	if err := os.Remove(t.Name()); err != nil {
		log.Println(err)
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
	// got file is rewritten with the updated page
	// Replaces techname to get a different page. A reference file is created.
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	BufferToFile(name, wantbuf)
	return FileCompare(name, wantf)
}

// Create a file from a buffer
func TestBufferToFile(t *testing.T) {
	b := new(bytes.Buffer)
	b.Write(wantb)
	BufferToFile(t.Name(), b)
	if err := os.Remove(t.Name()); err != nil {
		log.Println(err)
	}

}

func TestBufferCompareNoDiff(t *testing.T) {
	var err error
	// Replaces techname to get a different page. A reference file is created.
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	if err = os.Remove(t.Name()); err != nil && !os.IsNotExist(err) {
		log.Println(err)
	}
	BufferToFile(t.Name(), wantbuf)
	if err = BufferCompare(wantbuf, wantf); err == nil {
		t.Error("no difference found")
	}
}

func TestBufferCompare(t *testing.T) {
	var err error
	// Replaces techname to get a different page. A reference file is created.
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	if err = os.Remove(t.Name()); err != nil && !os.IsNotExist(err) {
		log.Println(err)
	}
	BufferToFile(t.Name(), wantbuf)
	if err = BufferCompare(wantbuf, t.Name()); err != nil {
		t.Errorf("difference found. %v", err)
	}
	if err = os.Remove(t.Name()); err != nil {
		log.Println(err)
	}
}

// Create a file from a ReadCloser (r.Body)
func TestReadCloserToFile(t *testing.T) {
	b := new(bytes.Buffer)
	b.Write(wantb)
	if err := ReadCloserToFile("gotbuffer.html", ioutil.NopCloser(b)); err != nil {
		t.Error(err)
	}
}

func TestReadCloserCompareNoDiff(t *testing.T) {
	var err error
	// Replaces techname to get a different page. A reference file is created.
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	if err = os.Remove(t.Name()); err != nil && !os.IsNotExist(err) {
		log.Println(err)
	}
	BufferToFile(t.Name(), wantbuf)
	if err := ReadCloserCompare(ioutil.NopCloser(wantbuf), wantf); err == nil {
		t.Error("no difference found")
	}
	if err = os.Remove(t.Name()); err != nil {
		log.Println(err)
	}
}

func TestReadCloserCompare(t *testing.T) {
	// Replaces techname to get a different page. A reference file is created.
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	var err error
	if err = os.Remove(t.Name()); err != nil && !os.IsNotExist(err) {
		log.Println(err)
	}
	BufferToFile(t.Name(), wantbuf)
	if err := ReadCloserCompare(ioutil.NopCloser(wantbuf), t.Name()); err != nil {
		t.Errorf("difference found: %v", err)
	}
	if err = os.Remove(t.Name()); err != nil {
		log.Println(err)
	}
}

// Benchmarks
// File operation is the most consuming. One file less means half the time.
// Buffer has a minor advantage over string.
func BenchmarkGetPageStringToFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageStringToFile("stringtofile.html"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetPageBufferToFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageBufferToFile(updatedf); err != nil {
			b.Fatal(err)
		}
	}
}

// No got file. Comparing buffer to want file. Got file created only if different
func GetPageBufferCompare() error {
	i, _, _, _ := runtime.Caller(0)
	funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), ".")
	if len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	}
	fn := funcname[1]
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	if _, err := os.Stat(fn); err != nil {
		BufferToFile(fn, wantbuf)
	}
	return BufferCompare(wantbuf, fn)
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
	funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), ".")
	if len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	}
	fn := funcname[1]
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	if _, err := os.Stat(fn); err != nil {
		BufferToFile(fn, wantbuf)
	}
	return ReadCloserCompare(ioutil.NopCloser(wantbuf), fn)
}

func BenchmarkGetPageReadCloserCompare(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageReadCloserCompare(); err != nil {
			b.Fatal(err)
		}
	}
}

func TestOutputDir(t *testing.T) {
	defer func() {
		if err := recover().(error); !os.IsNotExist(err) {
			t.Errorf("Recovering failed with %v", err)
		}
	}()
	OutputDir("doesnotexist")
}

// Errors on files are compared to go lang values. Only Windows and Linux message are foreseen.
// The error returned by FileCompare is not the file error only to provide relevant information.
func IsFileError(errm string) bool {
	if runtime.GOOS == "windows" {
		return strings.Contains(errm, ERROR_FILE_NOT_FOUND)
	}
	return strings.Contains(errm, EEXIST)
}

// Panic-ing on invalid file
func recoverFileSystem(t *testing.T) {
	if err := recover().(error); !os.IsNotExist(err) {
		t.Errorf("Recovering failed with %v", err)
	}
}

func TestStringToFilePanicFilename(t *testing.T) {
	defer recoverFileSystem(t)
	StringToFile("", nil)
}

func TestBufferToFilePanicFilename(t *testing.T) {
	defer recoverFileSystem(t)
	BufferToFile("", nil)
}

func TestReadCloserToFilePanicFilename(t *testing.T) {
	defer recoverFileSystem(t)
	ReadCloserToFile("", nil)
}

func TestBufferCompareFileFail(t *testing.T) {
	if err := BufferCompare(nil, "willfail"); !os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestReadCloserCompareFileFail(t *testing.T) {
	if err := ReadCloserCompare(nil, "willfail"); !os.IsNotExist(err) {
		t.Error(err)
	}
}

// Panic-ing on nil content is not increasing coverage
func recoverNilContent(t *testing.T) {
	if r := fmt.Sprint(recover()); !strings.Contains(r, "invalid memory address or nil pointer dereference") {
		t.Errorf("Recovering failed with %v", r)
	}
	os.Remove("nilcontent") // File is created
}

/* Not testing
func TestStringToFilePanicContent(t *testing.T) {
	defer recoverFileSystem(t)
	StringToFile("nilcontent", []byte(nil))
}

*/

func TestBufferToFilePanicContent(t *testing.T) {
	defer recoverNilContent(t)
	BufferToFile("nilcontent", nil)
	t.Fatalf("nil content did not panic")
}

func TestReadCloserToFilePanicContent(t *testing.T) {
	defer recoverNilContent(t)
	_ = ReadCloserToFile("nilcontent", nil)
	t.Fatalf("nil content did not panic")
}

func TestFileCompareDoesNotExist(t *testing.T) {
	OutputDir("output")
	if err := fmt.Sprint(FileCompare("doesnotmatter", "doesnotexist")); !IsFileError(err) {
		t.Errorf("Non-existent got file not returned but %v", err)
	}
	if err := fmt.Sprint(FileCompare("doesnotexist", "originalpage.html")); !IsFileError(err) {
		t.Errorf("Non-existent want file not returned but %v", err)
	}
}

func TestFileCompareDifference(t *testing.T) {
	createTestFiles()
	if err := FileCompare("afile", "abfile"); fmt.Sprint(err) != "want file is larger by 1 bytes" {
		t.Errorf("%v", err)
	}
	if err := FileCompare("abfile", "afile"); fmt.Sprint(err) != "got file is larger by 1 bytes" {
		t.Errorf("%v", err)
	}
	if err := FileCompare("abfile", "acfile"); fmt.Sprint(err) != `got "c", want "b" at 1` {
		t.Errorf("%v", err)
	}
}

func createTestFiles() {
	OutputDir("output")
	b := new(bytes.Buffer)
	b.WriteString("a")
	BufferToFile("afile", b)
	b.WriteByte('b')
	BufferToFile("abfile", b)
	b.Reset()
	b.WriteString("abc")
	BufferToFile("abcfile", b)
	b.Reset()
	b.WriteString("ac")
	BufferToFile("acfile", b)
}

func removeTestFiles() {
	p, err := os.Getwd()
	if err != nil {
		return
	}
	if filepath.Base(p) == wd {
		os.RemoveAll(p)
	}
}

func TestBufferCompareDifference(t *testing.T) {
	b := new(bytes.Buffer)
	b.WriteString("ac")
	if err := BufferCompare(b, "acfile"); err != nil {
		t.Errorf("%v", err)
	}
	// TODO Add dump file existence and size
	b.Reset()
	b.WriteString("ac")
	if err := BufferCompare(b, "abfile"); fmt.Sprint(err) != `got 'c', want "b" at 1` {
		t.Errorf("%v", err)
	}
	b.Reset()
	b.WriteString("ab")
	if err := BufferCompare(b, "afile"); fmt.Sprint(err) != "got buffer is too long by 1" {
		t.Errorf("%v", err)
	}
	if c, err := b.ReadByte(); err != nil || c != 'b' {
		t.Errorf("unreadbyte failed: %q", c)
	}
	b.Reset()
	b.WriteString("a")
	if err := BufferCompare(b, "acfile"); fmt.Sprint(err) != `got EOF and last byte 'c' is missing` {
		t.Errorf("%v", err)
	}
	b.Reset()
	b.WriteString("a")
	if err := BufferCompare(b, "abcfile"); fmt.Sprint(err) != `TestBufferCompareDifference : got EOF, want 'b' at 1. Buffer is missing 2` {
		t.Errorf("%v", err)
	}
}

func TestReadCloserCompareDifference(t *testing.T) {
	b := new(bytes.Buffer)
	b.WriteString("ac")
	if err := ReadCloserCompare(ioutil.NopCloser(b), "acfile"); err != nil {
		t.Errorf("%v", err)
	}
	b.Reset()
	b.WriteString("ac")
	if err := ReadCloserCompare(ioutil.NopCloser(b), "abfile"); !strings.Contains(fmt.Sprint(err),
		`got "c", want "b" at 1`) {
		t.Errorf("%v", err)
	}
	b.Reset()
	b.WriteString("ab")
	// Length should be 1 but the last byte read from the response is not written to file
	// but can be recovered in the error message
	if err := ReadCloserCompare(ioutil.NopCloser(b), "afile"); !strings.Contains(fmt.Sprint(err),
		`got response is too long by 0. Last read byte`) {
		fn := "TestReadCloserCompareDifference"
		if fs, errf := os.Stat(fn); errf == nil && fs.Size() == 0 {
			os.Remove(fn)
		}
		t.Errorf("%v", err)
	}
	b.Reset()
	b.WriteString("a")
	if err := ReadCloserCompare(ioutil.NopCloser(b), "acfile"); !strings.Contains(fmt.Sprint(err), `got EOF, want "c" at 1. Response is missing 1`) {
		t.Errorf("%v", err)
	}
	b.Reset()
	b.WriteString("a")
	if err := ReadCloserCompare(ioutil.NopCloser(b), "abcfile"); !strings.Contains(fmt.Sprint(err), `got EOF, want "b" at 1. Response is missing 2`) {
		t.Errorf("%v", err)
	}
}

// Creating file write errors
func TestStringToFilePanicContent(t *testing.T) {
	t.Skip("setting file permissions does not fail test")
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("no error recovered %v", err)
		}
		if err.(error) != nil {
			t.Errorf("Recovering failed with %v", err)
		}
		OutputDir("output")
	}()
	err := os.Mkdir("willpanic", 0000) // Read only dir
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir("willpanic"); err != nil {
		t.Fatal(err)
	}
	StringToFile("willpanic", []byte{'a'})
	if err = os.Chdir(".."); err != nil {
		t.Fatal(err)
	}
}
