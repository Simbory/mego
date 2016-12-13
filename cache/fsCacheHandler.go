package cache

import (
	"github.com/fsnotify/fsnotify"
	"strings"
)

type fsCacheHandler struct {
	cacheManager *CacheManager
	cacheKey     string
}

// CanHandle detect the fsnotify path can handled by current detector
func (cd *fsCacheHandler) CanHandle(path string) bool {
	for name, data := range cd.cacheManager.dataMap {
		if len(data.dependencies) == 0 {
			continue
		}
		for _, file := range data.dependencies {
			if strings.EqualFold(file, path) {
				cd.cacheKey = name
				return true
			}
		}
	}
	return false
}

// Handle handle the fsnotify changes
func (cd *fsCacheHandler) Handle(ev *fsnotify.Event) {
	cd.cacheManager.locker.Lock()
	delete(cd.cacheManager.dataMap, cd.cacheKey)
	cd.cacheManager.locker.Lock()
}

func newCacheDetector(manager *CacheManager) *fsCacheHandler {
	return &fsCacheHandler{cacheManager: manager}
}
