package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"crypto/md5"
)

type sessionUUID [16]byte

func (id sessionUUID) string() string {
	bytes := [16]byte(id)
	str := fmt.Sprintf("%x%x%x%x%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
	return strings.ToUpper(str)
}

func md5Bytes(s string) []byte {
	h := md5.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

func uuidRandBytes() sessionUUID {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return sessionUUID{0}
	}
	bytes := md5Bytes(base64.URLEncoding.EncodeToString(b))
	if len(bytes) != 16 {
		return sessionUUID{0}
	}
	var uuidBytes [16]byte
	copy(uuidBytes[:], bytes)
	return sessionUUID(uuidBytes)
}

func newSessionId() sessionUUID {
	if runtime.GOOS == "windows" {
		uuid := uuidRandBytes()
		return uuid
	}
	f, err := os.Open("/dev/urandom")
	if err != nil {
		return uuidRandBytes()
	}
	defer f.Close()

	b := []byte{}
	_, err = f.Read(b)
	if err != nil || len(b) != 16 {
		return uuidRandBytes()
	}
	uuid := sessionUUID{}
	copy(uuid[:], b)
	return uuid
}