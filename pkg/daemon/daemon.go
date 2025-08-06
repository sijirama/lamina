package daemon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"lamina/pkg/indexer"
)

// Daemon manages the background indexing and watching process.
type Daemon struct {
	indexer *indexer.Indexer
}

// NewDaemon creates a new Daemon.
func NewDaemon() (*Daemon, error) {
	idx, err := indexer.NewIndexer()
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}
	return &Daemon{indexer: idx}, nil
}

// Start runs the daemon, handling signals for graceful shutdown and config reload.
func (d *Daemon) Start(ctx context.Context) error {
	// Start indexer (includes watcher)
	if err := d.indexer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start indexer: %w", err)
	}

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	for {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				fmt.Println("🛑 Stopping daemon")
				return nil
			case syscall.SIGHUP:
				fmt.Println("🔄 Reloading config")
				if err := d.indexer.watcher.Reload(); err != nil {
					fmt.Printf("⚠️ Failed to reload watcher: %v\n", err)
				}
			}
		case <-ctx.Done():
			fmt.Println("🛑 Stopping daemon")
			return nil
		}
	}
}
