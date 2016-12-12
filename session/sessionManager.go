package session

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
	"github.com/Simbory/mego"
	"container/list"
	"errors"
)

type SessionConfig struct {
	ManagerName     string `xml:"manager,attr"`
	CookieName      string `xml:"cookieName,attr"`
	EnableSetCookie bool   `xml:"enableSetCookie,attr"`
	GcLifetime      int64  `xml:"gclifetime,attr"`
	MaxLifetime     int64  `xml:"maxLifetime,attr"`
	Secure          bool   `xml:"secure,attr"`
	HTTPOnly        bool   `xml:"httpOnly,attr"`
	CookieLifeTime  int    `xml:"cookieLifeTime,attr"`
	ProviderConfig  string `xml:"providerConfig,attr"`
	Domain          string `xml:"domain,attr"`
}

// SessionManager the session manager struct
type SessionManager struct {
	provider SessionProvider
	config   *SessionConfig
}

var (
	sessionProvides = make(map[string]SessionProvider)
)

var manager *SessionManager

func init() {
	RegSessionProvider("memory", &memSessionProvider{list: list.New(), sessions: make(map[string]*list.Element)})
}

func RegSessionProvider(name string, provider SessionProvider) {
	mego.AssertLock()
	if provider == nil {
		panic(errors.New("The parameter 'provider' canot be nil"))
	}
	if _, dup := sessionProvides[name]; dup {
		panic(errors.New("session: Register called twice for provider " + name))
	}
	sessionProvides[name] = provider
}

func UseSession(sessionConfig *SessionConfig) {
	mego.AssertLock()
	m,err := newSessionManager(sessionConfig.ManagerName, sessionConfig)
	if err != nil {
		panic(err)
	}
	manager = m
	go manager.GC()
}

func newSessionManager(provideName string, config *SessionConfig) (*SessionManager, error) {
	provider, ok := sessionProvides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}
	config.EnableSetCookie = true
	if config.MaxLifetime == 0 {
		config.MaxLifetime = config.GcLifetime
	}
	err := provider.SessionInit(config.MaxLifetime, config.ProviderConfig)
	if err != nil {
		return nil, err
	}

	return &SessionManager{
		provider: provider,
		config:   config,
	}, nil
}

func (manager *SessionManager) getSessionID(r *http.Request) (string, error) {
	cookie, errs := r.Cookie(manager.config.CookieName)
	if errs != nil || cookie.Value == "" || cookie.MaxAge < 0 {
		return "", nil
	}
	// HTTP Request contains cookie for sessionid info.
	return url.QueryUnescape(cookie.Value)
}

// Set cookie with https.
func (manager *SessionManager) isSecure(req *http.Request) bool {
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

func (manager *SessionManager) sessionID() string {
	uuid := newSessionId()
	return uuid.string(), nil
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
	sessionID = manager.sessionID()
	session, err := manager.provider.SessionRead(sessionID)
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

// GetSessionStore Get SessionStore by its id.
func (manager *SessionManager) GetSessionStore(sid string) (sessions SessionStore, err error) {
	sessions, err = manager.provider.SessionRead(sid)
	return
}

// GC Start session gc process.
// it can do gc in times after gc lifetime.
func (manager *SessionManager) GC() {
	manager.provider.SessionGC()
	time.AfterFunc(time.Duration(manager.config.GcLifetime)*time.Second, func() { manager.GC() })
}

// SessionRegenerateID Regenerate a session id for this SessionStore who's id is saving in http request.
func SessionRegenerateID(ctx *mego.Context) (session SessionStore) {
	r := ctx.Request()
	w := ctx.Response()
	sid := manager.sessionID()
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		//delete old cookie
		session, _ = manager.provider.SessionRead(sid)
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

// GetActiveSession Get all active sessions count number.
func GetActiveSession() int {
	return manager.provider.SessionAll()
}