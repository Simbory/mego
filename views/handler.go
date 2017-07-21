package views

import (
	"github.com/fsnotify/fsnotify"
	"os"
	"path"
	"strings"
)

type compileHandler struct {
	engine *ViewEngine
}

func (vh *compileHandler) CanHandle(path string) bool {
	return strings.HasPrefix(path, vh.engine.viewDir) && strings.HasSuffix(strings.ToLower(path), strings.ToLower(vh.engine.viewExt))
}

func (vh *compileHandler) Handle(ev *fsnotify.Event) {
	strFile := strings.ToLower(path.Clean(ev.Name))
	if ev.Op&fsnotify.Remove == fsnotify.Remove {
		if !strings.HasSuffix(strFile, vh.engine.viewExt) {
			vh.engine.watcher.RemoveWatch(ev.Name)
		}
	} else {
		if state, err := os.Stat(ev.Name); err == nil && state.IsDir() {
			vh.engine.watcher.AddWatch(ev.Name, true)
		}
	}
	vh.engine.Clear()
}
