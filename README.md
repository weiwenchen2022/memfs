# memfs: A simple in-memory io/fs.FS filesystem

memfs is an in-memory implementation of Go's io/fs.FS interface.
The goal is to make it easy and quick to build an fs.FS filesystem
when you don't have any complex requirements.

`io/fs` docs: https://tip.golang.org/pkg/io/fs/

## Usage

```
package main

import (
	"fmt"
	"log"
	"io/fs"

	"memfs"
)

func main() {
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
}
```
