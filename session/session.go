package session

import (
	"container/list"
	"errors"
	"github.com/Simbory/mego"
	"net/http"
	"net/url"
	"time"
	"github.com/google/uuid"
)

var (
	providers = make(map[string]SessionProvider)
)

var manager *SessionManager
var config *SessionConfig

func init() {
	mego.OnStart(func() {
		if config == nil {
			return
		}
		if len(config.ManagerName) == 0 {
			config.ManagerName = "memory"
		}
		if len(config.CookieName) == 0 {
			config.CookieName = "MEGO_SESSIONID"
		}
		if config.GcLifetime == 0 {
			config.GcLifetime = 3600
		}
		if config.MaxLifetime == 0 {
			config.MaxLifetime = 3600
		}
		if config.ManagerName == "memory" {
			providers["memory"] = &memSessionProvider{list: list.New(), sessions: make(map[string]*list.Element)}
		}
		m, err := newSessionManager(config.ManagerName, config)
		if err != nil {
			panic(err)
		}
		manager = m
		go manager.GC()
	})
}

func RegSessionProvider(name string, provider SessionProvider) {
	mego.AssertUnlocked()
	if provider == nil {
		panic(errors.New("The parameter 'provider' canot be nil"))
	}
	if _, dup := providers[name]; dup {
		panic(errors.New("session: Register called twice for provider " + name))
	}
	providers[name] = provider
}

func UseSession(sessionConfig *SessionConfig) {
	mego.AssertUnlocked()
	if sessionConfig == nil {
		sessionConfig = &SessionConfig{
			ManagerName:     "memory",
			CookieName:      "MEGO_SESSIONID",
			EnableSetCookie: true,
			GcLifetime:      3600,
			MaxLifetime:     3600,
			HTTPOnly:        true,
		}
	}
	config = sessionConfig
}

func newSessionId() string {
	id,_ := uuid.NewUUID()
	return id.String()
}

// Start generate or read the session id from http request.
// if session id exists, return Store with this id.
func Start(ctx *mego.Context) Store {
	r := ctx.Request()
	w := ctx.Response()
	id, err := manager.getSessionID(r)
	if err != nil {
		return nil
	}
	if id != "" && manager.provider.SessionExist(id) {
		return manager.provider.SessionRead(id)
	}
	// Generate a new store
	id = newSessionId()
	store := manager.provider.SessionRead(id)
	cookie := &http.Cookie{
		Name:     manager.config.CookieName,
		Value:    url.QueryEscape(id),
		Path:     "/",
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
func Destroy(ctx *mego.Context) {
	r := ctx.Request()
	w := ctx.Response()
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		return
	}
	sid, _ := url.QueryUnescape(cookie.Value)
	manager.provider.SessionDestroy(sid)
	if manager.config.EnableSetCookie {
		expiration := time.Now()
		cookie = &http.Cookie{
			Name:     manager.config.CookieName,
			Path:     "/",
			HttpOnly: true,
			Expires:  expiration,
			MaxAge:   -1,
		}
		http.SetCookie(w, cookie)
	}
}

// RegenerateID Regenerate a session id for this Store who's id is saving in http request.
func RegenerateID(ctx *mego.Context) (session Store) {
	r := ctx.Request()
	w := ctx.Response()
	sid := newSessionId()
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		//delete old cookie
		session = manager.provider.SessionRead(sid)
		cookie = &http.Cookie{Name: manager.config.CookieName,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: manager.config.HTTPOnly,
			Secure:   manager.isSecure(r),
			Domain:   manager.config.Domain,
		}
	} else {
		oldSessionId, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.provider.SessionRegenerate(oldSessionId, sid)
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
