package session

import (
	"errors"
	"github.com/simbory/mego"
	"github.com/simbory/mego/assert"
)

var defaultManager *Manager

// UseAsDefault use the given session manager as the default
func UseAsDefault(manager *Manager) {
	assert.NotNil("manager", manager)
	defaultManager = manager
}

func CreateManager(server *mego.Server, config *Config, provider Provider) *Manager {
	assert.NotNil("server", server)
	assert.NotNil("provider", provider)
	if config == nil {
		config = new(Config)
		config.CookieName = "SESSION_ID"
		config.GcLifetime = 3600
		config.MaxLifetime = 3600
		config.HTTPOnly = true
	}
	m := newSessionManager(server, config, provider)
	go m.GC()
	return m
}

func Default() *Manager {
	if defaultManager == nil {
		panic(errors.New("You need to call UseDefault() first when you get the default session manager"))
	}
	return defaultManager
}