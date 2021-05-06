package ultimate_cedar

import (
	"bytes"
	"net/http"
	"strings"
)

type tree struct {
	Router map[string]*router
	Map    map[string]Handler
}

func exec(router2 *router, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		router2.Method.GET(w, r)
		break
	case "POST":
		router2.Method.POST(w, r)
		break
	case "DELETE":
		router2.Method.DELETE(w, r)
		break
	case "HEAD":
		router2.Method.HEAD(w, r)
		break
	case "OPTIONS":
		router2.Method.OPTIONS(w, r)
		break
	case "PUT":
		router2.Method.PUT(w, r)
		break
	case "PATCH":
		router2.Method.PATCH(w, r)
		break
	}
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
	spt := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for _, s := range spt {
		var rx router
		// 这就是需要进行匹配的
		if s[0] == ':' {
			rx = router{
				Next:       make(map[string]*router),
				Method:     setMethod(mth, handler),
				Path:       path,
				IsMatching: true,
				Key:        "*",
			}
		} else {
			rx = router{
				Next:       make(map[string]*router),
				Method:     setMethod(mth, handler),
				Path:       path,
				Key:        s,
				IsMatching: false,
			}
		}
		// 这是第一次加载
		if _, ok := rut[s]; ok {
			rut = t.Router[rx.Key].Next
		} else {
			rut[rx.Key] = &rx
			rut = rut[rx.Key].Next
		}
	}
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
		spt := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		rut := t.Router
		count := len(spt)
		for k, v := range spt {
			if rut["*"] != nil {
				if rut["*"].IsMatching {
					r.URL.Fragment = v
					if k == count-1 {
						exec(rut["*"], w, r)
						return
					}
				}
				rut = rut["*"].Next
			} else {
				if rut[v] == nil {
					goto end
				}
				if k == len(spt)-1 {
					exec(rut[v], w, r)
					return
				}
				rut = rut[v].Next
			}
		}
	}
end:
	w.WriteHeader(404)
	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(bytes.NewBufferString(`{"x":404,"msg":"not fount"}`).Bytes())
	return

}
