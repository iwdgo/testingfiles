package testingfiles

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ExtractCommon creates a file commonf containing all lines found in all files selected by globf filter.
// It returns an error if the intersection is empty.
func ExtractCommon(contextFilesPath, globf, commonf string) error {
	fl, err := filepath.Glob(filepath.Join(contextFilesPath, globf))
	if err != nil {
		return err
	}
	if len(fl) == 0 {
		return fs.ErrNotExist
	}
	intersec, err := requiredFeatures(fl[0])
	if err != nil {
		return err
	}
	fl = fl[1:]
	i := 0
	for _, f := range fl {
		flines, err1 := requiredFeatures(f)
		if err1 != nil {
			return err1
		}
		for {
			if intersec[i] == flines[0] {
				i++
				flines = flines[1:]
			} else if intersec[i] < flines[0] {
				// feature is unknown in checked file and must be removed from intersection
				intersec = slices.Delete(intersec, i, i+1)
			} else {
				// skip feature unknown to intersection
				flines = flines[1:]
			}
			if i == len(intersec) {
				break
			} else if len(flines) == 0 {
				intersec = intersec[:i]
				break
			}
		}
		if len(intersec) == 0 {
			log.Printf("intersection is empty as %s has nothing in common with other files", filepath.Base(f))
			return os.ErrNotExist
		}
		i = 0
	}
	if err = os.WriteFile(filepath.Join(contextFilesPath, commonf),
		[]byte(strings.Join(intersec, "\n")), os.ModePerm); err != nil {
		return err
	}
	log.Printf("intersection has %v line(s) written to %s", len(intersec), commonf)
	return nil
}

// CreateSupplements removes all lines of a baselinef file from all files in globf
func CreateSupplements(contextFilesPath, globf, baselinef string) error {
	fl, err := filepath.Glob(filepath.Join(contextFilesPath, globf))
	if err != nil {
		return err
	}
	if len(fl) == 0 {
		return fs.ErrNotExist
	}
	baseline, err := requiredFeatures(filepath.Join(contextFilesPath, baselinef))
	if err != nil {
		return err
	}
	base := baseline
	for _, f := range fl {
		flines, err1 := requiredFeatures(f)
		if err1 != nil {
			return err1
		}
		var supplement []string
		for {
			if base[0] == flines[0] {
				base = base[1:]
				flines = flines[1:]
			} else if base[0] > flines[0] {
				supplement = append(supplement, flines[0])
				flines = flines[1:]
				if len(flines) == 0 {
					return errors.New(fmt.Sprintf("common feature is missing: %s", base[0]))
				}
			} else {
				return errors.New(fmt.Sprintf("common feature is missing: %s", base[0]))
			}
			if len(base) == 0 {
				break
			}
		}
		err = os.WriteFile(f, []byte(strings.Join(supplement, "\n")), os.ModePerm)
		if err != nil {
			return err
		}
		log.Printf("%s has %v line(s)", f, len(supplement))
		base = baseline
	}
	return nil
}

func requiredFeatures(filename string) ([]string, error) {
	bs, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	bss := string(bs)
	//  TODO CRLF line ending is not handled
	lines := strings.Split(bss, "\n")
	// A last empty line is removed as it is usually an artifact
	if strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		log.Printf("%s is empty", filename)
		return nil, fs.ErrNotExist
	}
	slices.Sort(lines)
	log.Printf("%s has %v lines", filename, len(lines))
	return lines, nil
}
