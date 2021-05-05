package ultimate_cedar

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type tree struct {
	Router map[string]*router
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
	spt := strings.Split(strings.TrimPrefix(path, "/"), "/")
	fmt.Println("spt=>", spt)
	var sp string
	for _, s := range spt {
		var rx router
		// 这就是需要进行匹配的
		if s[0] == ':' {
			rx = router{
				Next:       nil,
				Method:     setMethod(mth, handler),
				Path:       path,
				IsMatching: true,
				Key:        s,
				URLData:    make(map[string]string),
			}
			rx.URLData[s[1:]] = ""
		} else {
			rx = router{
				Next:       nil,
				Method:     setMethod(mth, handler),
				Path:       path,
				Key:        s,
				IsMatching: false,
			}
		}
		if t.Router == nil {
			t.Router = make(map[string]*router)
			t.Router[s] = new(router)
			t.Router[s].Next = rx.Next
			t.Router[s].Path = rx.Path
			t.Router[s].URLData = rx.URLData
			t.Router[s].Method = rx.Method
			t.Router[s].Key = rx.Key
			t.Router[s].IsMatching = rx.IsMatching
			rut = t.Router
		} else {
			// 找到上部分
			fmt.Println("----", rut, s, sp)
			if rut[sp].Next == nil {
				rut[sp].Next = make(map[string]*router)
				rut[sp].Next[s] = new(router)
			}
			rut[sp].Next[s].URLData = rx.URLData
			rut[sp].Next[s].Path = rx.Path
			rut[sp].Next[s].Method = rx.Method
			rut[sp].Next[s].Key = rx.Key
			rut[sp].Next[s].IsMatching = rx.IsMatching
			rut = rut[sp].Next
		}
		sp = s

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
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write(bytes.NewBufferString(`{"x":404,"msg":"not fount"}`).Bytes())
		return
	}
	for t.Router != nil {

	}
	handler(w, r)
}
