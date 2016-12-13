package view

import (
	"github.com/Simbory/mego/watcher"
	"github.com/fsnotify/fsnotify"
	"os"
	"path"
	"strings"
)

type fsViewHandler struct {
	fsWatcher *watcher.FileWatcher
}

func (vh *fsViewHandler) CanHandle(path string) bool {
	return strings.HasPrefix(path, viewDir())
}

func (vh *fsViewHandler) Handle(ev *fsnotify.Event) {
	strFile := strings.ToLower(path.Clean(ev.Name))
	if ev.Op&fsnotify.Remove == fsnotify.Remove {
		if !strings.HasSuffix(strFile, viewSingleton.viewExt) {
			vh.fsWatcher.RemoveWatch(ev.Name)
		}
	} else {
		if state, err := os.Stat(ev.Name); err == nil && state.IsDir() {
			vh.fsWatcher.AddWatch(ev.Name)
		}
	}
	viewSingleton.compileViews()
}
