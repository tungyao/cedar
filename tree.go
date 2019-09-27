package tnwb

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type Trie struct {
	num  int64
	root *Son
	file *os.File
}
type Son struct {
	key      string // /a
	path     string // /a
	deep     int    //
	child    map[string]*Son
	terminal bool
	method   string
	handler  http.HandlerFunc
}

func NewSon(method string, path string, handler http.HandlerFunc, deep int) *Son {
	return &Son{
		key:     path,
		path:    path,
		deep:    deep,
		handler: handler,
		method:  method,
		child:   make(map[string]*Son),
	}
}
func NewRouter() *Trie {
	f, err := os.OpenFile("/var/log/twngo/log.log", os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
	}
	return &Trie{
		num: 1,
		root: NewSon("GET", "/", func(writer http.ResponseWriter, request *http.Request) {
			_, _ = fmt.Fprint(writer, "index")
		}, 1),
		file: f,
	}
}
func (mux *Trie) Insert(method string, path string, handler http.HandlerFunc) {
	son := mux.root //son 是指针，不是普通变量
	pattern := strings.TrimPrefix(path, "/")
	res := strings.Split(pattern, "/")
	if son.key != path { //匹配不成功才加入链表
		for _, key := range res { //遍历数组
			if son.child[key] == nil { //第一个son节点是不是空 ，如果是数据和节点key放进去
				node := NewSon(method, key, handler, son.deep+1) //生成新的节点数据
				node.child = make(map[string]*Son)               //初始化该节点的内存
				node.terminal = false                            //false 表面 我下面还有节点
				son.child[key] = node                            //将数据放入刚刚初始化的节点
			}
			son = son.child[key] //将这个子节点作为下一次遍历的son父节点）
		}
	}
	son.terminal = true
	son.handler = handler
	son.path = path
	son.method = method
}

func (mux *Trie) Find(key string) (string, http.HandlerFunc) {
	son := mux.root
	pattern := strings.TrimPrefix(key, "/")
	res := strings.Split(pattern, "/")
	path := ""
	var han http.HandlerFunc
	var method string
	if son.key != key {
		for _, key := range res {
			if son.child[key] == nil {
				return "", nil
			} else {
				path += son.child[key].key
				han = son.child[key].handler
				method = son.child[key].method
			}
			son = son.child[key]
		}
	} else {
		return son.method, son.handler
	}
	return method, han
}
