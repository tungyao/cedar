package cedar

import (
	"fmt"
	_ "log"
	"net/http"
	"strings"
)

type Trie struct {
	num        int64
	pattern    string
	root       *Son
	globalFunc []*GlobalFunc
	middle     map[string]func(w http.ResponseWriter, r *http.Request) bool
}
type Son struct {
	key           string // /a
	path          string // /a
	deep          int    //
	child         map[string]*Son
	terminal      bool
	method        string
	midle         string
	fuzzy         bool
	fuzzyPosition string
	handlerFunc   http.HandlerFunc
	handler       http.Handler
}
type GlobalFunc struct {
	Name string
	Fn   func(w http.ResponseWriter, r *http.Request) error
}

func NewSon(method string, path string, handlerFunc http.HandlerFunc, handler http.Handler, deep int) *Son {
	return &Son{
		key:         path,
		path:        path,
		deep:        deep,
		handlerFunc: handlerFunc,
		handler:     handler,
		method:      method,
		child:       make(map[string]*Son),
	}
}
func NewRouter() *Trie {
	fmt.Println("-----------Register router-----------")
	return &Trie{
		num: 1,
		root: NewSon("GET", "/", func(writer http.ResponseWriter, request *http.Request) {
			_, _ = fmt.Fprint(writer, "index")
		}, nil, 1),
		middle:  make(map[string]func(w http.ResponseWriter, r *http.Request) bool),
		pattern: "/",
	}
}
func (mux *Trie) Insert(method string, path string, handlerFunc http.HandlerFunc, handler http.Handler, name []string) {
	switch method {
	case http.MethodGet:
		fmt.Println(method, "\t", path[:len(path)-4])
	case http.MethodConnect:
		fmt.Println(method, "\t", path[:len(path)-8])
	case http.MethodDelete:
		fmt.Println(method, "\t", path[:len(path)-7])
	case http.MethodHead:
		fmt.Println(method, "\t", path[:len(path)-4])
	case http.MethodOptions:
		fmt.Println(method, "\t", path[:len(path)-8])
	case http.MethodPost:
		fmt.Println(method, "\t", path[:len(path)-5])
	case http.MethodPut:
		fmt.Println(method, "\t", path[:len(path)-4])
	case http.MethodTrace:
		fmt.Println(method, "\t", path[:len(path)-6])
	}
	son := mux.root
	pattern := strings.TrimPrefix(path, "/")
	res := strings.Split(pattern, mux.pattern)
	// res = res[:len(res)-1]
	tson := mux.root
	if son.key != path {
		for _, key := range res {
			if son.child[key] == nil {
				son.child[key] = &Son{
					key:         "",
					path:        "",
					deep:        0,
					child:       make(map[string]*Son),
					terminal:    false,
					midle:       "",
					method:      method,
					handlerFunc: nil,
					handler:     nil,
				}
				//fuzP, fuzB := fPostion(key)
				//tson = son.child[key]
				//tson.fuzzy = fuzB
				//tson.fuzzyPosition = fuzP
			}
			fuzP, fuzB := fPostion(key)
			son.fuzzyPosition = fuzP
			son.fuzzy = fuzB
			son.method = method
			son = son.child[key]
			tson = son
		}
	}
	tson.handler = handler
	tson.handlerFunc = handlerFunc
	tson.method = method
	tson.key = path
	tson.method = method
	tson.terminal = true
	fuzP, fuzB := fPostion(path)
	tson.fuzzyPosition = fuzP
	tson.fuzzy = fuzB
	if len(name) > 0 {
		tson.midle = name[0]
	}
}

func (mux *Trie) Find(key string) (string, http.HandlerFunc, http.Handler, string, string) {
	son := mux.root
	pattern := strings.TrimPrefix(key, "/")
	res := strings.Split(pattern, mux.pattern)
	path := ""
	param := ""
	var han http.HandlerFunc = nil
	var hand http.Handler = nil
	var method string
	if son.key != key && !son.fuzzy {
		swichs := false
		fuzzy := ""
		paths := ""
		for _, key := range res {
			if son.child[key] == nil && swichs == false {
				return "", nil, nil, "", ""
			}
			if fuzzy != "" {
				key = fuzzy
			}
			path += son.child[key].key
			han = son.child[key].handlerFunc
			hand = son.child[key].handler
			method = son.child[key].method
			if son.child[key].fuzzy {
				swichs = true
				fuzzy = son.child[key].fuzzyPosition
			} else {
				param = getParam(paths, pattern, method)
				swichs = false
				fuzzy = ""
			}
			paths += key + "/"
			son = son.child[key]
		}
	} else {
		param = getParam(son.fuzzyPosition, pattern, method)
		return son.method, son.handlerFunc, son.handler, son.midle, param
	}
	return method, han, hand, son.midle, param
}
func (mux *Trie) Middle(name string, fn func(w http.ResponseWriter, r *http.Request) bool) {
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
func getParam(position, path, method string) string {
	if len(position) > len(path)-len(method)-1 {
		return ""
	}
	return path[len(position) : len(path)-len(method)-1]
}
func fPostion(path string) (string, bool) {
	for k, v := range path {
		if v == ':' {
			return path[k:], true
		}
	}
	return path, false
}
