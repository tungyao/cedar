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
}
type Son struct {
	key         string // /a
	path        string // /a
	deep        int    //
	child       map[string]*Son
	terminal    bool
	method      string
	handlerFunc http.HandlerFunc
	handler     http.Handler
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
		pattern: "/",
	}
}
func (mux *Trie) Insert(method string, path string, handlerFunc http.HandlerFunc, handler http.Handler) {
	switch method {
	case http.MethodGet:
		fmt.Println(method, "\t", path[:len(path)-3])
	case http.MethodConnect:
		fmt.Println(method, "\t", path[:len(path)-7])
	case http.MethodDelete:
		fmt.Println(method, "\t", path[:len(path)-6])
	case http.MethodHead:
		fmt.Println(method, "\t", path[:len(path)-4])
	case http.MethodOptions:
		fmt.Println(method, "\t", path[:len(path)-7])
	case http.MethodPost:
		fmt.Println(method, "\t", path[:len(path)-4])
	case http.MethodPut:
		fmt.Println(method, "\t", path[:len(path)-3])
	case http.MethodTrace:
		fmt.Println(method, "\t", path[:len(path)-5])
	}
	son := mux.root // son 是指针，不是普通变量
	pattern := strings.TrimPrefix(path, "/")
	res := strings.Split(pattern, mux.pattern)
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
					method:      "",
					handlerFunc: nil,
					handler:     nil,
				}
				tson = son.child[key]
			}
			son = son.child[key]
		}
	}
	tson.key = path
	tson.method = method
	tson.terminal = true
	tson.handler = handler
	tson.handlerFunc = handlerFunc
}

func (mux *Trie) Find(key string) (string, http.HandlerFunc, http.Handler) {
	son := mux.root
	pattern := strings.TrimPrefix(key, "/")
	res := strings.Split(pattern, mux.pattern)
	path := ""
	var han http.HandlerFunc = nil
	var hand http.Handler = nil
	var method string
	// var is_delete bool
	if son.key != key {
		for _, key := range res {
			if son.child[key] == nil {
				return "", nil, nil
			} else {
				path += son.child[key].key
				han = son.child[key].handlerFunc
				hand = son.child[key].handler
				method = son.child[key].method
			}
			son = son.child[key]
		}
	} else {
		return son.method, son.handlerFunc, son.handler
	}
	return method, han, hand
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
