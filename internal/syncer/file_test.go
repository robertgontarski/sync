package syncer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCopyFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test.txt")
	dstPath := filepath.Join(dstDir, "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	if err := CopyFile(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", string(dstContent), string(content))
	}
}

func TestCopyFileCreatesDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test.txt")
	dstPath := filepath.Join(dstDir, "subdir", "nested", "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	if err := CopyFile(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("destination file was not created")
	}
}

func TestCompareByMetadata_Identical(t *testing.T) {
	dir := t.TempDir()

	path1 := filepath.Join(dir, "file1.txt")
	path2 := filepath.Join(dir, "file2.txt")

	content := []byte("same content")
	modTime := time.Now().Truncate(time.Second)

	if err := os.WriteFile(path1, content, 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(path2, content, 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	if err := os.Chtimes(path1, modTime, modTime); err != nil {
		t.Fatalf("failed to set modtime for file1: %v", err)
	}
	if err := os.Chtimes(path2, modTime, modTime); err != nil {
		t.Fatalf("failed to set modtime for file2: %v", err)
	}

	identical, err := CompareByMetadata(path1, path2)
	if err != nil {
		t.Fatalf("CompareByMetadata failed: %v", err)
	}
	if !identical {
		t.Error("files should be identical")
	}
}

func TestCompareByMetadata_DifferentSize(t *testing.T) {
	dir := t.TempDir()

	path1 := filepath.Join(dir, "file1.txt")
	path2 := filepath.Join(dir, "file2.txt")

	if err := os.WriteFile(path1, []byte("short"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(path2, []byte("longer content"), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	identical, err := CompareByMetadata(path1, path2)
	if err != nil {
		t.Fatalf("CompareByMetadata failed: %v", err)
	}
	if identical {
		t.Error("files should not be identical")
	}
}

func TestCompareByMetadata_DifferentModTime(t *testing.T) {
	dir := t.TempDir()

	path1 := filepath.Join(dir, "file1.txt")
	path2 := filepath.Join(dir, "file2.txt")

	content := []byte("same content")
	if err := os.WriteFile(path1, content, 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(path2, content, 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	modTime1 := time.Now().Truncate(time.Second)
	modTime2 := modTime1.Add(-time.Hour)

	if err := os.Chtimes(path1, modTime1, modTime1); err != nil {
		t.Fatalf("failed to set modtime for file1: %v", err)
	}
	if err := os.Chtimes(path2, modTime2, modTime2); err != nil {
		t.Fatalf("failed to set modtime for file2: %v", err)
	}

	identical, err := CompareByMetadata(path1, path2)
	if err != nil {
		t.Fatalf("CompareByMetadata failed: %v", err)
	}
	if identical {
		t.Error("files should not be identical")
	}
}

func TestCompareByChecksum_Identical(t *testing.T) {
	dir := t.TempDir()

	path1 := filepath.Join(dir, "file1.txt")
	path2 := filepath.Join(dir, "file2.txt")

	content := []byte("same content")
	if err := os.WriteFile(path1, content, 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(path2, content, 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	identical, err := CompareByChecksum(path1, path2)
	if err != nil {
		t.Fatalf("CompareByChecksum failed: %v", err)
	}
	if !identical {
		t.Error("files should be identical")
	}
}

func TestCompareByChecksum_Different(t *testing.T) {
	dir := t.TempDir()

	path1 := filepath.Join(dir, "file1.txt")
	path2 := filepath.Join(dir, "file2.txt")

	if err := os.WriteFile(path1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(path2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	identical, err := CompareByChecksum(path1, path2)
	if err != nil {
		t.Fatalf("CompareByChecksum failed: %v", err)
	}
	if identical {
		t.Error("files should not be identical")
	}
}

func TestCalculateChecksum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	checksum, err := CalculateChecksum(path)
	if err != nil {
		t.Fatalf("CalculateChecksum failed: %v", err)
	}

	expected := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if checksum != expected {
		t.Errorf("checksum mismatch: got %s, want %s", checksum, expected)
	}
}

func TestEnsureDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c")

	if err := EnsureDir(path); err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("path should be a directory")
	}
}

func TestEnsureDir_Existing(t *testing.T) {
	dir := t.TempDir()

	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir failed on existing directory: %v", err)
	}
}
