package memfs_test

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"

	"memfs"

	"github.com/google/go-cmp/cmp"
)

func TestMemFS(t *testing.T) {
	rootFS := memfs.New()
	err := rootFS.MkdirAll("foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	var gotPaths []string
	err = fs.WalkDir(rootFS, ".", func(path string, d fs.DirEntry, err error) error {
		gotPaths = append(gotPaths, path)

		if !d.IsDir() {
			return fmt.Errorf("%q is not a directory", path)
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	expectedPaths := []string{
		".",
		"foo",
		"foo/bar",
	}

	if diff := cmp.Diff(expectedPaths, gotPaths); diff != "" {
		t.Fatalf("WalkDir mismatch %s", diff)
	}

	err = rootFS.WriteFile("foo/baz/buz.txt", []byte("buz"))
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Expected missing directory error but got none")
	}

	_, err = fs.ReadFile(rootFS, "foo/baz/buz.txt")
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("Expected no such file but got no error")
	}

	data := []byte("baz")
	err = rootFS.WriteFile("foo/bar/baz.txt", data)
	if err != nil {
		t.Fatal(err)
	}

	content, err := fs.ReadFile(rootFS, "foo/bar/baz.txt")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(data, content); diff != "" {
		t.Fatalf("write/read baz.txt mismatch %s", diff)
	}
}
