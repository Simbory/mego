package session

import (
	"net/http"
	"net/url"
	"time"
	"github.com/simbory/mego"
	"github.com/google/uuid"
	"encoding/gob"
	"github.com/simbory/mego/assert"
	"sync"
)

type Config struct {
	CookieName      string `xml:"cookieName,attr"`
	CookiePath      string `xml:"cookiePath,attr"`
	EnableSetCookie bool   `xml:"enableSetCookie,attr"`
	GcLifetime      int64  `xml:"gcLifetime,attr"`
	MaxLifetime     int64  `xml:"maxLifetime,attr"`
	Secure          bool   `xml:"secure,attr"`
	HTTPOnly        bool   `xml:"httpOnly,attr"`
	CookieLifeTime  int    `xml:"cookieLifeTime,attr"`
	ProviderConfig  string `xml:"providerConfig,attr"`
	Domain          string `xml:"domain,attr"`
}

// Manager the session manager struct
type Manager struct {
	provider    Provider
	config      *Config
	initialized bool
	lock        sync.RWMutex
}

func (manager *Manager) initialize() {
	// Double-Check Locking
	if manager.initialized {
		return
	}
	manager.lock.Lock()
	defer manager.lock.Unlock()
	if manager.initialized {
		return
	}
	assert.PanicErr(manager.provider.Init(manager.config.GcLifetime, manager.config.ProviderConfig))
	go manager.gc()
	manager.initialized = true
}

func (manager *Manager) getSessionID(r *http.Request) (string, error) {
	cookie, errs := r.Cookie(manager.config.CookieName)
	if errs != nil || cookie.Value == "" || cookie.MaxAge < 0 {
		return "", nil
	}
	// HTTP Request contains cookie for sessionid info.
	return url.QueryUnescape(cookie.Value)
}

// Set cookie with https.
func (manager *Manager) isSecure(req *http.Request) bool {
	if !manager.config.Secure {
		return false
	}
	if req.URL.Scheme != "" {
		return req.URL.Scheme == "https"
	}
	if req.TLS == nil {
		return false
	}
	return true
}

// GetSessionStore Get Storage by its id.
func (manager *Manager) GetSessionStore(sid string) (sessions Storage) {
	sessions = manager.provider.Read(sid)
	return
}

// gc Start session gc process.
// it can do gc in times after gc lifetime.
func (manager *Manager) gc() {
	manager.provider.GC()
	time.AfterFunc(time.Duration(manager.config.GcLifetime)*time.Second, func() {
		manager.gc()
	})
}

// Start generate or read the session id from http request.
// if session id exists, return Storage with this id.
func (manager *Manager) Start(ctx *mego.HttpCtx) Storage {
	manager.initialize()
	r := ctx.Request()
	w := ctx.Response()
	id, err := manager.getSessionID(r)
	if err != nil {
		return nil
	}
	if id != "" && manager.provider.Exist(id) {
		return manager.provider.Read(id)
	}
	// Generate a new store
	id = newSessionId()
	store := manager.provider.Read(id)
	cookie := &http.Cookie{
		Name:     manager.config.CookieName,
		Value:    url.QueryEscape(id),
		Path:     manager.config.CookiePath,
		HttpOnly: manager.config.HTTPOnly,
		Secure:   manager.isSecure(r),
	}
	if len(manager.config.Domain) > 0 {
		cookie.Domain = manager.config.Domain
	}
	if manager.config.CookieLifeTime > 0 {
		cookie.MaxAge = manager.config.CookieLifeTime
		cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifeTime) * time.Second)
	}
	if manager.config.EnableSetCookie {
		http.SetCookie(w, cookie)
	}
	r.AddCookie(cookie)
	return store
}

// Destroy Destroy session by its id in http request cookie.
func (manager *Manager) Destroy(ctx *mego.HttpCtx) {
	manager.initialize()
	r := ctx.Request()
	w := ctx.Response()
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		return
	}
	sid, _ := url.QueryUnescape(cookie.Value)
	manager.provider.Destroy(sid)
	if manager.config.EnableSetCookie {
		expiration := time.Now().Add(-10000)
		cookie = &http.Cookie{
			Name:     manager.config.CookieName,
			Path:     manager.config.CookiePath,
			HttpOnly: manager.config.HTTPOnly,
			Expires:  expiration,
			MaxAge:   -1,
		}
		http.SetCookie(w, cookie)
	}
}

// RegenerateID Regenerate a session id for this Storage who's id is saving in http request.
func (manager *Manager) RegenerateID(ctx *mego.HttpCtx) (session Storage) {
	manager.initialize()
	r := ctx.Request()
	w := ctx.Response()
	sid := newSessionId()
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		//delete old cookie
		session = manager.provider.Read(sid)
		cookie = &http.Cookie{
			Name: manager.config.CookieName,
			Value:                  url.QueryEscape(sid),
			Path:                   manager.config.CookiePath,
			HttpOnly:               manager.config.HTTPOnly,
			Secure:                 manager.isSecure(r),
			Domain:                 manager.config.Domain,
		}
	} else {
		oldSessionId, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.provider.Regenerate(oldSessionId, sid)
		cookie.Value = url.QueryEscape(sid)
		cookie.HttpOnly = true
		cookie.Path = "/"
	}
	if manager.config.CookieLifeTime > 0 {
		cookie.MaxAge = manager.config.CookieLifeTime
		cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifeTime) * time.Second)
	}
	if manager.config.EnableSetCookie {
		http.SetCookie(w, cookie)
	}
	r.AddCookie(cookie)
	return
}

func RegisterType(value interface{}) {
	gob.Register(value)
}

func RegisterTypeName(name string, value interface{}) {
	gob.RegisterName(name, value)
}

func newSessionId() string {
	id,err := uuid.NewUUID()
	assert.PanicErr(err)
	return id.String()
}
