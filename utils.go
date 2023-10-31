// Package testingfiles provides primitives to use files as reference for testing
package testingfiles

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// OutputDir changes the default dir to the folder where reference files (want files) are stored.
// Only the base of the directory is expected. If found, change default directory to it.
// When not found, check if ../test contains the folder.
func OutputDir(s string) {
	ex, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if filepath.Base(ex) != s { // No need to change
		err = os.Chdir(filepath.Join(ex, s))
		if err != nil {
			err = os.Chdir(filepath.Join(filepath.Dir(ex), "test", s))
			if err != nil {
				panic(err) // subdirectory is probably missing
			}
		}
	}
}

// StringToFile produces a file named fname with the content
func StringToFile(fname string, content []byte) {
	wfile, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = wfile.Close()
		if err != nil {
			panic(err)
		}
	}()

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
	defer func() {
		err = wfile.Close()
		if err != nil {
			panic(err)
		}
	}()

	_, err = wfile.Write(content.Bytes())
	if err != nil {
		panic(err)
	}
}

// ReadCloserToFile creates a file named fname with the content
func ReadCloserToFile(fname string, content io.ReadCloser) error {
	wfile, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = wfile.Close()
		if err != nil {
			panic(err)
		}
	}()
	c, err := io.ReadAll(content)
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
	filew, err := os.Open(want)
	if err != nil {
		return err
	}
	defer func() {
		_ = filew.Close()
	}()

	fileg, err := os.Open(got)
	if err != nil {
		return err
	}
	defer func() {
		_ = filew.Close()
	}()

	bw, bg := make([]byte, 1), make([]byte, 1)
	index := 0          // Index in file to locate error
	for err != io.EOF { // Until the end of the file
		_, err = filew.Read(bw)
		if err != io.EOF { // While not EOF, read the other file too
			if err != nil { // there's still an error
				return err
			}
			_, err = fileg.Read(bg)
			if err != nil {
				wfileInfo, _ := filew.Stat()
				return fmt.Errorf("want file is larger by %d bytes", wfileInfo.Size()-int64(index))
			}
		}
		// Another byte was read from want file
		if !bytes.Equal(bw, bg) {
			return fmt.Errorf("got %q, want %q at %d", bw, bg, index)
		}
		index++
	}
	// EOF on reference (want) file has been reached.
	_, err = fileg.Read(bg)
	// If EOF is not returned, got file is larger than want file which has index-1 length
	if err != io.EOF {
		gfileInfo, _ := fileg.Stat()
		return fmt.Errorf("got file is larger by %d bytes", gfileInfo.Size()-int64(index-1))
	}
	// Both files which are identical
	return nil
}

// BufferCompare compares the buffer to a file.
// If an error occurs, got file is created and the error is returned.
// First char in the got file is the erroneous char.
// If identical, nil is returned.
// First byte index is 0
func BufferCompare(got *bytes.Buffer, want string) error {
	wantf, err := os.Open(want)
	if err != nil {
		return err
	}
	defer func() {
		_ = wantf.Close()
	}()

	// Build got filename.
	fileg := callerName("buffercomparedefault")
	b1 := make([]byte, 1)
	var b2 byte
	index := 0          // Index in file to locate error
	for err != io.EOF { // Until the end of the reference file (want)
		_, err = wantf.Read(b1)
		if err != io.EOF { // While not EOF, read the buffer
			if err != nil {
				return err // error on file was not io.EOF
			}

			b2, err = got.ReadByte()
			// Requires git 2.22.0 on Windows
			// If EOF is returned, buffer is too short and exhausted.
			if err != nil {
				wantfInfo, _ := wantf.Stat()
				// Last byte of the file is sometimes returned with io.EOF
				if wantfInfo.Size()-int64(index) == 1 && err == io.EOF {
					if b1[0] == b2 {
						log.Println("BufferCompare: last byte returned with io.EOF")
						// Overriding error
						return nil
					}
					// Occurs when original buffer is used
					return fmt.Errorf("got %v and last byte %q is missing", err, b1[0])
				}
				return fmt.Errorf("%s : got %v, want %q at %d. Buffer is missing %d",
					fileg, err, b1[0], index, wantfInfo.Size()-int64(index))
			}

			if b1[0] != b2 {
				_ = got.UnreadByte() // recover the erroneous char
				BufferToFile(fmt.Sprintf("got_%s", fileg), got)
				return fmt.Errorf("got %q, want %q at %d", b2, b1, index)
			}
			index++
		} else if err != nil && err != io.EOF {
			return fmt.Errorf("%s : read from want failed: %v", fileg, err)
		}
	}
	// EOF on want file has been reached
	_, err = got.ReadByte()
	// If EOF is not returned, buffer is longer than the file which is exhausted.
	if err != io.EOF {
		_ = got.UnreadByte()
		BufferToFile(fmt.Sprintf("got_%s", fileg), got)
		return fmt.Errorf("got buffer is too long by %d", got.Len())
	}
	return nil
}

// ReadCloserCompare compares a ReadCloser to a file.
// If an error occurs, got file is created and the error is returned.
// Last read byte is absent from the got file but available in the error message.
// If identical, nil is returned.
// Logic and method are identical to *buffer.Bytes but duplicating the code avoids ReadAll.
// First byte index is 0
// TODO Benchmark ReadAll against specific byte by byte code
func ReadCloserCompare(got io.ReadCloser, want string) error {
	wantf, err := os.Open(want)
	if err != nil {
		return err
	}
	defer func() {
		_ = wantf.Close()
	}()

	// Build got filename.
	fileg := callerName("readclosercomparedefault")
	gotf := fmt.Sprintf("got_%s", fileg)

	// Actual comparison
	wantb, gotb := make([]byte, 1), make([]byte, 1)
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
						log.Println("ReadCloserCompare: last byte returned with io.EOF")
						return nil
					}
				}
				return fmt.Errorf("%s : got %v, want %q at %d. Response is missing %d",
					fileg, err, wantb, index, wantfInfo.Size()-int64(index))
			} else if err != nil && err != io.EOF {
				return fmt.Errorf("%s: %v\n", fileg, err)
			}
			if !bytes.Equal(gotb, wantb) {
				_ = ReadCloserToFile(gotf, got)
				return fmt.Errorf("%s : got %q, want %q at %d", fileg, gotb, wantb, index)
			}
			index++
		} else if err != nil && err != io.EOF {
			return fmt.Errorf("%s : read from want failed: %v", fileg, err)
		}
	}
	// EOF on reference file has been reached, check the got buffer
	_, err = got.Read(gotb)
	// If EOF is not produced, response is longer than file
	if err != io.EOF {
		// The read byte of the response is not written to file
		err := ReadCloserToFile(gotf, got)
		if err == nil {
			gotInfo, _ := os.Stat(gotf)
			return fmt.Errorf("%s : got response is too long by %d. Last read byte %q", fileg, gotInfo.Size(), gotb)
		}
		return fmt.Errorf("%s : got response is too long. Writing file failed with %v", fileg, err)
	}
	return nil
}

// callerName returns the name of the function that called the testingfiles func.
// It returns the default if none is found.
func callerName(d string) (f string) {
	i, _, _, _ := runtime.Caller(2) // Skipping test and testingfile func
	// TODO Migrate to Caller name
	funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), ".")
	if len(funcname) == 1 {
		log.Printf("callerName: func name not found")
		return d // name is empty
	}
	return funcname[1]
}
