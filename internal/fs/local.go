package fs

import (
	"io"
	"os"
	"path/filepath"
	"time"
)

type LocalFS struct{}

func NewLocalFS() *LocalFS {
	return &LocalFS{}
}

func (l *LocalFS) Stat(path string) (FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}
	return FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

func (l *LocalFS) Walk(root string, fn WalkFunc) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fn(path, FileInfo{}, err)
		}
		return fn(path, FileInfo{
			Name:    info.Name(),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		}, nil)
	})
}

func (l *LocalFS) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (l *LocalFS) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

func (l *LocalFS) Remove(path string) error {
	return os.Remove(path)
}

func (l *LocalFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (l *LocalFS) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

func (l *LocalFS) Chtimes(path string, atime, mtime time.Time) error {
	return os.Chtimes(path, atime, mtime)
}

func (l *LocalFS) Close() error {
	return nil
}
