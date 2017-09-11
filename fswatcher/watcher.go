package fswatcher

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"os"
	pathPkg "path"
	"path/filepath"
	"strings"
)

// Handler the fswatcher handler interface
type Handler interface {
	CanHandle(path string) bool
	Handle(ev *fsnotify.Event)
}

// ErrorHandler the fsnotify error handler
type ErrorHandler func(error)

// FileWatcher the file fswatcher struct
type FileWatcher struct {
	watcher        *fsnotify.Watcher
	handlers       []Handler
	errorProcessor ErrorHandler
	started        bool
}

// AddWatch add path to watch
func (fw *FileWatcher) AddWatch(path string, subDir bool) error {
	path = pathPkg.Clean(strings.Replace(path, "\\", "/", -1))
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	err = fw.watcher.Add(path)
	if subDir && stat.IsDir() {
		filepath.Walk(path, func(p string, info os.FileInfo, er error) error {
			if err != nil {
				return nil
			}
			if er != nil {
				err = er
				return er
			}
			if info.IsDir() {
				fw.AddWatch(p, false)
			}
			return nil
		})
	}
	return err
}

// RemoveWatch remove the path from fswatcher. And then, the system will stop watching the changes.
func (fw *FileWatcher) RemoveWatch(path string) error {
	return fw.watcher.Remove(path)
}

// AddHandler add file fswatcher handler
func (fw *FileWatcher) AddHandler(handler Handler) error {
	if handler == nil {
		return errors.New("the parameter 'handler' cannot be nil")
	}
	fw.handlers = append(fw.handlers, handler)
	return nil
}

// SetErrorHandler set the fsnotify error handler
func (fw *FileWatcher) SetErrorHandler(h ErrorHandler) {
	fw.errorProcessor = h
}

// Start star the file fswatcher
func (fw *FileWatcher) Start() {
	if fw.started {
		return
	}
	fw.started = true
	go func() {
		for {
			select {
			case ev := <-fw.watcher.Events:
				ev.Name = strings.Replace(pathPkg.Clean(ev.Name), "\\", "/", -1)
				for _, detector := range fw.handlers {
					if detector.CanHandle(ev.Name) {
						detector.Handle(&ev)
					}
				}
			case err := <-fw.watcher.Errors:
				if fw.errorProcessor != nil {
					fw.errorProcessor(err)
				}
			}
			if !fw.started {
				break
			}
		}
	}()
}

func (fw *FileWatcher) Stop() {
	fw.watcher.Close()
	fw.started = false
}

// NewWatcher create the new fswatcher
func NewWatcher() (*FileWatcher, error) {
	t, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &FileWatcher{
		watcher: t,
	}
	return w, nil
}
