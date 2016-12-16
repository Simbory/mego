package session

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
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

// GetSessionStore Get Store by its id.
func (manager *SessionManager) GetSessionStore(sid string) (sessions Store) {
	sessions = manager.provider.SessionRead(sid)
	return
}

// GC Start session gc process.
// it can do gc in times after gc lifetime.
func (manager *SessionManager) GC() {
	manager.provider.SessionGC()
	time.AfterFunc(time.Duration(manager.config.GcLifetime)*time.Second, func() { manager.GC() })
}

func newSessionManager(provideName string, config *SessionConfig) (*SessionManager, error) {
	provider, ok := providers[provideName]
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
