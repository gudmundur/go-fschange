// package fschange
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Watcher struct {
	Since time.Time

	Events chan Event
	Errors chan error
}

type Event struct {
	Name string
}

func NewWatcher(since time.Time) (*Watcher, error) {
	w := &Watcher{
		Since:  since,
		Events: make(chan Event),
		Errors: make(chan error),
	}
	return w, nil
}

func (w *Watcher) Add(path string) error {
	w.walk(path)
	return nil
}

func (w *Watcher) Close() error {
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

	if w.Since.Before(info.ModTime()) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			w.Errors <- err
			return err
		}
		w.Events <- Event{absPath}
	}

	return nil
}

func main() {
	sinceTime := time.Now().Add(-time.Hour * 24)
	fmt.Println(sinceTime)
	w, _ := NewWatcher(sinceTime)

	go func() {
		for event := range w.Events {
			fmt.Println(event)
		}
	}()

	w.Add("../")
	w.Close()
}
