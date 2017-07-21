package disk

import (
	"encoding/gob"
	"fmt"
	"github.com/simbory/mego/assert"
	"github.com/simbory/mego/session"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

// provider Implement the provider interface
type provider struct {
	lock        sync.RWMutex        // locker
	sessions    map[string]*storage // map in memory
	maxLifetime int64
	savePath    string
}

// Init init the current session provider
func (prov *provider) Init(maxLifeTime int64, _ string) error {
	prov.savePath = path.Clean(strings.Replace(prov.savePath, "\\", "/", -1))
	assert.NotEmpty("savePath", prov.savePath)
	stat, err := os.Stat(prov.savePath)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(prov.savePath, 0777)
		} else {
			return err
		}
	} else {
		if !stat.IsDir() {
			return fmt.Errorf("Invalid session provider folder: %s", prov.savePath)
		}
	}
	prov.maxLifetime = maxLifeTime
	prov.savePath = prov.savePath
	prov.sessions = make(map[string]*storage)

	infos, err := ioutil.ReadDir(prov.savePath)
	if err != nil {
		return err
	}
	if len(infos) > 0 {
		for _, info := range infos {
			if info.IsDir() || !strings.HasSuffix(info.Name(), ".session") {
				continue
			}
			filePath := path.Clean(strings.Replace(info.Name(), "\\", "/", -1))
			fileName := filePath[strings.LastIndex(filePath, "/")+1:]
			sid := fileName[0:strings.Index(fileName, ".")]
			storage := prov.newStorage(sid)
			data := storage.readValue()
			if data != nil {
				storage.timeAccessed = data.Time
				storage.value = data.Value
			} else {
				continue
			}
			prov.sessions[sid] = storage
		}
	}
	return nil
}

func (prov *provider) newStorage(sid string) *storage {
	newStore := &storage{
		sid:          sid,
		timeAccessed: time.Now(),
		value:        nil,
		diskDir:      prov.savePath,
	}
	return newStore
}

// Read get memory session store by sid
func (prov *provider) Read(sid string) session.Storage {
	prov.lock.RLock()
	if element, ok := prov.sessions[sid]; ok {
		go prov.Update(sid)
		prov.lock.RUnlock()
		return element
	}
	prov.lock.RUnlock()
	prov.lock.Lock()
	newStore := prov.newStorage(sid)
	prov.sessions[sid] = newStore
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

// Regenerate generate new sid for session store in session
func (prov *provider) Regenerate(oldSid, sid string) (session.Storage, error) {
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
	prov.sessions[sid] = newStore
	prov.lock.Unlock()
	return newStore, nil
}

// Destroy delete session store in session by id
func (prov *provider) Destroy(sid string) error {
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
func (prov *provider) All() int {
	return len(prov.sessions)
}

// Update expand Time of session store by id in memory session
func (prov *provider) Update(sid string) error {
	prov.lock.Lock()
	defer prov.lock.Unlock()
	if element, ok := prov.sessions[sid]; ok {
		element.timeAccessed = time.Now()
		element.saveValue()
	}
	return nil
}

// GC clean expired session stores in memory session
func (prov *provider) GC() {
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

func NewProvider(dir string) session.Provider {
	gob.RegisterName("__session_value", &storage_value{})
	return &provider{savePath: dir}
}
