package session

import (
	"time"
	"sync"
	"net/http"
	"os"
	"encoding/gob"
	"bytes"
	"io/ioutil"
)

// diskStorage disk session store.
// it saved sessions in a map in disk.
type diskStorage struct {
	sid          string                      //session id
	timeAccessed time.Time                   //last access Time
	value        map[interface{}]interface{} //session store
	diskFile     string
	lock         sync.RWMutex
}

type diskValue struct {
	Time  time.Time
	Value map[interface{}]interface{}
}

// Set Value to memory session
func (st *diskStorage) Set(key, value interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value[key] = value
	st.timeAccessed = time.Now()
	return st.saveValue()
}

// Get Value from memory session by key
func (st *diskStorage) Get(key interface{}) interface{} {
	st.lock.RLock()
	defer st.lock.RUnlock()
	if st.value == nil {
		value,err := st.readValue()
		if err != nil {
			panic(err)
			return nil
		}
		if value == nil {
			return nil
		}
		st.value = value.Value
		st.timeAccessed = time.Now()
		st.saveValue()
	}
	if v, ok := st.value[key]; ok {
		return v
	}
	return nil
}

// Delete in memory session by key
func (st *diskStorage) Delete(key interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	delete(st.value, key)
	return st.saveValue()
}

// Flush clear all values in memory session
func (st *diskStorage) Flush() error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value = make(map[interface{}]interface{})
	return st.saveValue()
}

// ID get this id of memory session store
func (st *diskStorage) ID() string {
	return st.sid
}

// Release Implement method, no used.
func (st *diskStorage) Release(w http.ResponseWriter) {
}

func (st *diskStorage) saveValue() error {
	if st.value == nil || len(st.value) == 0 {
		return os.Remove(st.diskFile)
	}
	buf := bytes.NewBuffer(nil)
	gobEncoder := gob.NewEncoder(buf)
	err := gobEncoder.Encode(&diskValue{st.timeAccessed, st.value})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(st.diskFile, buf.Bytes(), 0666)
}

func (st *diskStorage) readValue() (*diskValue, error) {
	stat,err := os.Stat(st.diskFile)
	if err != nil {
		return nil,err
	}
	if stat.IsDir() {
		return nil, nil
	}

	f,err := os.Open(st.diskFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gobDecoder := gob.NewDecoder(f)
	value := &diskValue{}
	err = gobDecoder.Decode(value)
	if err != nil {
		return nil,err
	}
	return value, nil
}