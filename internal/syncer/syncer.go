package syncer

import (
	"os"
	"path/filepath"

	"github.com/robertgontarski/sync/internal/cli"
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

func (s *Syncer) Sync() error {
	srcInfo, err := os.Stat(s.config.SourceDir)
	if err != nil {
		return err
	}

	if !srcInfo.IsDir() {
		return os.ErrInvalid
	}

	if err := EnsureDir(s.config.TargetDir); err != nil {
		return err
	}

	if err := s.syncSource(); err != nil {
		return err
	}

	if s.config.DeleteMissing {
		if err := s.deleteOrphans(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Syncer) syncSource() error {
	return filepath.Walk(s.config.SourceDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			s.logger.Error("failed to access %s: %v", srcPath, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(s.config.SourceDir, srcPath)
		if err != nil {
			s.logger.Error("failed to get relative path for %s: %v", srcPath, err)
			return nil
		}

		dstPath := filepath.Join(s.config.TargetDir, relPath)

		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			s.logger.Info("copying %s", relPath)
			if err := CopyFile(srcPath, dstPath); err != nil {
				s.logger.Error("failed to copy %s: %v", relPath, err)
			}
			return nil
		}

		identical, err := CompareFiles(srcPath, dstPath, s.config.UseChecksum)
		if err != nil {
			s.logger.Error("failed to compare %s: %v", relPath, err)
			return nil
		}

		if !identical {
			s.logger.Info("updating %s", relPath)
			if err := CopyFile(srcPath, dstPath); err != nil {
				s.logger.Error("failed to update %s: %v", relPath, err)
			}
		}

		return nil
	})
}

func (s *Syncer) deleteOrphans() error {
	return filepath.Walk(s.config.TargetDir, func(dstPath string, info os.FileInfo, err error) error {
		if err != nil {
			s.logger.Error("failed to access %s: %v", dstPath, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(s.config.TargetDir, dstPath)
		if err != nil {
			s.logger.Error("failed to get relative path for %s: %v", dstPath, err)
			return nil
		}

		srcPath := filepath.Join(s.config.SourceDir, relPath)

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			s.logger.Info("deleting %s", relPath)
			if err := os.Remove(dstPath); err != nil {
				s.logger.Error("failed to delete %s: %v", relPath, err)
			}
		}

		return nil
	})
}
