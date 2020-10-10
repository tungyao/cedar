package cedar

import (
	"fmt"
	_ "log"
	"net/http"
	"strings"
	"sync"
)

type HandlerFunc func(http.ResponseWriter, *http.Request, *Core)

type Trie struct {
	num        int64
	pattern    string
	root       *Son
	globalFunc []*GlobalFunc
	middle     map[string]func(w http.ResponseWriter, r *http.Request, c *Core) bool
	sessions   *sessions
}
type Son struct {
	key           string // /a
	path          string // /a
	deep          int    //
	child         map[string]*Son
	terminal      bool
	method        string
	middle        string
	fuzzy         bool
	fuzzyPosition string
	handlerFunc   HandlerFunc
	handler       http.Handler
	next          map[string]*Son
}
type GlobalFunc struct {
	Name string
	Fn   func(w http.ResponseWriter, r *http.Request, co *Core) error
}

func NewSon(method string, path string, handlerFunc HandlerFunc, handler http.Handler, deep int) *Son {
	return &Son{
		key:         path,
		path:        path,
		deep:        deep,
		handlerFunc: handlerFunc,
		handler:     handler,
		method:      method,
		child:       make(map[string]*Son),
		next:        make(map[string]*Son),
	}
}
func NewRouter(sessionSetting ...string) *Trie {
	fmt.Println("-----------Register router-----------")
	self := "localhost"
	domino := "localhost"
	if len(sessionSetting) > 1 {
		self = sessionSetting[0]
		domino = sessionSetting[1]
	}
	NewSession(0)
	return &Trie{
		num: 1,
		root: NewSon("GET", "/", func(writer http.ResponseWriter, request *http.Request, r *Core) {
			_, _ = fmt.Fprint(writer, "index")
		}, nil, 1),
		middle:  make(map[string]func(w http.ResponseWriter, r *http.Request, c *Core) bool),
		pattern: "/",
		sessions: &sessions{
			Mutex:  sync.Mutex{},
			Self:   byt(self),
			op:     0,
			Domino: domino,
		},
	}
}

func (mux *Trie) insert(method string, path string, handlerFunc HandlerFunc, handler http.Handler, name []string) {
	switch method {
	case http.MethodGet:
		fmt.Println(method, "\t", path)
	case http.MethodConnect:
		fmt.Println(method, "\t", path)
	case http.MethodDelete:
		fmt.Println(method, "\t", path)
	case http.MethodHead:
		fmt.Println(method, "\t", path)
	case http.MethodOptions:
		fmt.Println(method, "\t", path)
	case http.MethodPost:
		fmt.Println(method, "\t", path)
	case http.MethodPut:
		fmt.Println(method, "\t", path)
	case http.MethodTrace:
		fmt.Println(method, "\t", path)
	}
	pattern := strings.TrimPrefix(path, "/")
	res := strings.Split(pattern, mux.pattern)
	tson := mux.root
	if tson.key != path {
		for _, key := range res {
			_, fuzB := fPostion(key)
			if fuzB { // 具有模糊查找功能 直接把key变成 *
				key = "*"
			}
			if tson.child[key] == nil {
				tson.child[key] = NewSon(method, key, nil, nil, 0)
				tson = tson.child[key]
			} else {
				// 这里又两种情况 可能key出现重复 需要放在next中
				tson = tson.child[key]
			}
		}
	}
	if tson.key == res[len(res)-1] && tson.method != method {
		tson.next[method] = NewSon(method, res[len(res)-1], nil, nil, 0)
		tson = tson.next[method]
	}
	fuzS, fuzB := fPostion(path)
	if fuzB { // 具有模糊查找功能 直接把key变成 *
		tson.key = "*"
	}
	tson.handler = handler
	tson.handlerFunc = handlerFunc
	tson.method = method
	tson.terminal = true
	tson.fuzzy = fuzB
	tson.fuzzyPosition = fuzS
	if len(name) > 0 {
		tson.middle = name[0]
	}
}

func (mux *Trie) Find(key string, methods string) (string, HandlerFunc, http.Handler, string, string, bool) {
	son := mux.root
	pattern := strings.TrimPrefix(key, "/")
	res := strings.Split(pattern, mux.pattern)
	path := ""
	param := ""
	if son.key != key && !son.fuzzy {
		paths := ""
		for _, key := range res {
			if son == nil {
				return "", nil, nil, "", "", false
			}
			if son.child["*"] != nil {
				param = getParam(paths, pattern)
				fmt.Println(paths, pattern)
				son = son.child["*"]
				continue
			}
			if son.child[key] != nil {
				path += son.child[key].key
				param = getParam(paths, pattern)
				paths += key + "/"
				son = son.child[key]
			} else {
				son = son.child[key]
			}
		}
	} else {
		param = getParam(son.fuzzyPosition, pattern)
		return son.method, son.handlerFunc, son.handler, son.middle, param, true
	}
	if son == nil {
		return "", nil, nil, "", "", false
	}
	if son.method == methods {
		goto end
	}
	son = son.next[methods]
	goto end
end:
	return son.method, son.handlerFunc, son.handler, son.middle, param, true
}
func (mux *Trie) Middle(name string, fn func(w http.ResponseWriter, r *http.Request, co *Core) bool) {
	mux.middle[name] = fn
}
func SplitString(str []byte, p []byte) []string {
	group := make([]string, 0)
	for i := 0; i < len(str); i++ {
		if str[i] == p[0] && i < len(str)-len(p) {
			if len(p) == 1 {
				return []string{string(str[:i]), string(str[i+1:])}
			} else {
				for j := 1; j < len(p); i++ {
					if str[i+j] != p[j] {
						continue
					}
					return []string{string(str[:i]), string(str[i+len(p):])}
				}
			}
		} else {
			continue
		}
	}
	return group
}
func getParam(position, path string) string {
	if len(position) > len(path)-1 {
		return ""
	}
	//kx := 0
	//for k, v := range path[len(position):] {
	//	if v == '/' {
	//		kx = k
	//		break
	//	}
	//}
	return path[len(position):]
}
func fPostion(path string) (string, bool) {
	for k, v := range path {
		if v == ':' {
			return path[k:], true
		}
	}
	return path, false
}
