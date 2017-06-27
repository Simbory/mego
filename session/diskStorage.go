package session

import (
	"time"
	"sync"
	"net/http"
	"os"
	"bytes"
	"io/ioutil"
	"encoding/gob"
)

// diskStorage disk session store.
// it saved sessions in a map in disk.
type diskStorage struct {
	sid          string                      //session id
	timeAccessed time.Time                   //last access Time
	value        map[string]interface{} //session store
	diskDir      string
	lock         sync.RWMutex
}

type diskValue struct {
	Time  time.Time `json:"time"`
	Value map[string]interface{} `json:"value"`
}

func (st *diskStorage) diskFile() string {
	return st.diskDir + "/" + st.sid + ".session"
}

// Set Value to memory session
func (st *diskStorage) Set(key string, value interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	if st.value == nil {
		st.value = make(map[string]interface{})
	}
	st.value[key] = value
	st.timeAccessed = time.Now()
	st.saveValue()
	return nil
}

// Get Value from memory session by key
func (st *diskStorage) Get(key string) interface{} {
	st.lock.RLock()
	defer st.lock.RUnlock()
	if st.value == nil {
		value := st.readValue()
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
func (st *diskStorage) Delete(key string) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	delete(st.value, key)
	st.saveValue()
	return nil
}

// Flush clear all values in session
func (st *diskStorage) Flush() error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value = make(map[string]interface{})
	st.saveValue()
	return nil
}

// ID get this id of memory session store
func (st *diskStorage) ID() string {
	return st.sid
}

// Release clear all the session data and release the disk space
func (st *diskStorage) Release(w http.ResponseWriter) {
	st.delFile()
	st.Flush()
}

func (st *diskStorage) delFile() {
	stat,err := os.Stat(st.diskFile())
	if err != nil {
		return
	}
	if stat.IsDir() {
		return
	}
	err = os.Remove(st.diskFile())
	if err != nil {
		go st.writeLog("Failed to remove session file form disk", err)
	}
}

func (st *diskStorage) writeLog(log string, err error)  {
	f,err := os.OpenFile(st.diskDir + "/session.log", os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	if err != nil {
		return
	}
	defer f.Close()
	now := time.Now().Format(time.RFC1123Z)
	buf := bytes.NewBuffer(nil)
	buf.WriteString(now)
	buf.WriteString(": " + log + "\r\n")
	if err != nil {
		for i := 0;i<len(now) + 2; i++ {
			buf.WriteByte(' ')
		}
		buf.WriteString("error - " + err.Error() + "\r\n")
	}
	for i := 0;i<len(now) + 2; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteString("session - " + st.sid + "\r\n")
	f.Write(buf.Bytes())
}

func (st *diskStorage) saveValue() {
	go func() {
		var err error
		if st.value == nil || len(st.value) == 0 {
			os.Remove(st.diskFile())
		} else {
			buf := bytes.NewBuffer(nil)
			gobEncoder := gob.NewEncoder(buf)
			err = gobEncoder.Encode(&diskValue{Time: st.timeAccessed, Value: st.value})
			if err == nil {
				err = ioutil.WriteFile(st.diskFile(), buf.Bytes(), 0666)
			}
		}
		if err != nil {
			logStr := "failed to save the session to disk"
			st.writeLog(logStr, err)
		}
	}()
}

func (st *diskStorage) readValue() *diskValue {
	stat,err := os.Stat(st.diskFile())
	if err != nil {
		if os.IsNotExist(err){
			return nil
		}
		go st.writeLog("failed to read the session", err)
		return nil
	}
	if stat.IsDir() {
		return nil
	}

	f,err := os.OpenFile(st.diskFile(), os.O_RDONLY,0644)
	if err != nil {
		go st.writeLog("failed to read the session file", err)
		return nil
	}
	defer f.Close()

	gobDecoder := gob.NewDecoder(f)
	value := &diskValue{}
	err = gobDecoder.Decode(value)
	if err != nil {
		go st.writeLog("failed to read the session file", err)
		return nil
	}
	return value
}