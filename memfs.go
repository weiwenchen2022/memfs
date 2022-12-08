package memfs

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

// implement fs.ReadDirFile
type dir struct {
	*dirEntry

	entries map[string]fs.DirEntry

	names []string
}

var _ fs.ReadDirFile = (*dir)(nil)

func (d *dir) Read(p []byte) (int, error) {
	return 0, &fs.PathError{
		Op:   "read",
		Path: d.name,
		Err:  errors.New("is directory"),
	}
}

func (d *dir) Stat() (fs.FileInfo, error) {
	return d.dirEntry.Info()
}

func (d *dir) Close() error {
	return nil
}

func (d *dir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.names == nil {
		d.names = make([]string, 0, len(d.entries))
		for name := range d.entries {
			d.names = append(d.names, name)
		}
	}

	if n <= 0 {
		n = len(d.names)
	}

	entries := make([]fs.DirEntry, 0, n)
	for i := 0; i < n && i < len(d.names); i++ {
		name := d.names[i]
		entries = append(entries, d.entries[name])
	}

	if n < len(d.names) {
		d.names = d.names[n:]
	} else {
		d.names = nil
	}

	return entries, nil
}

// implement fs.FileInfo
type fileInfo struct {
	name    string
	size    int64
	modTime time.Time
	mode    fs.FileMode
}

var _ fs.FileInfo = (*fileInfo)(nil)

func (fi *fileInfo) Name() string {
	return fi.name
}

func (fi *fileInfo) Size() int64 {
	return fi.size
}

func (fi *fileInfo) Mode() fs.FileMode {
	return fi.mode
}

func (fi *fileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi *fileInfo) IsDir() bool {
	return fi.Mode().IsDir()
}

func (fi *fileInfo) Sys() interface{} {
	return nil
}

// Implements fs.DirEntry
type dirEntry struct {
	*fileInfo
}

var _ fs.DirEntry = (*dirEntry)(nil)

func (de *dirEntry) Type() fs.FileMode {
	return de.Mode().Type()
}

func (de *dirEntry) Info() (fs.FileInfo, error) {
	return de.fileInfo, nil
}

// implement fs.File
type file struct {
	*dirEntry

	content *bytes.Buffer
	closed  bool
}

var _ fs.File = (*file)(nil)

func (f *file) Read(p []byte) (int, error) {
	if f.closed {
		return 0, fs.ErrClosed
	}

	return f.content.Read(p)
}

func (f *file) Stat() (fs.FileInfo, error) {
	if f.closed {
		return nil, fs.ErrClosed
	}

	info, _ := f.dirEntry.Info()
	fi := info.(*fileInfo)
	fi.size = int64(f.content.Len())

	return fi, nil
}

func (f *file) Close() error {
	if f.closed {
		return fs.ErrClosed
	}

	f.closed = true
	return nil
}

// FS is an in-memory filesystem that implements
// io/fs.FS
type FS struct {
	root *dir
}

var _ fs.FS = (*FS)(nil)

// New creates a new in-memory FileSystem.
func New() *FS {
	return &FS{
		root: &dir{
			dirEntry: &dirEntry{
				fileInfo: &fileInfo{
					name:    ".",
					size:    0,
					modTime: time.Now(),
					mode:    fs.ModeDir | 0644,
				},
			},

			entries: make(map[string]fs.DirEntry),
		},
	}
}

// Open opens the named file.
func (fsys *FS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	if name == "." || name == "" {
		return fsys.root, nil
	}

	var cur *dir = fsys.root
	parts := strings.Split(name, "/")

	for i, part := range parts {
		entry := cur.entries[part]
		if entry == nil {
			goto errNotExist
		}

		f, ok := entry.(*file)
		if ok {
			if i == len(parts)-1 {
				return &file{
					dirEntry: f.dirEntry,
					content:  bytes.NewBuffer(f.content.Bytes()),
				}, nil
			}

			goto errNotExist
		}

		d, ok := entry.(*dir)
		if !ok {
			goto errNotExist
		}

		cur = d
	}

	return cur, nil

errNotExist:
	return nil, &fs.PathError{
		Op:   "open",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

// MkdirAll creates a directory named path,
// along with any necessary parents, and returns nil,
// or else returns an error.
// If path is already a directory, MkdirAll does nothing
// and returns nil.
func (fsys *FS) MkdirAll(path string) error {
	if !fs.ValidPath(path) {
		return fs.ErrInvalid
	}

	if path == "." {
		return nil
	}

	var cur *dir = fsys.root
	parts := strings.Split(path, "/")
	for _, part := range parts {
		entry := cur.entries[part]
		if entry == nil {
			entry = &dir{
				dirEntry: &dirEntry{
					fileInfo: &fileInfo{
						name:    part,
						size:    0,
						modTime: time.Now(),
						mode:    fs.ModeDir | 0644,
					},
				},

				entries: make(map[string]fs.DirEntry),
			}

			cur.entries[part] = entry
		} else {
			if _, ok := entry.(*dir); !ok {
				return fmt.Errorf("%q is not directory", part)
			}
		}

		cur = entry.(*dir)
	}

	return nil
}

// WriteFile writes data to a file named by filename.
// If the file does not exist, WriteFile creates it;
// otherwise WriteFile truncates it before writing.
func (fsys *FS) WriteFile(name string, data []byte) error {
	if !fs.ValidPath(name) {
		return &fs.PathError{
			Op:   "write",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	var err error
	var dir *dir = fsys.root
	path := filepath.Dir(name)
	if path != "." {
		dir, err = fsys.getDir(path)
		if err != nil {
			return err
		}
	}

	buf := make([]byte, len(data))
	copy(buf, data)

	filename := filepath.Base(name)
	dir.entries[filename] = &file{
		dirEntry: &dirEntry{
			fileInfo: &fileInfo{
				name:    filename,
				size:    int64(len(data)),
				modTime: time.Now(),
				mode:    0644,
			},
		},

		content: bytes.NewBuffer(buf),
	}

	return nil
}

func (fsys *FS) getDir(path string) (*dir, error) {
	parts := strings.Split(path, "/")

	var cur *dir = fsys.root
	for _, part := range parts {
		entry := cur.entries[part]
		if entry == nil {
			return nil, fmt.Errorf("%q is not exist %w",
				part, fs.ErrNotExist)
		}

		var ok bool
		cur, ok = entry.(*dir)
		if !ok {
			return nil, fmt.Errorf("%q is not directory %w",
				part, fs.ErrNotExist)
		}
	}

	return cur, nil
}
