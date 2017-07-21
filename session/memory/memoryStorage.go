package memory

import (
	"net/http"
	"sync"
	"time"
)

// storage memory session store.
// it saved sessions in a map in memory.
type storage struct {
	sid          string                 //session id
	timeAccessed time.Time              //last access Time
	value        map[string]interface{} //session store
	lock         sync.RWMutex
}

// Set Value to memory session
func (st *storage) Set(key string, value interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value[key] = value
	return nil
}

// Get Value from memory session by key
func (st *storage) Get(key string) interface{} {
	st.lock.RLock()
	defer st.lock.RUnlock()
	if v, ok := st.value[key]; ok {
		return v
	}
	return nil
}

// Delete in memory session by key
func (st *storage) Delete(key string) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	delete(st.value, key)
	return nil
}

// Flush clear all values in memory session
func (st *storage) Flush() error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value = make(map[string]interface{})
	return nil
}

// ID get this id of memory session store
func (st *storage) ID() string {
	return st.sid
}

// Release Implement method, no used.
func (st *storage) Release(w http.ResponseWriter) {
}
