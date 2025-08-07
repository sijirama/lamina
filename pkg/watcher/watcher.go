package watcher

import (
	"context"
	"fmt"
	"lamina/pkg/config"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher monitors directories for changes.
type FileWatcher struct {
	watcher        *fsnotify.Watcher
	events         chan string // Channel for file events to be processed by indexer
	ignorePatterns []string
}

func NewFileWatcher() (*FileWatcher, error) {

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	ignorePatterns := config.GetIgnorePatterns()

	return &FileWatcher{
		watcher:        watcher,
		events:         make(chan string, 100), // Buffered channel for events
		ignorePatterns: ignorePatterns,
	}, nil
}

// Start begins watching the configured watch_paths.
func (fw *FileWatcher) Start(ctx context.Context) error {
	paths := config.GetWatchPaths()

	for _, path := range paths {
		if err := fw.addPath(path); err != nil {
			return fmt.Errorf("failed to watch path %s: %w", path, err)
		}
	}

	go fw.watchEvents(ctx)
	return nil
}

// addPath adds a directory to the watcher recursively.
func (fw *FileWatcher) addPath(path string) error {
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			for _, pattern := range fw.ignorePatterns {
				if matched, _ := filepath.Match(pattern, info.Name()); matched {
					return filepath.SkipDir
				}
			}
			return fw.watcher.Add(p)
		}
		return nil
	})
}

func (fw *FileWatcher) Events() <-chan string {
	return fw.events
}

// watchEvents processes filesystem events.
func (fw *FileWatcher) watchEvents(ctx context.Context) {
	defer fw.watcher.Close()
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if fw.shouldIgnore(event.Name) {
				continue
			}
			if event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Remove == fsnotify.Remove {
				fw.events <- event.Name
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("⚠️ Watcher error: %v\n", err)
		case <-ctx.Done():
			return
		}
	}
}

// shouldIgnore checks if a path matches any ignore patterns.
func (fw *FileWatcher) shouldIgnore(path string) bool {
	base := filepath.Base(path)

	if strings.HasSuffix(base, "~") {
		return true
	}

	for _, pattern := range fw.ignorePatterns {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
	}
	return false
}
