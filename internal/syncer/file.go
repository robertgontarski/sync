package syncer

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

func CopyFile(src, dst string) error {
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	if err := dstFile.Chmod(srcInfo.Mode()); err != nil {
		return err
	}

	return os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
}

func CompareFiles(src, dst string, useChecksum bool) (bool, error) {
	if useChecksum {
		return CompareByChecksum(src, dst)
	}

	return CompareByMetadata(src, dst)
}

func CompareByMetadata(src, dst string) (bool, error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}

	dstInfo, err := os.Stat(dst)
	if err != nil {
		return false, err
	}

	if srcInfo.Size() != dstInfo.Size() {
		return false, nil
	}

	srcModTime := srcInfo.ModTime().Truncate(1e9)
	dstModTime := dstInfo.ModTime().Truncate(1e9)

	return srcModTime.Equal(dstModTime), nil
}

func CompareByChecksum(src, dst string) (bool, error) {
	srcChecksum, err := CalculateChecksum(src)
	if err != nil {
		return false, err
	}

	dstChecksum, err := CalculateChecksum(dst)
	if err != nil {
		return false, err
	}

	return srcChecksum == dstChecksum, nil
}

func CalculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
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

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
