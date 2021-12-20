package ultimate_cedar

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type tree struct {
	Router   map[string]*router
	Map      map[string]HandlerFunc
	template [2]func(err error) []byte
}
type Groups struct {
	Tree       *tree
	Path       string
	Middleware []MiddlewareChain
}

func exec(router2 *router, r Request) HandlerFunc {
	switch r.Method {
	case "GET":
		return router2.Method.GET
	case "POST":
		return router2.Method.POST
	case "DELETE":
		return router2.Method.DELETE
	case "HEAD":
		return router2.Method.HEAD
	case "OPTIONS":
		return router2.Method.OPTIONS
	case "PUT":
		return router2.Method.PUT
	case "PATCH":
		return router2.Method.PATCH
	case "CONNECT":
		return router2.Method.CONNECT
	}
	return nil
}
func setMethod(mth string, handler HandlerFunc) method {
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
	case "CONNECT":
		m.CONNECT = handler
	}
	return m
}

// 专门用来存放
func (t *tree) append(mth, path string, handlerFunc HandlerFunc, chain MiddlewareChain) {
	p := strings.TrimPrefix(path, "/")
	switch mth {
	case http.MethodGet:
		fmt.Println(mth, "\t", p)
	case http.MethodConnect:
		fmt.Println(mth, "\t", p)
	case http.MethodDelete:
		fmt.Println(mth, "\t", p)
	case http.MethodHead:
		fmt.Println(mth, "\t", p)
	case http.MethodOptions:
		fmt.Println(mth, "\t", p)
	case http.MethodPost:
		fmt.Println(mth, "\t", p)
	case http.MethodPut:
		fmt.Println(mth, "\t", p)
	case http.MethodTrace:
		fmt.Println(mth, "\t", p)
	}
	if strings.Index(path, ":") == -1 {
		t.Map[mth+p] = chain.Handler(handlerFunc)
		return
	}
	// 要处理两种状态 带 : 的
	rut := t.Router
	spt := strings.Split(p, "/")
	for _, s := range spt {
		var rx router
		// 这就是需要进行匹配的
		if s[0] == ':' {
			rx = router{
				Next:        make(map[string]*router),
				Method:      setMethod(mth, chain.Handler(handlerFunc)),
				Path:        path,
				IsMatching:  true,
				Key:         "*",
				MatchingKey: make(map[string]string),
			}
			rx.MatchingKey[s[1:]] = ""
		} else {
			rx = router{
				Next:       make(map[string]*router),
				Method:     setMethod(mth, chain.Handler(handlerFunc)),
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
func (t *tree) find(r Request) HandlerFunc {
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/")
	if h, ok := t.Map[r.Method+r.URL.Path]; ok {
		return h
	}
	spt := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	rut := t.Router
	count := len(spt)
	for k, v := range spt {
		if rut["*"] != nil {
			if rut["*"].IsMatching {
				for k := range rut["*"].MatchingKey {
					r.Data.set(k, v)
				}
				if k == count-1 {
					return exec(rut["*"], r)
				}
			}
			rut = rut["*"].Next
		} else {
			if rut[v] == nil {
				return nil
			}
			if k == len(spt)-1 {
				return exec(rut[v], r)
			}
			rut = rut[v].Next
		}
	}
	// 查找不到必定需要模糊匹配
	return nil
}
func (t *tree) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	e := &en{
		r:   r,
		ctx: ctx,
	}
	q := &qu{
		r:    r,
		ctx:  ctx,
		data: &pData{data: make(map[string]string)},
	}
	rq := Request{r, e, q, &data{data: make(map[string]string)}}
	handler := t.find(rq)
	if handler != nil {
		wx := ResponseWriter{
			ResponseWriter: w,
			Json:           new(Json),
		}
		wx.Json.t = t
		wx.writer = w
		wx.header = make(map[string]string)
		wx.header["content-type"] = "application/json"
		wx.status = 200
		handler(wx, rq)
		return
	}
	w.WriteHeader(404)
	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(bytes.NewBufferString(`{"x":404,"msg":"not fount"}`).Bytes())
}

func (t *tree) ErrorTemplate(f func(err error) []byte) {
	t.template[0] = f
}

var debug = false

func (t *tree) Debug() {
	log.Println("cedar will into the debug mode")
	os.Setenv("ultimate-cedar-debug", "yes")
	debug = true
}
