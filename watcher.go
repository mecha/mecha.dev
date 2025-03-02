package main

import (
	"log/slog"

	"github.com/fsnotify/fsnotify"
)

type DirWatcher struct {
	dirpath   string
	callback  EventCallback
	stopChan  chan bool
	recursive bool
}

type EventCallback func(event fsnotify.Event)

// Creates a new watcher for a directory, that calls the given callback with
// file system events.
func NewDirWatcher(path string, cb EventCallback) *DirWatcher {
	return &DirWatcher{path, cb, nil, false}
}

// Starts listening for file system events.
func (dw *DirWatcher) Start() error {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	dw.stopChan = make(chan bool)

	go func() {
	loop:
		for {
			select {
			case event, ok := <-fsw.Events:
				if !ok {
					break loop
				}
				dw.callback(event)

			case err, ok := <-fsw.Errors:
				if !ok {
					break loop
				}
				slog.Error(err.Error())

			case _ = <-dw.stopChan:
				break loop
			}
		}

		fsw.Close()
		close(dw.stopChan)
		dw.stopChan = nil
	}()

	err = fsw.Add(dw.dirpath)
	if err != nil {
		return err
	}

	return nil
}

// Stops listening for file system events.
func (dw *DirWatcher) Stop() {
	if dw.stopChan != nil {
		dw.stopChan <- true
	}
}

// Returns whether the watcher is currently listening for file system events.
func (dw *DirWatcher) IsActive() bool {
	return dw.stopChan != nil
}
