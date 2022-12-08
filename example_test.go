package memfs_test

import (
	"fmt"
	"io/fs"
	"log"

	"memfs"
)

func ExampleMemFS() {
	rootFS := memfs.New()

	err := rootFS.MkdirAll("dir1/dir2")
	if err != nil {
		log.Fatal(err)
	}

	err = rootFS.WriteFile("dir1/dir2/f1.txt", []byte("incinerating-unsubstantial"))
	if err != nil {
		log.Fatal(err)
	}

	err = fs.WalkDir(rootFS, ".", func(path string, d fs.DirEntry, err error) error {
		fmt.Println(path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	content, err := fs.ReadFile(rootFS, "dir1/dir2/f1.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", content)
	// Output:
	// 	.
	// dir1
	// dir1/dir2
	// dir1/dir2/f1.txt
	// incinerating-unsubstantial
}
