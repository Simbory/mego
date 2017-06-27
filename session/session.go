package session

import (
	"container/list"
	"errors"
	"encoding/gob"
	"github.com/Simbory/mego"
)

var defaultManager *Manager

func UseDefault(server *mego.Server) {
	defaultManager = CreateManager(server,nil, nil)
}

// UseAsDefault use the given session manager as the default
func UseAsDefault(manager *Manager) {
	if manager != nil {
		defaultManager = manager
	}
}

func CreateManager(server *mego.Server, config *Config, provider Provider) *Manager {
	if config == nil {
		config = new(Config)
		config.ManagerName = "memory"
		config.CookieName = "MEGO_SESSIONID"
		config.GcLifetime = 3600
		config.MaxLifetime = 3600
	}
	if provider == nil {
		provider = &memoryProvider{
			list: list.New(),
			sessions: make(map[string]*list.Element),
		}
	}
	m, err := newSessionManager(server, config, provider)
	if err != nil {
		panic(err)
	}
	go m.GC()
	return m
}

func Default() *Manager {
	if defaultManager == nil {
		panic(errors.New("You need to call UseDefault() first when you get the default session manager"))
	}
	return defaultManager
}

func NewMemoryProvider() Provider {
	return &memoryProvider{
		list: list.New(),
		sessions: make(map[string]*list.Element),
	}
}

func NewDiskProvider(dir string) Provider {
	gob.Register(&diskValue{})
	return &diskProvider{savePath: dir}
}