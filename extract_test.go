package testingfiles

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

const (
	commonline = "pkg syscall, const ETHERTYPE_PAE = 34958"
	commonf    = "intersec.txt"
	globf      = "case_*.txt"
	// TODO TestMain sets output directory as default which is inconvenient
	path = "../testdata"
)

func TestExtractCommon(t *testing.T) {
	w := t.TempDir()
	if err := os.CopyFS(w, os.DirFS(path)); err != nil {
		t.Fatal(err)
	}
	if err := ExtractCommon(w, globf, commonf); err != nil {
		t.Error(err)
	}
	b, err := os.ReadFile(filepath.Join(w, commonf))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != commonline {
		t.Fatalf("got %s, want %s", string(b), commonline)
	}
}

func TestExtractCommon_empty(t *testing.T) {
	dir := t.TempDir()
	// One line and an empty file
	if err := os.WriteFile(filepath.Join(dir, "case_3.txt"), fmt.Appendf([]byte(commonline), "\n"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "case_empty.txt"), []byte(""), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	notcreated := "not_created"
	checkCommonFile := func() {
		err := ExtractCommon(dir, "case_*.txt", notcreated)
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("got %v, want %v", err, fs.ErrNotExist)
		}
		b, err := os.ReadFile(filepath.Join(dir, notcreated))
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("intersection file found when none is expected with %q", b)
		}
	}
	checkCommonFile()
	// an empty file and one line
	if err := os.Rename(filepath.Join(dir, "case_empty.txt"), filepath.Join(dir, "case_0.txt")); err != nil {
		t.Fatal(err)
	}
	checkCommonFile()
	// a differing line each
	if err := os.WriteFile(filepath.Join(dir, "case_one.txt"), []byte("nothing in common"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	checkCommonFile()
}

func TestExtractCommon_abridged(t *testing.T) {
	w := t.TempDir()
	c2, err := requiredFeatures(filepath.Join(path, "case_1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	// common part ends before the other file is exhausted
	want := 2
	c2 = c2[:want]
	err = os.CopyFS(w, os.DirFS(path))
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(w, "case_2.txt"), []byte(strings.Join(c2, "\n")), fs.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = ExtractCommon(w, globf, commonf)
	if err != nil {
		t.Fatal(err)
	}
	l, err1 := requiredFeatures(filepath.Join(w, commonf))
	if err1 != nil {
		t.Error(err)
	}
	if got := len(l); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExtractCommon_nofile(t *testing.T) {
	w := t.TempDir()
	intersec := "not_created"
	want := fs.ErrNotExist
	if err := ExtractCommon(w, "notfound", intersec); !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
	if _, err := os.ReadFile(filepath.Join(w, intersec)); !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
}

func TestExtractCommon_emptyfile(t *testing.T) {
	w := t.TempDir()
	intersec := "not_created"
	want := fs.ErrNotExist
	// second file is empty
	if err := os.WriteFile(filepath.Join(w, "case_1.txt"), []byte(commonline), fs.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(w, "case_2.txt"), nil, fs.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := ExtractCommon(w, globf, intersec); !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
	if _, err := os.ReadFile(filepath.Join(w, intersec)); !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
	// two files with the same common line
	if err := os.WriteFile(filepath.Join(w, "case_2.txt"), []byte(commonline), fs.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := ExtractCommon(w, globf, ""); !errors.Is(err, syscall.EISDIR) {
		t.Errorf("got %v, want %v", err, want)
	}
	if _, err := os.ReadFile(filepath.Join(w, intersec)); !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
}

func TestExtractCommon_direrror(t *testing.T) {
	dir := t.TempDir()[3:]
	intersec := "failed_dir"
	want := fs.ErrNotExist
	if err := ExtractCommon(dir, "notfound", intersec); !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
	if _, err := os.ReadFile(filepath.Join(dir, intersec)); !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
}

func TestCreateSupplement(t *testing.T) {
	w := t.TempDir()
	if err := os.CopyFS(w, os.DirFS(path)); err != nil {
		t.Fatal(err)
	}
	b, err := requiredFeatures(filepath.Join(w, "case_1.txt"))
	if err != nil {
		t.Error(err)
	}
	want := len(b) - 1
	if err = ExtractCommon(w, globf, commonf); err != nil {
		t.Error(err)
	}
	if err = CreateSupplements(w, globf, commonf); err != nil {
		t.Fatal(err)
	}
	b, err = requiredFeatures(filepath.Join(w, "case_1.txt"))
	if err != nil {
		t.Error(err)
	}
	if got := len(b); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCreateSupplement_onefile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, commonf), nil, fs.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := CreateSupplements(dir, commonf, commonf); err == nil {
		t.Fatal("unexpected success")
	}
	if err := os.WriteFile(filepath.Join(dir, commonf), []byte(commonline), fs.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := CreateSupplements(dir, commonf, commonf); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, commonf))
	if err != nil {
		t.Fatal(err)
	}
	// common line is gone
	if got, want := len(b), 0; got != want {
		t.Fatalf("length of intersection: got %v, want %v", got, want)
	}
	// file is empty
	if err = os.WriteFile(filepath.Join(dir, commonf), []byte(commonline), fs.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "case_1.txt"), nil, fs.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err = CreateSupplements(dir, globf, commonf); err == nil {
		t.Error("unexpected success")
	}
	if want := os.ErrNotExist; !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
	// common line is not known to file which is before common line
	if err = os.WriteFile(filepath.Join(dir, "case_1.txt"), []byte("an unknown line"), fs.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err = CreateSupplements(dir, globf, commonf); err == nil {
		t.Fatal("unexpected success")
	}
	want := errors.New("common feature is missing: pkg syscall, const ETHERTYPE_PAE = 34958")
	if errors.Is(want, err) {
		t.Errorf("got %v, want %v", err, want)
	}
	// common line is not known to file which is after common line
	if err = os.WriteFile(filepath.Join(dir, "case_1.txt"), []byte("software"), fs.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err = CreateSupplements(dir, globf, commonf); err == nil {
		t.Fatal("unexpected success")
	}
	if errors.Is(want, err) {
		t.Errorf("got %v, want %v", err, want)
	}
	// common line is not known to file which is after common line
	want = fs.ErrNotExist
	if err = os.WriteFile(filepath.Join(dir, "case_1.txt"), []byte(commonline), fs.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(dir, 0222); err != nil {
		t.Error(err)
	}
	if err = CreateSupplements(dir, globf, commonf); !os.IsNotExist(err) {
		t.Errorf("got %v, want %v", err, want)
	}
	if err = os.Chmod(dir, fs.ModePerm); err != nil {
		t.Error(err)
	}
}

func TestCreateSupplement_globerror(t *testing.T) {
	t.Skip("no globf pattern to trigger error found for now")
	dir := ""
	err := CreateSupplements(dir, "+", commonf)
	if want := filepath.ErrBadPattern; !errors.Is(err, want) {
		t.Fatalf("%v: got %v, want %v", dir, err, want)
	}
}

func TestCreateSupplement_fileerror(t *testing.T) {
	dir := t.TempDir()
	fn := "not_exist"
	err := CreateSupplements(dir, fn, commonf)
	if want := fs.ErrNotExist; !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}
