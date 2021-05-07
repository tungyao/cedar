package ultimate_cedar

import (
	"net/http"
)

// 在想能不能借助数组来存放路由
// /ab/bc
// /ab/bc/:id
// /ab/bc/:id/nv
// 只需要做到这个三种匹配就能完成绝大部分的路由匹配

// Handler 对原来的方法进行重写
type Handler func(ResponseWriter, *http.Request)
type ResponseWriter struct {
	http.ResponseWriter
	*Json
}
type Json struct{}

type method struct {
	GET     Handler
	POST    Handler
	DELETE  Handler
	HEAD    Handler
	OPTIONS Handler
	PUT     Handler
	PATCH   Handler
}

// 每一个节点应该存在以下几个字段
// Method method 方法可以有多个方便映射
// IsMatching bool 是否是需要做模糊匹配的
// Next 下一段路由
// Path 匹配的路由
type router struct {
	Next       map[string]*router
	Method     method
	Path       string
	IsMatching bool
	Key        string
}

func NewRouter() *tree {
	r := new(tree)
	r.Router = make(map[string]*router)
	r.Map = make(map[string]Handler)
	return r

}
func (t *tree) Get(path string, handler Handler) {
	t.append("GET", path, handler)
}

func (t *tree) POST(path string, handler Handler) {
	t.append("POST", path, handler)
}

func (t *tree) DELETE(path string, handler Handler) {
	t.append("DELETE", path, handler)
}

func (t *tree) HEAD(path string, handler Handler) {
	t.append("HEAD", path, handler)
}

func (t *tree) OPTIONS(path string, handler Handler) {
	t.append("OPTIONS", path, handler)
}

func (t *tree) PUT(path string, handler Handler) {
	t.append("PUT", path, handler)
}

func (t *tree) PATCH(path string, handler Handler) {
	t.append("PATCH", path, handler)
}
