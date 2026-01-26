package main

import (
	"os"

	"github.com/robertgontarski/sync/internal/cli"
	"github.com/robertgontarski/sync/internal/logger"
	"github.com/robertgontarski/sync/internal/syncer"
)

func main() {
	config := cli.Parse()
	log := logger.New()

	s := syncer.New(config, log)

	if err := s.Sync(); err != nil {
		log.Error("Synchronization failed: %v", err)
		os.Exit(1)
	}

	log.Info("Synchronization completed")
}
