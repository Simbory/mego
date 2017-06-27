package session

import (
	"time"
	"sync"
	"net/http"
)

// memoryStorage memory session store.
// it saved sessions in a map in memory.
type memoryStorage struct {
	sid          string                      //session id
	timeAccessed time.Time                   //last access Time
	value        map[string]interface{} //session store
	lock         sync.RWMutex
}

// Set Value to memory session
func (st *memoryStorage) Set(key string, value interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value[key] = value
	return nil
}

// Get Value from memory session by key
func (st *memoryStorage) Get(key string) interface{} {
	st.lock.RLock()
	defer st.lock.RUnlock()
	if v, ok := st.value[key]; ok {
		return v
	}
	return nil
}

// Delete in memory session by key
func (st *memoryStorage) Delete(key string) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	delete(st.value, key)
	return nil
}

// Flush clear all values in memory session
func (st *memoryStorage) Flush() error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value = make(map[string]interface{})
	return nil
}

// ID get this id of memory session store
func (st *memoryStorage) ID() string {
	return st.sid
}

// Release Implement method, no used.
func (st *memoryStorage) Release(w http.ResponseWriter) {
}

