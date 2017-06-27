package session

import (
	"net/http"
)

// Storage the session store interface
type Storage interface {
	Set(key string, value interface{}) error //set session Value
	Get(key string) interface{}  //get session Value
	Delete(key string) error     //delete session Value
	ID() string                       //back current sessionID
	Release(w http.ResponseWriter)    //release the resource & save data to provider & return the data
	Flush() error                     //delete all data
}