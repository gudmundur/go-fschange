// Package fschange notifies of file changes since provided time
package fschange

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/fsnotify.v1"
)

// Watcher watches a set of directories, delivering events to a channel.
type Watcher struct {
	Since time.Time
	until time.Time

	Events chan Event
	Errors chan error

	watcher *fsnotify.Watcher
}

// Event represents a single file system notification.
type Event struct {
	Name string
	Op   Op
}

// Op describes a set of file operations.
type Op uint32

// These are the generalized file operations that can trigger a notification.
const (
	Create Op = 1 << iota
	Write
	Remove
	Rename
	Chmod
)

// NewWatcher establishes a new file system watcher and notifies of any changes after since.
func NewWatcher(since time.Time) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		Since:   since,
		until:   time.Now(),
		Events:  make(chan Event),
		Errors:  make(chan error),
		watcher: watcher,
	}

	go w.watch()

	return w, nil
}

// Add starts watching the directory recursively.
func (w *Watcher) Add(path string) error {
	w.walk(path)
	return nil
}

// Closes all watches and channels
func (w *Watcher) Close() error {
	w.watcher.Close()
	close(w.Events)
	close(w.Errors)
	return nil
}

func (w *Watcher) walk(path string) {
	filepath.Walk(path, w.walkFunc)
}

func (w *Watcher) walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		w.Errors <- err
		return err
	}

	if info.IsDir() {
		err = w.watcher.Add(path)
		if err != nil {
			w.Errors <- err
			return err
		}
	}

	modTime := info.ModTime()
	if modTime.After(w.Since) && modTime.Before(w.until) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			w.Errors <- err
			return err
		}
		w.Events <- Event{absPath, Create}
	}

	return nil
}

func (w *Watcher) watch() {
	for {
		select {
		case event := <-w.watcher.Events:
			w.Events <- Event{
				Name: event.Name,
				Op:   Op(event.Op),
			}
		case err := <-w.watcher.Errors:
			w.Errors <- err
		}
	}
}
