package ultimate_cedar

import (
	"log"
	"net/http"
	"strings"
)

type tree struct {
	Router *router
	Map    map[string]Handler
}

func setMethod(mth string, handler Handler) method {
	m := method{}
	switch mth {
	case "GET":
		m.GET = handler
		break
	case "POST":
		m.POST = handler
		break
	case "DELETE":
		m.DELETE = handler
		break
	case "HEAD":
		m.HEAD = handler
		break
	case "OPTIONS":
		m.OPTIONS = handler
		break
	case "PUT":
		m.PUT = handler
		break
	case "PATCH":
		m.PATCH = handler
		break
	}
	return m
}

// 专门用来存放
func (t *tree) append(mth, path string, handler Handler) {
	if strings.Index(path, ":") == -1 {
		t.Map[mth+path] = handler
		return
	}
	// 要处理两种状态 带 : 的
	rut := t.Router
	for _, s := range strings.Split(path, "/") {
		if s == "" {
			continue
		}
		var rx *router
		// 这就是需要进行匹配的
		if s[0] == ':' {
			rx = &router{
				Next:       nil,
				Method:     setMethod(mth, handler),
				Path:       s,
				IsMatching: true,
				URLData:    make(map[string]string),
			}
			rx.URLData[s[1:]] = ""
		} else {
			rx = &router{
				Next:       nil,
				Method:     setMethod(mth, handler),
				Path:       s,
				IsMatching: false,
			}
		}
		log.Println(rut == nil, rut)
		if rut == nil {
			rut = rx
		} else {
			rut = rut.Next
		}

	}
	log.Println(rut)
}

// 这里有个更快的算法 用hash算法
func (t *tree) find(path, method string) Handler {
	if h, ok := t.Map[method+path]; ok {
		return h
	}

	// 查找不到必定需要模糊匹配
	return nil
}
func (t *tree) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := t.find(r.URL.Path, r.Method)
	if handler == nil {
		w.WriteHeader(404)
		return
	}
	handler(w, r)
}
