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

	fmt.Println("Watcher started")

	// Index all watch_paths on startup
	if err := i.indexWatchPaths(ctx); err != nil {
		return fmt.Errorf("failed to index watch paths: %w", err)
	}

	// Process file events from watcher
	go i.processEvents(ctx)
	return nil
}

// indexWatchPaths indexes all configured watch_paths.
func (i *Indexer) indexWatchPaths(ctx context.Context) error {
	paths := config.GetWatchPaths()

	fmt.Println("Indexing paths", paths)

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

		fmt.Println("Index path", path)
		err = i.indexFile(ctx, path)
		fmt.Println("Error indexing file: ", err.Error())
		return nil
	})
}

// processEvents handles file change events from the watcher.
func (i *Indexer) processEvents(ctx context.Context) {
	for {
		select {
		case filePath := <-i.watcher.Events():
			err := i.indexFile(ctx, filePath)
			fmt.Println("Error indexing file: ", err.Error())
		case <-ctx.Done():
			return
		}
	}
}
