package fs

import (
	"io"
	"os"
	"time"
)

type FileInfo struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
}

type WalkFunc func(path string, info FileInfo, err error) error

type FileSystem interface {
	Stat(path string) (FileInfo, error)
	Walk(root string, fn WalkFunc) error
	Open(path string) (io.ReadCloser, error)
	Create(path string) (io.WriteCloser, error)
	Remove(path string) error
	MkdirAll(path string, perm os.FileMode) error
	Chmod(path string, mode os.FileMode) error
	Chtimes(path string, atime, mtime time.Time) error
	Close() error
}
