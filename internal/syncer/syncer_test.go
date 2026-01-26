package syncer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/robertgontarski/sync/internal/cli"
	"github.com/robertgontarski/sync/internal/logger"
)

func setupTest(t *testing.T) (string, string, *bytes.Buffer) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	logBuf := &bytes.Buffer{}
	return srcDir, dstDir, logBuf
}

func createFile(t *testing.T, path string, content string) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}
	return string(content)
}

func TestSync_NewFiles(t *testing.T) {
	srcDir, dstDir, logBuf := setupTest(t)

	createFile(t, filepath.Join(srcDir, "file1.txt"), "content1")
	createFile(t, filepath.Join(srcDir, "subdir", "file2.txt"), "content2")

	config := &cli.Config{
		SourceDir:     srcDir,
		TargetDir:     dstDir,
		DeleteMissing: false,
		UseChecksum:   false,
	}

	s := New(config, logger.NewWithWriter(logBuf))
	if err := s.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if content := readFile(t, filepath.Join(dstDir, "file1.txt")); content != "content1" {
		t.Errorf("file1.txt content mismatch: got %q, want %q", content, "content1")
	}
	if content := readFile(t, filepath.Join(dstDir, "subdir", "file2.txt")); content != "content2" {
		t.Errorf("subdir/file2.txt content mismatch: got %q, want %q", content, "content2")
	}
}

func TestSync_UpdateChangedFiles(t *testing.T) {
	srcDir, dstDir, logBuf := setupTest(t)

	createFile(t, filepath.Join(srcDir, "file.txt"), "new content")
	createFile(t, filepath.Join(dstDir, "file.txt"), "old content")

	oldTime := time.Now().Add(-time.Hour).Truncate(time.Second)
	if err := os.Chtimes(filepath.Join(dstDir, "file.txt"), oldTime, oldTime); err != nil {
		t.Fatalf("failed to set modtime: %v", err)
	}

	config := &cli.Config{
		SourceDir:     srcDir,
		TargetDir:     dstDir,
		DeleteMissing: false,
		UseChecksum:   false,
	}

	s := New(config, logger.NewWithWriter(logBuf))
	if err := s.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if content := readFile(t, filepath.Join(dstDir, "file.txt")); content != "new content" {
		t.Errorf("file.txt content mismatch: got %q, want %q", content, "new content")
	}
}

func TestSync_SkipIdenticalFiles(t *testing.T) {
	srcDir, dstDir, logBuf := setupTest(t)

	content := "same content"
	modTime := time.Now().Truncate(time.Second)

	createFile(t, filepath.Join(srcDir, "file.txt"), content)
	createFile(t, filepath.Join(dstDir, "file.txt"), content)

	if err := os.Chtimes(filepath.Join(srcDir, "file.txt"), modTime, modTime); err != nil {
		t.Fatalf("failed to set modtime: %v", err)
	}
	if err := os.Chtimes(filepath.Join(dstDir, "file.txt"), modTime, modTime); err != nil {
		t.Fatalf("failed to set modtime: %v", err)
	}

	config := &cli.Config{
		SourceDir:     srcDir,
		TargetDir:     dstDir,
		DeleteMissing: false,
		UseChecksum:   false,
	}

	s := New(config, logger.NewWithWriter(logBuf))
	if err := s.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if bytes.Contains(logBuf.Bytes(), []byte("copying")) || bytes.Contains(logBuf.Bytes(), []byte("updating")) {
		t.Error("identical file should not be copied or updated")
	}
}

func TestSync_DeleteMissing(t *testing.T) {
	srcDir, dstDir, logBuf := setupTest(t)

	createFile(t, filepath.Join(srcDir, "keep.txt"), "keep")
	createFile(t, filepath.Join(dstDir, "keep.txt"), "keep")
	createFile(t, filepath.Join(dstDir, "orphan.txt"), "orphan")

	modTime := time.Now().Truncate(time.Second)
	if err := os.Chtimes(filepath.Join(srcDir, "keep.txt"), modTime, modTime); err != nil {
		t.Fatalf("failed to set modtime: %v", err)
	}
	if err := os.Chtimes(filepath.Join(dstDir, "keep.txt"), modTime, modTime); err != nil {
		t.Fatalf("failed to set modtime: %v", err)
	}

	config := &cli.Config{
		SourceDir:     srcDir,
		TargetDir:     dstDir,
		DeleteMissing: true,
		UseChecksum:   false,
	}

	s := New(config, logger.NewWithWriter(logBuf))
	if err := s.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dstDir, "keep.txt")); os.IsNotExist(err) {
		t.Error("keep.txt should exist")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "orphan.txt")); !os.IsNotExist(err) {
		t.Error("orphan.txt should be deleted")
	}
}

func TestSync_DeleteMissingDisabled(t *testing.T) {
	srcDir, dstDir, logBuf := setupTest(t)

	createFile(t, filepath.Join(srcDir, "keep.txt"), "keep")
	createFile(t, filepath.Join(dstDir, "orphan.txt"), "orphan")

	config := &cli.Config{
		SourceDir:     srcDir,
		TargetDir:     dstDir,
		DeleteMissing: false,
		UseChecksum:   false,
	}

	s := New(config, logger.NewWithWriter(logBuf))
	if err := s.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dstDir, "orphan.txt")); os.IsNotExist(err) {
		t.Error("orphan.txt should not be deleted when DeleteMissing is false")
	}
}

func TestSync_UseChecksum(t *testing.T) {
	srcDir, dstDir, logBuf := setupTest(t)

	createFile(t, filepath.Join(srcDir, "file.txt"), "same content")
	createFile(t, filepath.Join(dstDir, "file.txt"), "same content")

	config := &cli.Config{
		SourceDir:     srcDir,
		TargetDir:     dstDir,
		DeleteMissing: false,
		UseChecksum:   true,
	}

	s := New(config, logger.NewWithWriter(logBuf))
	if err := s.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if bytes.Contains(logBuf.Bytes(), []byte("updating")) {
		t.Error("files with same checksum should not be updated")
	}
}

func TestSync_SourceNotExists(t *testing.T) {
	_, dstDir, logBuf := setupTest(t)

	config := &cli.Config{
		SourceDir:     "/nonexistent/path",
		TargetDir:     dstDir,
		DeleteMissing: false,
		UseChecksum:   false,
	}

	s := New(config, logger.NewWithWriter(logBuf))
	if err := s.Sync(); err == nil {
		t.Error("Sync should fail when source does not exist")
	}
}

func TestSync_SourceIsFile(t *testing.T) {
	srcDir, dstDir, logBuf := setupTest(t)

	srcFile := filepath.Join(srcDir, "file.txt")
	createFile(t, srcFile, "content")

	config := &cli.Config{
		SourceDir:     srcFile,
		TargetDir:     dstDir,
		DeleteMissing: false,
		UseChecksum:   false,
	}

	s := New(config, logger.NewWithWriter(logBuf))
	if err := s.Sync(); err == nil {
		t.Error("Sync should fail when source is a file")
	}
}

func TestSync_CreatesTargetDir(t *testing.T) {
	srcDir, dstDir, logBuf := setupTest(t)

	createFile(t, filepath.Join(srcDir, "file.txt"), "content")

	newTarget := filepath.Join(dstDir, "new", "nested", "target")

	config := &cli.Config{
		SourceDir:     srcDir,
		TargetDir:     newTarget,
		DeleteMissing: false,
		UseChecksum:   false,
	}

	s := New(config, logger.NewWithWriter(logBuf))
	if err := s.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(newTarget, "file.txt")); os.IsNotExist(err) {
		t.Error("file should be synced to new target directory")
	}
}
