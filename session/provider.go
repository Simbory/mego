package session

// Provider contains global session methods and saved SessionStores.
// it can operate a Storage by its id.
type Provider interface {
	Init(gcLifetime int64, config string) error
	Read(sid string) Storage
	Exist(sid string) bool
	Regenerate(oldSid, sid string) (Storage, error)
	Destroy(sid string) error
	All() int //get all active session
	Update(sid string) error
	GC()
}