package session

import (
	"sync"
	"time"
	"os"
	"fmt"
	"io/ioutil"
	"strings"
	"path"
	"errors"
)

// memoryProvider Implement the provider interface
type diskProvider struct {
	lock        sync.RWMutex            // locker
	sessions    map[string]*diskStorage // map in memory
	maxLifetime int64
	savePath    string
}

// Init init memory session
func (prov *diskProvider) Init(maxLifeTime int64, _ string) error {
	prov.savePath = path.Clean(strings.Replace(prov.savePath, "\\", "/", -1))
	if len(prov.savePath) == 0 {
		return errors.New("Invalid session provider folder: the 'savePath' cannot be empty.")
	}
	stat,err := os.Stat(prov.savePath)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(prov.savePath, 0666)
		} else {
			return err
		}
	} else {
		if !stat.IsDir() {
			fmt.Errorf("Invalid session provider folder: %s", prov.savePath)
		}
	}
	prov.maxLifetime = maxLifeTime
	prov.savePath = prov.savePath
	prov.sessions = make(map[string]*diskStorage)

	infos,err := ioutil.ReadDir(prov.savePath)
	if err != nil {
		return err
	}
	if len(infos) > 0 {
		for _, info := range infos {
			if info.IsDir() || !strings.HasSuffix(info.Name(), ".session") {
				continue
			}
			filePath := path.Clean(strings.Replace(info.Name(), "\\", "/", -1))
			fileName := filePath[strings.LastIndex(filePath, "/") + 1:]
			sid := fileName[0:strings.Index(fileName, ".")]
			storage := prov.newStorage(sid)
			data:= storage.readValue()

			storage.timeAccessed = data.Time
			storage.value = data.Value
			prov.sessions[sid] = storage
		}
	}
	return nil
}

func (prov *diskProvider) newStorage(sid string) *diskStorage {
	newStore := &diskStorage{
		sid: sid,
		timeAccessed: time.Now(),
		value: nil,
		diskDir: prov.savePath,
	}
	prov.sessions[sid] = newStore
	return newStore
}

// Read get memory session store by sid
func (prov *diskProvider) Read(sid string) Storage {
	prov.lock.RLock()
	if element, ok := prov.sessions[sid]; ok {
		go prov.Update(sid)
		prov.lock.RUnlock()
		return element
	}
	prov.lock.RUnlock()
	prov.lock.Lock()
	newStore := prov.newStorage(sid)
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

// Regenerate generate new sid for session store in session
func (prov *diskProvider) Regenerate(oldSid, sid string) (Storage, error) {
	prov.lock.RLock()
	if element, ok := prov.sessions[oldSid]; ok {
		go prov.Update(oldSid)
		prov.lock.RUnlock()
		prov.lock.Lock()
		element.delFile()
		element.sid = sid
		element.timeAccessed = time.Now()
		element.saveValue()
		prov.sessions[sid] = element
		delete(prov.sessions, oldSid)
		prov.lock.Unlock()
		return element, nil
	}
	prov.lock.RUnlock()
	prov.lock.Lock()
	newStore := prov.newStorage(sid)
	prov.lock.Unlock()
	return newStore, nil
}

// Destroy delete session store in session by id
func (prov *diskProvider) Destroy(sid string) error {
	prov.lock.Lock()
	defer prov.lock.Unlock()
	if element, ok := prov.sessions[sid]; ok {
		delete(prov.sessions, sid)
		element.delFile()
		return nil
	}
	return nil
}

// All get count number of memory session
func (prov *diskProvider) All() int {
	return len(prov.sessions)
}

// Update expand Time of session store by id in memory session
func (prov *diskProvider) Update(sid string) error {
	prov.lock.Lock()
	defer prov.lock.Unlock()
	if element, ok := prov.sessions[sid]; ok {
		element.timeAccessed = time.Now()
		element.saveValue()
	}
	return nil
}

// GC clean expired session stores in memory session
func (prov *diskProvider) GC() {
	prov.lock.RLock()
	for _, element := range prov.sessions {
		if element == nil {
			break
		}
		if (element.timeAccessed.Unix() + prov.maxLifetime) < time.Now().Unix() {
			prov.lock.RUnlock()
			prov.lock.Lock()
			delete(prov.sessions, element.sid)
			element.delFile()
			prov.lock.Unlock()
			prov.lock.RLock()
		} else {
			break
		}
	}
	prov.lock.RUnlock()
}