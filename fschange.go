// package fschange
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/fsnotify.v1"
)

type Watcher struct {
	Since time.Time
	until time.Time

	Events chan Event
	Errors chan error

	watcher *fsnotify.Watcher
}

type Event struct {
	Name string
}

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

func (w *Watcher) Add(path string) error {
	w.walk(path)
	return nil
}

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
		w.Events <- Event{absPath}
	}

	return nil
}

func (w *Watcher) watch() {
	for {
		select {
		case event := <-w.watcher.Events:
			w.Events <- Event{event.Name}
		case err := <-w.watcher.Errors:
			w.Errors <- err
		}
	}
}

func main() {
	sinceTime := time.Now().Add(-time.Hour * 24)
	w, _ := NewWatcher(sinceTime)

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-w.Events:
				fmt.Println(event)
			case err := <-w.Errors:
				fmt.Println(err)
			}
		}
	}()

	w.Add("../")
	<-done
	w.Close()
}
