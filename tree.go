package cedar

import (
	"fmt"
	_ "log"
	"net/http"
	"strings"
)

type Trie struct {
	num     int64
	pattern string
	root    *Son
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
	return &Trie{
		num: 1,
		root: NewSon("GET", "/", func(writer http.ResponseWriter, request *http.Request) {
			_, _ = fmt.Fprint(writer, "index")
		}, nil, 1),
		pattern: "/",
	}
}
func (mux *Trie) Insert(method string, path string, handlerFunc http.HandlerFunc, handler http.Handler) {
	son := mux.root //son 是指针，不是普通变量
	pattern := strings.TrimPrefix(path, "/")
	res := strings.Split(pattern, mux.pattern)
	if son.key != path { //匹配不成功才加入链表
		for _, key := range res { //遍历数组
			if son.child[key] == nil { //第一个son节点是不是空 ，如果是数据和节点key放进去
				node := NewSon(method, key, handlerFunc, handler, son.deep+1) //生成新的节点数据
				node.child = make(map[string]*Son)                            //初始化该节点的内存
				node.terminal = false                                         //false 表面 我下面还有节点
				son.child[key] = node                                         //将数据放入刚刚初始化的节点
			}
			son = son.child[key] //将这个子节点作为下一次遍历的son父节点）
		}
	}
	son.terminal = true
	son.handlerFunc = handlerFunc
	son.handler = handler
	son.path = path
	son.method = method
}

func (mux *Trie) Find(key string) (string, http.HandlerFunc, http.Handler) {
	son := mux.root
	pattern := strings.TrimPrefix(key, "/")
	res := strings.Split(pattern, mux.pattern)
	path := ""
	var han http.HandlerFunc
	var hand http.Handler
	var method string
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
