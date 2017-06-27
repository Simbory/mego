package session

import (
	"sync"
	"container/list"
	"time"
	"os"
	"fmt"
)

// memoryProvider Implement the provider interface
type diskProvider struct {
	lock        sync.RWMutex             // locker
	sessions    map[string]*list.Element // map in memory
	list        *list.List               // for gc
	maxLifetime int64
	savePath    string
}

// Init init memory session
func (prov *diskProvider) Init(maxLifeTime int64, savePath string) error {
	prov.maxLifetime = maxLifeTime
	stat,err := os.Stat(savePath)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		fmt.Errorf("Invalid session provider folder: %s", savePath)
	}
	return nil
}

// Read get memory session store by sid
func (prov *diskProvider) Read(sid string) Storage {
	prov.lock.RLock()
	if element, ok := prov.sessions[sid]; ok {
		go prov.Update(sid)
		prov.lock.RUnlock()
		return element.Value.(*memoryStorage)
	}
	prov.lock.RUnlock()
	prov.lock.Lock()
	newStore := &memoryStorage{sid: sid, timeAccessed: time.Now(), value: make(map[interface{}]interface{})}
	element := prov.list.PushFront(newStore)
	prov.sessions[sid] = element
	prov.lock.Unlock()
	return newStore
}

// Exist check session store exist in memory session by sid
func (prov *diskProvider) Exist(sid string) bool {
	prov.lock.RLock()
	defer prov.lock.RUnlock()
	if _, ok := prov.sessions[sid]; ok {
		return true
	}
	return false
}

// Regenerate generate new sid for session store in memory session
func (prov *diskProvider) Regenerate(oldSid, sid string) (Storage, error) {
	prov.lock.RLock()
	if element, ok := prov.sessions[oldSid]; ok {
		go prov.Update(oldSid)
		prov.lock.RUnlock()
		prov.lock.Lock()
		element.Value.(*memoryStorage).sid = sid
		prov.sessions[sid] = element
		delete(prov.sessions, oldSid)
		prov.lock.Unlock()
		return element.Value.(*memoryStorage), nil
	}
	prov.lock.RUnlock()
	prov.lock.Lock()
	newStore := &memoryStorage{sid: sid, timeAccessed: time.Now(), value: make(map[interface{}]interface{})}
	element := prov.list.PushFront(newStore)
	prov.sessions[sid] = element
	prov.lock.Unlock()
	return newStore, nil
}

// Destroy delete session store in memory session by id
func (prov *diskProvider) Destroy(sid string) error {
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
func (prov *diskProvider) All() int {
	return prov.list.Len()
}

// Update expand Time of session store by id in memory session
func (prov *diskProvider) Update(sid string) error {
	prov.lock.Lock()
	defer prov.lock.Unlock()
	if element, ok := prov.sessions[sid]; ok {
		element.Value.(*memoryStorage).timeAccessed = time.Now()
		prov.list.MoveToFront(element)
		return nil
	}
	return nil
}

// GC clean expired session stores in memory session
func (prov *diskProvider) GC() {
	prov.lock.RLock()
	for {
		element := prov.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*memoryStorage).timeAccessed.Unix() + prov.maxLifetime) < time.Now().Unix() {
			prov.lock.RUnlock()
			prov.lock.Lock()
			prov.list.Remove(element)
			delete(prov.sessions, element.Value.(*memoryStorage).sid)
			prov.lock.Unlock()
			prov.lock.RLock()
		} else {
			break
		}
	}
	prov.lock.RUnlock()
}
