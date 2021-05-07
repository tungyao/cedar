package ultimate_cedar

import (
	"net/http"
	"strings"
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
	CONNECT Handler
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

func (t *tree) Post(path string, handler Handler) {
	t.append("POST", path, handler)
}

func (t *tree) Delete(path string, handler Handler) {
	t.append("DELETE", path, handler)
}

func (t *tree) Head(path string, handler Handler) {
	t.append("HEAD", path, handler)
}

func (t *tree) Options(path string, handler Handler) {
	t.append("OPTIONS", path, handler)
}

func (t *tree) Put(path string, handler Handler) {
	t.append("PUT", path, handler)
}

func (t *tree) Patch(path string, handler Handler) {
	t.append("PATCH", path, handler)
}

func (t *tree) Trace(path string, handler Handler) {
	t.append("TRACE", path, handler)
}

func (t *tree) Connect(path string, handler Handler) {
	t.append("CONNECT", path, handler)
}

func (gup *Groups) Get(path string, handlerFunc Handler) {
	gup.Tree.Get(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc)
}

func (gup *Groups) Head(path string, handlerFunc Handler) {
	gup.Tree.Head(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc)
}

func (gup *Groups) Post(path string, handlerFunc Handler) {
	gup.Tree.Post(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc)
}

func (gup *Groups) Put(path string, handlerFunc Handler) {
	gup.Tree.Put(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc)
}

func (gup *Groups) Patch(path string, handlerFunc Handler) {
	gup.Tree.Patch(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc)
}

func (gup *Groups) Delete(path string, handlerFunc Handler) {
	gup.Tree.Delete(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc)
}

func (gup *Groups) Connect(path string, handlerFunc Handler) {
	gup.Tree.Connect(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc)
}

func (gup *Groups) Trace(path string, handlerFunc Handler) {
	gup.Tree.Trace(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc)
}

func (gup *Groups) Options(path string, handlerFunc Handler) {
	gup.Tree.Options(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc)
}

func (gup *Groups) Group(path string, fn func(Groups *Groups)) {
	g := new(Groups)
	g.Path = gup.Path + "/" + strings.TrimPrefix(path, "/")
	g.Tree = gup.Tree
	fn(g)
}

func (t *tree) Group(path string, fn func(groups *Groups)) {
	g := new(Groups)
	g.Tree = t
	g.Path = path
	fn(g)
}
