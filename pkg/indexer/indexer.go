package indexer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/tmc/langchaingo/embeddings"
	"lamina/pkg/config"
	"lamina/pkg/watcher"
)

// Indexer manages file indexing.
type Indexer struct {
	watcher   *watcher.FileWatcher
	embedder  *embeddings.Embedder
	filetypes []*regexp.Regexp
}

// NewIndexer creates a new Indexer with a FileWatcher.
func NewIndexer() (*Indexer, error) {
	w, err := watcher.NewFileWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}
	return &Indexer{watcher: w}, nil
}

// Start indexes watch_paths and listens for file events.
func (i *Indexer) Start(ctx context.Context) error {
	// Start the watcher
	if err := i.watcher.Start(ctx); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	// Index all watch_paths on startup
	if err := i.indexWatchPaths(ctx); err != nil {
		return fmt.Errorf("failed to index watch paths: %w", err)
	}

	// Process file events from watcher
	go i.processEvents(ctx)
	return nil
}

// IndexPath indexes a specific path and adds it to watch_paths.
func (i *Indexer) IndexPath(ctx context.Context, path string) error {
	// Add path to watcher and config
	if err := i.watcher.AddPath(path); err != nil {
		return fmt.Errorf("failed to add path %s: %w", path, err)
	}

	// Index the path
	return i.indexPath(ctx, path)
}

// indexWatchPaths indexes all configured watch_paths.
func (i *Indexer) indexWatchPaths(ctx context.Context) error {
	paths := config.GetWatchPaths()

	for _, path := range paths {
		if err := i.indexPath(ctx, path); err != nil {
			return fmt.Errorf("failed to index path %s: %w", path, err)
		}
	}
	return nil
}

// indexPath indexes a single directory.
func (i *Indexer) indexPath(ctx context.Context, path string) error {
	ignorePatterns := config.GetIgnorePatterns()

	return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			for _, pattern := range ignorePatterns {
				if matched, _ := filepath.Match(pattern, info.Name()); matched {
					return filepath.SkipDir
				}
			}
			return nil
		}
		// TODO: Extract file content, generate LLM embeddings, store in DB
		fmt.Printf("Indexing: %s\n", filePath)
		return nil
	})
}

// processEvents handles file change events from the watcher.
func (i *Indexer) processEvents(ctx context.Context) {
	for {
		select {
		case filePath := <-i.watcher.Events():
			// TODO: Re-index or remove file from DB
			fmt.Printf("Processing change: %s\n", filePath)
		case <-ctx.Done():
			return
		}
	}
}
