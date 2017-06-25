package cache

import (
	"github.com/fsnotify/fsnotify"
	"strings"
)

type fileHandler struct {
	cacheManager *Manager
	cacheKeys    []string
}

// CanHandle detect the fsnotify path can handled by current detector
func (cd *fileHandler) CanHandle(path string) bool {
	cd.cacheKeys = nil
	found := false
	for name, data := range cd.cacheManager.dataMap {
		if len(data.dependencies) == 0 {
			continue
		}
		for _, file := range data.dependencies {
			if strings.EqualFold(file, path) {
				cd.cacheKeys = append(cd.cacheKeys, name)
				found = true
			}
		}
	}
	return found
}

// Handle handle the fsnotify changes
func (cd *fileHandler) Handle(ev *fsnotify.Event) {
	cd.cacheManager.locker.Lock()
	for _, key := range cd.cacheKeys {
		cd.cacheManager.dataMap[key] = nil
		delete(cd.cacheManager.dataMap, key)
	}
	cd.cacheManager.locker.Lock()
}