package syncer

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/robertgontarski/sync/internal/cli"
	"github.com/robertgontarski/sync/internal/fs"
	"github.com/robertgontarski/sync/internal/logger"
)

type Syncer struct {
	config *cli.Config
	logger *logger.Logger
}

func New(config *cli.Config, log *logger.Logger) *Syncer {
	return &Syncer{
		config: config,
		logger: log,
	}
}

func createFS(pathInfo fs.PathInfo, config *cli.Config) (fs.FileSystem, error) {
	if !pathInfo.IsRemote {
		return fs.NewLocalFS(), nil
	}

	username := pathInfo.User
	if username == "" {
		u, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("cannot determine current user: %w", err)
		}
		username = u.Username
	}

	return fs.NewSFTPFS(fs.SFTPConfig{
		User:         username,
		Host:         pathInfo.Host,
		Port:         config.Port,
		IdentityFile: config.IdentityFile,
		Password:     config.Password,
	})
}

// joinPath joins path elements using the appropriate separator for the filesystem.
func joinPath(filesystem fs.FileSystem, elem ...string) string {
	if _, ok := filesystem.(*fs.SFTPFS); ok {
		return path.Join(elem...)
	}
	return filepath.Join(elem...)
}

// relPath computes the relative path using the appropriate separator for the filesystem.
func relPath(filesystem fs.FileSystem, basepath, targpath string) (string, error) {
	if _, ok := filesystem.(*fs.SFTPFS); ok {
		// For SFTP, use forward-slash based relative path computation.
		if !strings.HasSuffix(basepath, "/") {
			basepath += "/"
		}
		if strings.HasPrefix(targpath, basepath) {
			return strings.TrimPrefix(targpath, basepath), nil
		}
		return "", fmt.Errorf("cannot make %s relative to %s", targpath, basepath)
	}
	return filepath.Rel(basepath, targpath)
}

func (s *Syncer) Sync() error {
	srcInfo := fs.ParsePath(s.config.SourceDir)
	dstInfo := fs.ParsePath(s.config.TargetDir)

	srcFS, err := createFS(srcInfo, s.config)
	if err != nil {
		return fmt.Errorf("source: %w", err)
	}
	defer srcFS.Close()

	dstFS, err := createFS(dstInfo, s.config)
	if err != nil {
		return fmt.Errorf("target: %w", err)
	}
	defer dstFS.Close()

	srcPath := srcInfo.Path
	dstPath := dstInfo.Path

	stat, err := srcFS.Stat(srcPath)
	if err != nil {
		return err
	}

	if !stat.IsDir {
		return os.ErrInvalid
	}

	if err := EnsureDir(dstFS, dstPath); err != nil {
		return err
	}

	if err := s.syncSource(srcFS, srcPath, dstFS, dstPath); err != nil {
		return err
	}

	if s.config.DeleteMissing {
		if err := s.deleteOrphans(srcFS, srcPath, dstFS, dstPath); err != nil {
			return err
		}
	}

	return nil
}

func (s *Syncer) syncSource(srcFS fs.FileSystem, srcRoot string, dstFS fs.FileSystem, dstRoot string) error {
	return srcFS.Walk(srcRoot, func(srcPath string, info fs.FileInfo, err error) error {
		if err != nil {
			s.logger.Error("failed to access %s: %v", srcPath, err)
			return nil
		}

		if info.IsDir {
			return nil
		}

		rel, err := relPath(srcFS, srcRoot, srcPath)
		if err != nil {
			s.logger.Error("failed to get relative path for %s: %v", srcPath, err)
			return nil
		}

		dstPath := joinPath(dstFS, dstRoot, rel)

		if _, err := dstFS.Stat(dstPath); err != nil {
			// File doesn't exist on destination - ensure parent dir and copy.
			dstDir := filepath.Dir(dstPath)
			if _, ok := dstFS.(*fs.SFTPFS); ok {
				dstDir = path.Dir(dstPath)
			}
			if err := EnsureDir(dstFS, dstDir); err != nil {
				s.logger.Error("failed to create directory for %s: %v", rel, err)
				return nil
			}
			s.logger.Info("copying %s", rel)
			if err := CopyFile(srcFS, srcPath, dstFS, dstPath); err != nil {
				s.logger.Error("failed to copy %s: %v", rel, err)
			}
			return nil
		}

		identical, err := CompareFiles(srcFS, srcPath, dstFS, dstPath, s.config.UseChecksum)
		if err != nil {
			s.logger.Error("failed to compare %s: %v", rel, err)
			return nil
		}

		if !identical {
			s.logger.Info("updating %s", rel)
			if err := CopyFile(srcFS, srcPath, dstFS, dstPath); err != nil {
				s.logger.Error("failed to update %s: %v", rel, err)
			}
		}

		return nil
	})
}

func (s *Syncer) deleteOrphans(srcFS fs.FileSystem, srcRoot string, dstFS fs.FileSystem, dstRoot string) error {
	return dstFS.Walk(dstRoot, func(dstPath string, info fs.FileInfo, err error) error {
		if err != nil {
			s.logger.Error("failed to access %s: %v", dstPath, err)
			return nil
		}

		if info.IsDir {
			return nil
		}

		rel, err := relPath(dstFS, dstRoot, dstPath)
		if err != nil {
			s.logger.Error("failed to get relative path for %s: %v", dstPath, err)
			return nil
		}

		srcPath := joinPath(srcFS, srcRoot, rel)

		if _, err := srcFS.Stat(srcPath); err != nil {
			s.logger.Info("deleting %s", rel)
			if err := dstFS.Remove(dstPath); err != nil {
				s.logger.Error("failed to delete %s: %v", rel, err)
			}
		}

		return nil
	})
}
