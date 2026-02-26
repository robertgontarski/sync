package syncer

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"github.com/robertgontarski/sync/internal/fs"
)

func CopyFile(srcFS fs.FileSystem, srcPath string, dstFS fs.FileSystem, dstPath string) error {
	srcInfo, err := srcFS.Stat(srcPath)
	if err != nil {
		return err
	}

	srcFile, err := srcFS.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := dstFS.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	if err := dstFS.Chmod(dstPath, srcInfo.Mode); err != nil {
		return err
	}

	return dstFS.Chtimes(dstPath, srcInfo.ModTime, srcInfo.ModTime)
}

func CompareFiles(srcFS fs.FileSystem, srcPath string, dstFS fs.FileSystem, dstPath string, useChecksum bool) (bool, error) {
	if useChecksum {
		return CompareByChecksum(srcFS, srcPath, dstFS, dstPath)
	}

	return CompareByMetadata(srcFS, srcPath, dstFS, dstPath)
}

func CompareByMetadata(srcFS fs.FileSystem, srcPath string, dstFS fs.FileSystem, dstPath string) (bool, error) {
	srcInfo, err := srcFS.Stat(srcPath)
	if err != nil {
		return false, err
	}

	dstInfo, err := dstFS.Stat(dstPath)
	if err != nil {
		return false, err
	}

	if srcInfo.Size != dstInfo.Size {
		return false, nil
	}

	srcModTime := srcInfo.ModTime.Truncate(1e9)
	dstModTime := dstInfo.ModTime.Truncate(1e9)

	return srcModTime.Equal(dstModTime), nil
}

func CompareByChecksum(srcFS fs.FileSystem, srcPath string, dstFS fs.FileSystem, dstPath string) (bool, error) {
	srcChecksum, err := CalculateChecksum(srcFS, srcPath)
	if err != nil {
		return false, err
	}

	dstChecksum, err := CalculateChecksum(dstFS, dstPath)
	if err != nil {
		return false, err
	}

	return srcChecksum == dstChecksum, nil
}

func CalculateChecksum(filesystem fs.FileSystem, path string) (string, error) {
	file, err := filesystem.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func EnsureDir(filesystem fs.FileSystem, path string) error {
	return filesystem.MkdirAll(path, os.FileMode(0755))
}
