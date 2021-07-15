package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

func watch(dir, file string, callback func()) (done chan struct{}, err error) {
	done = make(chan struct{})
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to watch addon for changes: %v", err)
	}

	debouncedCallback := debounce(callback, 1*time.Second)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Name == file || event.Name == "./"+file { // file name is prepended with ./ on linux
					debouncedCallback()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				if err != nil {
					log.Printf("error watching addon for changes: %v", err)
				}
			case <-done:
				return
			}
		}
	}()
	return done, nil
}

func debounce(f func(), d time.Duration) func() {
	lock := sync.Mutex{}
	var cancel chan struct{}
	return func() {
		lock.Lock()
		defer lock.Unlock()

		if cancel != nil {
			close(cancel)
		}
		cancel = make(chan struct{})
		go func() {
			select {
			case <-time.After(d):
				f()
			case <-cancel:
			}
		}()
	}
}
