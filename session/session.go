package session

import (
	"container/list"
	"errors"
	"github.com/Simbory/mego"
	"net/http"
	"net/url"
	"time"
)

var (
	sessionProvides = make(map[string]SessionProvider)
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
			config.CookieName = "mego.SessionID"
		}
		if config.GcLifetime == 0 {
			config.GcLifetime = 3600
		}
		if config.MaxLifetime == 0 {
			config.MaxLifetime = 3600
		}
		if config.ManagerName == "memory" {
			sessionProvides["memory"] = &memSessionProvider{list: list.New(), sessions: make(map[string]*list.Element)}
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
	mego.AssertNotLock()
	if provider == nil {
		panic(errors.New("The parameter 'provider' canot be nil"))
	}
	if _, dup := sessionProvides[name]; dup {
		panic(errors.New("session: Register called twice for provider " + name))
	}
	sessionProvides[name] = provider
}

func UseSession(sessionConfig *SessionConfig) {
	mego.AssertNotLock()
	if sessionConfig == nil {
		sessionConfig = &SessionConfig{
			ManagerName:     "memory",
			CookieName:      "mego.SessionID",
			EnableSetCookie: true,
			GcLifetime:      3600,
			MaxLifetime:     3600,
			HTTPOnly:        true,
		}
	}
	config = sessionConfig
}

// SessionStart generate or read the session id from http request.
// if session id exists, return SessionStore with this id.
func SessionStart(ctx *mego.Context) SessionStore {
	r := ctx.Request()
	w := ctx.Response()
	sessionID, err := manager.getSessionID(r)
	if err != nil {
		return nil
	}
	if sessionID != "" && manager.provider.SessionExist(sessionID) {
		return manager.provider.SessionRead(sessionID)
	}
	// Generate a new session
	sessionID = newSessionId().string()
	session := manager.provider.SessionRead(sessionID)
	cookie := &http.Cookie{
		Name:     manager.config.CookieName,
		Value:    url.QueryEscape(sessionID),
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
	return session
}

// SessionDestroy Destroy session by its id in http request cookie.
func SessionDestroy(ctx *mego.Context) {
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

// SessionRegenerateID Regenerate a session id for this SessionStore who's id is saving in http request.
func SessionRegenerateID(ctx *mego.Context) (session SessionStore) {
	r := ctx.Request()
	w := ctx.Response()
	sid := newSessionId().string()
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
