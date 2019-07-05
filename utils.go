// Package testingfiles provides primitives to use files as reference for testing
package testingfiles

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// OutputDir corrects the default dir to the base folder s where reference files (want files) are stored.
// The file is searched above and below working directory.
func OutputDir(s string) {
	ex, err := os.Getwd() // Executable() is where Go runs not where the files are created
	if err != nil {
		panic(err)
	}
	if filepath.Base(ex) != s { // No need to change
		err = os.Chdir("./" + s)
		if err != nil {
			err = os.Chdir("../test/" + s) // go to test/<want-files>
			if err != nil {
				panic(err) // subdirectory is probably missing
			}
		}
	}
}

// StringToFile produces a file named fname with the content
func StringToFile(fname string, content []byte) {
	wfile, err := os.Create(fname)
	defer wfile.Close()
	if err != nil {
		panic(err)
	}

	_, err = wfile.Write(content)
	if err != nil {
		panic(err)
	}
}

// BufferToFile produces a file named fname with the content
func BufferToFile(fname string, content *bytes.Buffer) {
	wfile, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer wfile.Close()

	_, err = wfile.Write(content.Bytes())
	if err != nil {
		panic(err)
	}
}

// ReadCloserToFile creates a file named fname with the content
func ReadCloserToFile(fname string, content io.ReadCloser) error {
	wfile, err := os.Create(fname)
	defer wfile.Close()
	if err != nil {
		panic(err)
	}
	c, err := ioutil.ReadAll(content)
	if err != nil {
		panic(err)
	}
	n, err := wfile.Write(c)
	if err != nil {
		panic(err)
	}
	if len(c) != n {
		return fmt.Errorf("file %s is missing %d bytes", fname, len(c)-n)
	}
	return nil
}

// FileCompare checks large outputs of a test when a file storage is more convenient or required.
// Names of the files to compare are passed as arguments and searched in the working directory.
func FileCompare(got, want string) error {
	rfile, err := os.Open(want)
	defer rfile.Close()
	if err != nil {
		return fmt.Errorf("want file %s open failed with %v", want, err)
	}

	pfile, err := os.Open(got)
	defer pfile.Close()
	if err != nil {
		return fmt.Errorf("got file %s open failed with %v", got, err)
	}

	b1, b2 := make([]byte, 1), make([]byte, 1)
	index := 0          // Index in file to locate error
	for err != io.EOF { // Until the end of the file
		_, err = rfile.Read(b1)
		if err != io.EOF { // While not EOF, read the other file too
			if err != nil {
				return err
			}
			_, err = pfile.Read(b2)
			if err != nil { // If EOF is returned, file is too short
				return err
			}
		}

		if !bytes.Equal(b1, b2) {
			return fmt.Errorf("got %q, want %q at %d", b1, b2, index)
		}
		index++
	}
	// EOF on reference file has been reached, let us check the produced file
	_, err = pfile.Read(b2)
	if err != io.EOF { // If EOF is not returned, produced file is shorter than reference
		rfileInfo, _ := rfile.Stat()
		return fmt.Errorf("got file is too short by %d", rfileInfo.Size()-int64(index))
	}
	return nil
}

// BufferCompare compares the buffer to a file.
// If an error occurs, got file is created and the error is returned.
// If identical, nil is returned.
func BufferCompare(got *bytes.Buffer, want string) error {
	wantf, err := os.Open(want)
	if err != nil {
		return fmt.Errorf("Reference file %s open failed with %v", want, err)
	}
	defer wantf.Close()

	// Finding caller name to
	i, _, _, _ := runtime.Caller(1) // Skipping the calling test
	funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), ".")
	if len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	}

	b1 := make([]byte, 1)
	var b2 byte
	index := 0          // Index in file to locate error
	for err != io.EOF { // Until the end of the file
		_, err = wantf.Read(b1)
		if err != io.EOF { // While not EOF, read the buffer
			if err != nil {
				return err // error on file was not io.EOF
			}

			b2, err = got.ReadByte()
			// If EOF is returned, buffer is too short and exhausted.
			if err != nil {
				wantfInfo, _ := wantf.Stat()
				// Last byte of the file is sometimes returned with io.EOF
				if wantfInfo.Size()-int64(index) == 1 && err == io.EOF {
					if b1[0] == b2 {
						log.Println("last byte returned with io.EOF")
						// Overriding error
						return nil
					} else {
						// Occurs when original buffer is used
						return fmt.Errorf("%v and last byte %q is missing", err, b1[0])
					}
				}
				return fmt.Errorf("%s : got %v, want %q at %d. Buffer is missing %d",
					funcname[1], err, b1[0], index, wantfInfo.Size()-int64(index))
			}

			if b1[0] != b2 {
				// The erroneous char is missing from the file but if got.UnreadByte() then
				// the file char is already read.
				BufferToFile(fmt.Sprintf("got_%s", funcname[1]), got)
				return fmt.Errorf("got %q, want %q at %d", b1, b2, index)
			}
			index++
		} else if err != nil && err != io.EOF {
			return fmt.Errorf("%s : read from want failed: %v", funcname[1], err)
		}
	}
	// EOF on want file has been reached
	b2, err = got.ReadByte()
	// If EOF is not returned, buffer is longer than the file which is exhausted.
	if err != io.EOF {
		BufferToFile(fmt.Sprintf("got_%s", funcname[1]), got)
		return fmt.Errorf("got buffer is too long by %d", got.Len())
	}
	return nil
}

// If an error occurs, got file is created and the error is returned.
// If identical, nil is returned.
// Logic and method are identical to *buffer.Bytes but duplicating the code avoids ReadAll.
// TODO Benchmark ReadAll agains specific byte by byte code
func ReadCloserCompare(got io.ReadCloser, want string) error {
	wantf, err := os.Open(want)
	if err != nil {
		return fmt.Errorf("Reference file %s open failed with %v", want, err)
	}
	defer wantf.Close()

	// Finding caller name to
	i, _, _, _ := runtime.Caller(1) // Skipping the calling test
	funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), ".")
	if len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	}

	// Actual comparison
	wantb, gotb := make([]byte, 1), make([]byte, 1)
	gotf := fmt.Sprintf("got_%s", funcname[1])
	n, index := 0, 0    // Index in file to locate error
	for err != io.EOF { // Until the end of the file
		_, err = wantf.Read(wantb)
		if err != io.EOF { // While file is not EOF, read the buffer
			if err != nil {
				return err // file reading failed, the read error is returned
			}

			n, err = got.Read(gotb)
			// Requires git 2.22.0 on Windows
			if err == io.EOF { // If EOF produced, buffer is too short (and empty)
				wantfInfo, _ := wantf.Stat()
				// Last byte of the file is returned with io.EOF
				if wantfInfo.Size()-int64(index) == 1 {
					if n == 1 && gotb[0] == wantb[0] {
						log.Println("last byte returned with io.EOF")
						return nil
					} else {
						// Occurs when original buffer is used
						return fmt.Errorf("last byte %q is missing", gotb[0])
					}
				}
				return fmt.Errorf("%s : got %v, want %q at %d. Buffer is missing %d",
					funcname[1], err, wantb, index, wantfInfo.Size()-int64(index))
			} else if err != nil && err != io.EOF {
				return fmt.Errorf("%s: %v\n", funcname[1], err)
			}
			if !bytes.Equal(gotb, wantb) {
				ReadCloserToFile(gotf, got)
				return fmt.Errorf("%s : got %q, want %q at %d", funcname[1], gotb, wantb, index)
			}
			index++
		} else if err != nil && err != io.EOF {
			return fmt.Errorf("%s : read from want failed: %v", funcname[1], err)
		}
	}
	// EOF on reference file has been reached, let us check the got buffer
	n, err = got.Read(gotb)
	if !bytes.Equal(gotb, wantb) && gotb[0] == '\n' {
		log.Printf("last LF returned: %d [%q]\n", n, gotb)
	}
	if err != io.EOF && n != 0 { // If EOF is not produced, file is too short
		ReadCloserToFile(gotf, got)
		gotInfo, _ := os.Stat(gotf)
		return fmt.Errorf("%s : got file is too short by %d", funcname[1], gotInfo.Size())
	} else if err != io.EOF && n == 0 {
		if err == nil {
			fmt.Printf("%d [%q]\n", n, gotb)
		} else {
			fmt.Printf("read %d [%q] with %v\n", n, gotb, err)
		}
	}
	return nil
}
