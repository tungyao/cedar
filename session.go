package cedar

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"sync"
	"time"

	spruce "github.com/tungyao/spruce-light"
)

// 这里是session组件的存
// 默认是 启用 开源通过 r.StopSession关闭
// 现版本是只能启用 LOCAL模式
var (
	X  *spruce.Hash
	OP int = -1
)

const (
	LOCAL = iota
	SpruceLocal
	SPRUCE
)

func NewSession(types int) {
	OP = types
	switch types {
	case LOCAL:
		X = spruce.CreateHash(4096)
	case SPRUCE:
	case SpruceLocal:
		// KV, _ = ap.NewPool(args[0].(int), args[1].(string))
	}
}

// stop session

func (mux *Trie) StopSession() {
	mux.sessionx = nil
}

// struct
type sessionx struct {
	sync.Mutex
	Self   []byte
	op     int
	Domino string
}
type Session struct {
	sync.RWMutex
	Cookie string
}
type SX struct {
	Key  string
	Body interface{}
}

// func (mux *SessionX) Delete(path string, handlerFunc http.HandlerFunc, handler http.Handler) {
//	mux.tree.Delete(mux.path+path, handlerFunc, handler)
// }
// UUID 64 bit
// 8-4-4-12 16hex string
func (si *sessionx) CreateUUID(xtr []byte) []byte {
	str := fmt.Sprintf("%x", xtr)
	strLow := ComplementHex(str[:(len(str)-1)/3], 8)
	strMid := ComplementHex(str[(len(str)-1)/3:(len(str)-1)*2/3], 4)
	si.Lock()
	defer si.Unlock()
	<-time.After(1 * time.Nanosecond)
	ti := time.Now().UnixNano()
	return []byte(fmt.Sprintf("%s-%x-%s-%s", strLow, ti, strMid, si.Self))
}
func CreateUUID(xtr []byte) []byte {
	str := fmt.Sprintf("%x", xtr)
	strLow := ComplementHex(str[:(len(str)-1)/3], 8)
	strMid := ComplementHex(str[(len(str)-1)/3:(len(str)-1)*2/3], 4)
	<-time.After(1 * time.Nanosecond)
	ti := time.Now().UnixNano()
	return []byte(fmt.Sprintf("%s-%x-%s-%s", strLow, ti, strMid, ""))
}
func Random() []byte {
	str := fmt.Sprintf("%x", newId())
	// strLow := ComplementHex(str[:(len(str)-1)/3], 8)
	strMid := ComplementHex(str[(len(str)-1)/3:(len(str)-1)*2/3], 4)
	<-time.After(1 * time.Nanosecond)
	ti := time.Now().UnixNano()
	return []byte(fmt.Sprintf("%x%s%s", ti, strMid, newId()))
}
func ComplementHex(s string, x int) string {
	if len(s) == x {
		return s
	}
	if len(s) < x {
		for i := 0; i < x-len(s); i++ {
			s += "0"
		}
	}
	if len(s) > x {
		return s[:x]
	}
	return s
}

// session function
func (sn Session) Set(key string, body []byte) {
	switch OP {
	case LOCAL:
		X.Set([]byte(sn.Cookie+key), body, 3600)
	case SPRUCE:
	case SpruceLocal:
		// kvSet([]byte(sn.Cookie+key), body, 3600)
	}
}
func (sn Session) Get(key string) interface{} {
	switch OP {
	case LOCAL:
		return X.Get([]byte(sn.Cookie + key))
	case SPRUCE:
	case SpruceLocal:
		// return kvGet([]byte(sn.Cookie + key))
	}
	return []byte("")
}
func (sn Session) Flush(key string) interface{} {
	switch OP {
	case LOCAL:
		return X.Delete([]byte(sn.Cookie + key))
	case SPRUCE:
	case SpruceLocal:
		// return kvDelete([]byte(sn.Cookie + key))
	}
	return []byte("")

}

// other function
func Sha1(b []byte) []byte {
	h := sha1.New()
	h.Write(b)
	return []byte(fmt.Sprintf("%x", h.Sum(nil)))
}
func newId() []byte {
	d := "abcdef012345689"
	da := make([]byte, 4)
	for i := 0; i < 4; i++ {
		<-time.After(time.Nanosecond * 10)
		da[i] = d[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(15)]
	}
	return da
}

// func kvSet(key, body []byte, exp int) []byte {
// 	KV.Get().Write(spruce.EntrySet(key, body, exp))
// 	return KV.Get().Read()
// }
// func kvGet(key []byte) []byte {
// 	KV.Get().Write(spruce.EntryGet(key))
// 	return KV.Get().Read()
// }
// func kvDelete(key []byte) []byte {
// 	KV.Get().Write(spruce.EntryDelete(key))
// 	return KV.Get().Read()
// }
