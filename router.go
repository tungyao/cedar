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
type Handler func(http.ResponseWriter, *http.Request)

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
	URLData    map[string]string
}

// split 将路由进行拆分 并加入到树中
func (r *router) split() {

}

func NewRouter() *tree {
	r := new(tree)
	r.Router = nil
	r.Map = make(map[string]Handler)
	return r

}
func (t *tree) Get(path string, handler Handler) {
	t.append("GET", path, handler)
}
