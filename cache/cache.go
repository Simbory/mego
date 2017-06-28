package cache

import (
	"errors"
	"fmt"
	"github.com/Simbory/mego/fswatcher"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

var (
	maxDate = time.Date(9999, 12, 31, 23, 59, 59, 999, time.UTC)
)

type dataEntity struct {
	data         interface{}
	dependencies []string
	expire       time.Time
}

// Manager implement the cache manager struct
type Manager struct {
	dataMap     map[string]*dataEntity
	gcInterval  time.Duration
	fileWatcher *fswatcher.FileWatcher
	locker      *sync.RWMutex
	started     bool
	timer       *time.Timer
}

func (c *Manager) fileUsage(fPath string) int {
	count := 0
	for _, data := range c.dataMap {
		if len(data.dependencies) > 0 {
			for _, f := range data.dependencies {
				if strings.EqualFold(fPath, f) {
					count = count + 1
				}
			}
		}
	}
	return count
}

// Add add cache data to memory
func (c *Manager) Set(name string, data interface{}, dependencyFiles []string, expired time.Duration) error {
	if len(name) == 0 {
		return errors.New("The parameter 'name' cannot be empty")
	}
	if data == nil {
		return errors.New("The parameter 'data' cannot be nil")
	}
	var dFiles []string
	if len(dependencyFiles) != 0 {
		for _, file := range dependencyFiles {
			if len(file) == 0 {
				continue
			}
			fPath := path.Clean(strings.Replace(file, "\\", "/", -1))
			if stat, err := os.Stat(fPath); err != nil || stat.IsDir() {
				return fmt.Errorf("Invalid dependency file. The file cannot be found: \"%s\".", file)
			}
			dFiles = append(dFiles, fPath)
		}
	}
	expireTime := ExpireTime(expired)
	c.locker.Lock()
	c.dataMap[name] = &dataEntity{
		data:         data,
		dependencies: dFiles,
		expire:       expireTime,
	}
	c.locker.Unlock()
	if len(dFiles) > 0 {
		for _, f := range dFiles {
			c.fileWatcher.AddWatch(f, false)
		}
	}
	return nil
}

// Get get the data by name
func (c *Manager) Get(name string) interface{} {
	if c.dataMap == nil {
		return nil
	}
	data, ok := c.dataMap[name]
	if ok {
		if time.Now().Before(data.expire) {
			return data.data
		}
		return nil
	}
	return nil
}

// Keys get all the cache keys
func (c *Manager) Keys(name string) []string {
	keys := make([]string, 0, len(c.dataMap))
	for key := range c.dataMap {
		keys = append(keys, key)
	}
	return keys
}

// Remove remove the cache from memory by key
func (c *Manager) Remove(name string) {
	if len(name) == 0 {
		return
	}
	data, ok := c.dataMap[name]
	if !ok {
		return
	}
	if len(data.dependencies) > 0 {
		for _, f := range data.dependencies {
			if c.fileUsage(f) == 1 && c.fileWatcher != nil {
				c.fileWatcher.RemoveWatch(f)
			}
		}
	}
	c.locker.Lock()
	delete(c.dataMap, name)
	c.locker.Unlock()
}

func (c *Manager) gc() {
	var now = time.Now()
	for name, data := range c.dataMap {
		if now.Before(data.expire) {
			c.locker.Lock()
			delete(c.dataMap, name)
			c.locker.Unlock()
		}
	}
	c.timer = time.AfterFunc(c.gcInterval, func() {
		c.gc()
	})
}

// start init the fswatcher and start the gc lifecycle
func (c *Manager) start() {
	if c.started || c.fileWatcher == nil {
		return
	} else {
		c.started = true
		c.fileWatcher.Start()
		c.fileWatcher.AddHandler(&fileHandler{cacheManager: c})
	}
	// start the cache gc lifecycle
	c.gc()
}

func (c *Manager) Stop() {
	c.timer.Stop()
	c.fileWatcher.Stop()
	c.dataMap = nil
}

const DefaultGCInterval = 10 * time.Second

func NewManager(gcInterval time.Duration) *Manager {
	if gcInterval <= 0 {
		panic(fmt.Errorf("Invalid gabage collection time interval: %d", gcInterval))
	}
	fw,err := fswatcher.NewWatcher()
	if err != nil {
		fw = nil
	}
	m := &Manager{
		locker:      &sync.RWMutex{},
		dataMap:     make(map[string]*dataEntity),
		gcInterval:  time.Duration(gcInterval),
		fileWatcher: fw,
	}
	m.start()
	return m
}

func ExpireTime(expired time.Duration) time.Time {
	if expired <= 0 {
		return maxDate
	}
	return time.Now().Add(expired)
}

var defaultManager *Manager

func UseDefault() {
	if defaultManager == nil {
		defaultManager = NewManager(DefaultGCInterval)
	}
}

func Default() *Manager {
	if defaultManager == nil {
		panic(errors.New("You need to call UseDefault() function first before getting the default cache manager"))
	}
	return defaultManager
}