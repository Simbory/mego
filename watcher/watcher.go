package watcher

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"path"
	"sync"
)

// Handler the watcher handler interface
type Handler interface {
	CanHandle(path string) bool
	Handle(ev *fsnotify.Event)
}

// ErrorHandler the fsnotify error handler
type ErrorHandler func(error)

// FileWatcher the file watcher struct
type FileWatcher struct {
	watcher        *fsnotify.Watcher
	handlers       []Handler
	errorProcessor ErrorHandler
	started        bool
}

// AddWatch add path to watch
func (fw *FileWatcher) AddWatch(path string) error {
	return fw.watcher.Add(path)
}

// RemoveWatch remove path from watcher
func (fw *FileWatcher) RemoveWatch(strFile string) error {
	return fw.watcher.Remove(strFile)
}

// AddHandler add file watcher handler
func (fw *FileWatcher) AddHandler(handler Handler) error {
	if handler == nil {
		return errors.New("The parameter 'handler' cannot be nil")
	}
	fw.handlers = append(fw.handlers, handler)
	return nil
}

// SetErrorHandler set the fsnotify error handler
func (fw *FileWatcher) SetErrorHandler(h ErrorHandler) {
	fw.errorProcessor = h
}

// Start star the file watcher
func (fw *FileWatcher) Start() {
	if fw.started {
		return
	}
	fw.started = true
	go func() {
		for {
			select {
			case ev := <-fw.watcher.Events:
				for _, detector := range fw.handlers {
					if detector.CanHandle(path.Clean(ev.Name)) {
						detector.Handle(&ev)
					}
				}
			case err := <-fw.watcher.Errors:
				if fw.errorProcessor != nil {
					fw.errorProcessor(err)
				}
			}
		}
	}()
}

// NewWatcher create the new watcher
func NewWatcher() (*FileWatcher, error) {
	tmpWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &FileWatcher{
		watcher: tmpWatcher,
	}
	return w, nil
}

var singleton *FileWatcher
var singletonLocker = sync.RWMutex{}

func Singleton() *FileWatcher {
	if singleton == nil {
		singletonLocker.Lock()
		if singleton == nil {
			s, err := NewWatcher()
			if err != nil {
				panic(err)
			}
			singleton = s
		}
		singletonLocker.Unlock()
	}
	return singleton
}
