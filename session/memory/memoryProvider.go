package memory

import (
	"sync"
	"container/list"
	"time"
	"github.com/simbory/mego/session"
)

// provider Implement the provider interface
type provider struct {
	lock        sync.RWMutex             // locker
	sessions    map[string]*list.Element // map in memory
	list        *list.List               // for gc
	maxLifetime int64
}

// Init init memory session
func (prov *provider) Init(maxLifeTime int64, savePath string) error {
	prov.maxLifetime = maxLifeTime
	return nil
}

// Read get memory session store by sid
func (prov *provider) Read(sid string) session.Storage {
	prov.lock.RLock()
	if element, ok := prov.sessions[sid]; ok {
		go prov.Update(sid)
		prov.lock.RUnlock()
		return element.Value.(*storage)
	}
	prov.lock.RUnlock()
	prov.lock.Lock()
	newStore := &storage{sid: sid, timeAccessed: time.Now(), value: make(map[string]interface{})}
	element := prov.list.PushFront(newStore)
	prov.sessions[sid] = element
	prov.lock.Unlock()
	return newStore
}

// Exist check session store exist in memory session by sid
func (prov *provider) Exist(sid string) bool {
	prov.lock.RLock()
	defer prov.lock.RUnlock()
	if _, ok := prov.sessions[sid]; ok {
		return true
	}
	return false
}

// Regenerate generate new sid for session store in memory session
func (prov *provider) Regenerate(oldSid, sid string) (session.Storage, error) {
	prov.lock.RLock()
	if element, ok := prov.sessions[oldSid]; ok {
		go prov.Update(oldSid)
		prov.lock.RUnlock()
		prov.lock.Lock()
		element.Value.(*storage).sid = sid
		prov.sessions[sid] = element
		delete(prov.sessions, oldSid)
		prov.lock.Unlock()
		return element.Value.(*storage), nil
	}
	prov.lock.RUnlock()
	prov.lock.Lock()
	newStore := &storage{sid: sid, timeAccessed: time.Now(), value: make(map[string]interface{})}
	element := prov.list.PushFront(newStore)
	prov.sessions[sid] = element
	prov.lock.Unlock()
	return newStore, nil
}

// Destroy delete session store in memory session by id
func (prov *provider) Destroy(sid string) error {
	prov.lock.Lock()
	defer prov.lock.Unlock()
	if element, ok := prov.sessions[sid]; ok {
		delete(prov.sessions, sid)
		prov.list.Remove(element)
		return nil
	}
	return nil
}

// All get count number of memory session
func (prov *provider) All() int {
	return prov.list.Len()
}

// Update expand Time of session store by id in memory session
func (prov *provider) Update(sid string) error {
	prov.lock.Lock()
	defer prov.lock.Unlock()
	if element, ok := prov.sessions[sid]; ok {
		element.Value.(*storage).timeAccessed = time.Now()
		prov.list.MoveToFront(element)
		return nil
	}
	return nil
}

// GC clean expired session stores in memory session
func (prov *provider) GC() {
	prov.lock.RLock()
	for {
		element := prov.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*storage).timeAccessed.Unix() + prov.maxLifetime) < time.Now().Unix() {
			prov.lock.RUnlock()
			prov.lock.Lock()
			prov.list.Remove(element)
			delete(prov.sessions, element.Value.(*storage).sid)
			prov.lock.Unlock()
			prov.lock.RLock()
		} else {
			break
		}
	}
	prov.lock.RUnlock()
}

func NewProvider() session.Provider {
	return &provider{
		list: list.New(),
		sessions: make(map[string]*list.Element),
	}
}