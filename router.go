package cedar

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	json "github.com/json-iterator/go"
)

// 在想能不能借助数组来存放路由
// /ab/bc
// /ab/bc/:id
// /ab/bc/:id/nv
// 只需要做到这个三种匹配就能完成绝大部分的路由匹配

// HandlerFunc Handler 对原来的方法进行重写
type HandlerFunc func(ResponseWriter, Request)
type ResponseWriter struct {
	http.ResponseWriter
	*Json
}
type Request struct {
	*http.Request
	*en
	Query *qu
	Data  *data
}

type pData struct {
	data map[string]string
}

func (_d *pData) Get(key string) string {
	return _d.data[key]
}
func (_d *pData) set(key, value string) {
	_d.data[key] = value
}

type data struct {
	data map[string]string
}

func (_d *data) Get(key string) string {
	return _d.data[key]
}
func (_d *data) set(key, value string) {
	_d.data[key] = value
}

type en struct {
	r   *http.Request
	ctx context.Context
}

func (e *en) Decode(any interface{}) error {
	b, err := io.ReadAll(e.r.Body)
	defer e.r.Body.Close()
	if err != nil {
		return err
	}
	if key := e.r.Header.Get("tyrant"); key != "" {
		dsk, err := base64.StdEncoding.DecodeString(string(b))
		if err != nil {
			return err
		}
		var (
			b1 int32
			b2 int32
			b3 int32
			d  int
		)
		var k = int32(len(key))
		var s = make([]rune, int(math.Floor(float64(len(dsk)/3))))
		for i := 0; i < len(s); i++ {
			b1 = int32(strings.IndexByte(key, dsk[d]))
			d++
			b2 = int32(strings.IndexByte(key, dsk[d]))
			d++
			b3 = int32(strings.IndexByte(key, dsk[d]))
			d++
			s[i] = b1*k*k + b2*k + b3
		}
		return json.Unmarshal(runes2str(s), any)
	}
	return json.Unmarshal(b, any)
}
func runes2str(s []int32) []byte {
	var p []byte
	for _, r := range s {
		buf := make([]byte, 3)
		if r > 128 {
			_ = utf8.EncodeRune(buf, r)
			p = append(p, buf...)
		} else {
			p = append(p, byte(r))
		}

	}
	return p
}

type qu struct {
	r    *http.Request
	ctx  context.Context
	data *pData
}

func (q *qu) Check(params ...string) (*pData, error) {
	v := q.r.URL.Query()
	if len(v) == 0 && len(params) != 0 {
		return nil, fmt.Errorf("query has been required")
	}
	for s, i := range q.r.URL.Query() {
		if !inArrayString(s, params) || len(i) == 0 {
			return nil, fmt.Errorf("%s must be required", s)
		}
		q.data.set(s, q.r.URL.Query().Get(s))
	}
	return q.data, nil
}

type Json struct {
	writer http.ResponseWriter
	header map[string]string
	status int
	data   []byte
	t      *tree
	sync.Once
}

func (j *Json) ContentType(contentType string) *Json {
	j.header["content-type"] = contentType
	return j
}
func (j *Json) AddHeader(name, value string) *Json {
	j.header[name] = value
	return j
}
func (j *Json) Data(any interface{}) *Json {
	switch any.(type) {
	case string:
		j.data = []byte(any.(string))
		return j
	case []byte:
		j.data = any.([]byte)
		return j
	case int:
		j.data = []byte(strconv.Itoa(any.(int)))
		return j
	case int64:
		j.data = []byte(strconv.Itoa(int(any.(int64))))
		return j
	case error:
		j.data = j.t.template[0](any.(error))
		return j
	}
	b, err := json.Marshal(any)
	if err != nil {
		log.Panicln(err)
	}
	j.data = b
	return j
}
func (j *Json) Status(status int) *Json {
	j.status = status
	return j
}
func (j *Json) Send() {
	for k, v := range j.header {
		if j.writer.Header().Get(k) != "" {
			j.writer.Header().Set(k, v)
		} else {
			j.writer.Header().Add(k, v)
		}
	}
	j.Do(func() {
		j.writer.WriteHeader(j.status)
	})
	_, _ = j.writer.Write(j.data)
}

func (j *Json) Encode(key string) *Json {
	j.header["tyrant"] = key
	var by = make([]byte, 0)
	var (
		b1 int32
		b2 int32
		b3 int32
	)
	var k = int32(len(key))
	for _, v := range bytes.Runes(j.data) {
		b1 = v % k
		v = (v - b1) / k
		b2 = v % k
		v = (v - b2) / k
		b3 = v % k
		by = append(by, key[b3], key[b2], key[b1])
	}
	j.data = []byte(base64.StdEncoding.EncodeToString(by))
	return j
}

type method struct {
	GET     HandlerFunc
	POST    HandlerFunc
	DELETE  HandlerFunc
	HEAD    HandlerFunc
	OPTIONS HandlerFunc
	PUT     HandlerFunc
	PATCH   HandlerFunc
	CONNECT HandlerFunc
}

// 每一个节点应该存在以下几个字段
// Method method 方法可以有多个方便映射
// IsMatching bool 是否是需要做模糊匹配的
// Next 下一段路由
// Path 匹配的路由
type router struct {
	Next        map[string]*router
	Method      method
	Path        string
	IsMatching  bool
	Key         string
	MatchingKey map[string]string
}

func NewRouter() *tree {
	r := new(tree)
	r.Router = make(map[string]*router)
	r.Map = make(map[string]HandlerFunc)
	return r

}
func (t *tree) Get(path string, handler HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if len(chain) > 0 {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(chain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	t.append("GET", path, handler, newChain)
}

func (t *tree) Post(path string, handler HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if len(chain) > 0 {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(newChain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	t.append("POST", path, handler, newChain)
}

func (t *tree) Delete(path string, handler HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if len(chain) > 0 {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(newChain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	t.append("DELETE", path, handler, newChain)
}

func (t *tree) Head(path string, handler HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if len(chain) > 0 {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(newChain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	t.append("HEAD", path, handler, newChain)
}

func (t *tree) Options(path string, handler HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if len(chain) > 0 {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(newChain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	t.append("OPTIONS", path, handler, newChain)
}

func (t *tree) Put(path string, handler HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if len(chain) > 0 {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(newChain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	t.append("PUT", path, handler, newChain)
}

func (t *tree) Patch(path string, handler HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if len(chain) > 0 {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(newChain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	t.append("PATCH", path, handler, newChain)
}

func (t *tree) Trace(path string, handler HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if len(chain) > 0 {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(newChain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	t.append("TRACE", path, handler, newChain)
}

func (t *tree) Connect(path string, handler HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if len(chain) > 0 {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(newChain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	t.append("CONNECT", path, handler, newChain)
}

func (gup *Groups) Get(path string, handlerFunc HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if (gup.Middleware != nil && len(gup.Middleware) > 0) || (chain != nil && len(chain) > 0) {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(gup.Middleware); i++ {
			newChain = append(newChain, gup.Middleware[i]...)
		}
		for i := 0; i < len(chain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	gup.Tree.Get(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc, newChain)
}

func (gup *Groups) Head(path string, handlerFunc HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if (gup.Middleware != nil && len(gup.Middleware) > 0) || (chain != nil && len(chain) > 0) {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(gup.Middleware); i++ {
			newChain = append(newChain, gup.Middleware[i]...)
		}
		for i := 0; i < len(chain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	gup.Tree.Head(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc, newChain)
}

func (gup *Groups) Post(path string, handlerFunc HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if (gup.Middleware != nil && len(gup.Middleware) > 0) || (chain != nil && len(chain) > 0) {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(gup.Middleware); i++ {
			newChain = append(newChain, gup.Middleware[i]...)
		}
		for i := 0; i < len(chain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	gup.Tree.Post(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc, newChain)
}

func (gup *Groups) Put(path string, handlerFunc HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if (gup.Middleware != nil && len(gup.Middleware) > 0) || (chain != nil && len(chain) > 0) {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(gup.Middleware); i++ {
			newChain = append(newChain, gup.Middleware[i]...)
		}
		for i := 0; i < len(chain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	gup.Tree.Put(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc, newChain)
}

func (gup *Groups) Patch(path string, handlerFunc HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if (gup.Middleware != nil && len(gup.Middleware) > 0) || (chain != nil && len(chain) > 0) {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(gup.Middleware); i++ {
			newChain = append(newChain, gup.Middleware[i]...)
		}
		for i := 0; i < len(chain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	gup.Tree.Patch(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc, newChain)
}

func (gup *Groups) Delete(path string, handlerFunc HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if (gup.Middleware != nil && len(gup.Middleware) > 0) || (chain != nil && len(chain) > 0) {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(gup.Middleware); i++ {
			newChain = append(newChain, gup.Middleware[i]...)
		}
		for i := 0; i < len(chain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	gup.Tree.Delete(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc, newChain)
}

func (gup *Groups) Connect(path string, handlerFunc HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if (gup.Middleware != nil && len(gup.Middleware) > 0) || (chain != nil && len(chain) > 0) {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(gup.Middleware); i++ {
			newChain = append(newChain, gup.Middleware[i]...)
		}
		for i := 0; i < len(chain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	gup.Tree.Connect(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc, newChain)
}

func (gup *Groups) Trace(path string, handlerFunc HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if (gup.Middleware != nil && len(gup.Middleware) > 0) || (chain != nil && len(chain) > 0) {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(gup.Middleware); i++ {
			newChain = append(newChain, gup.Middleware[i]...)
		}
		for i := 0; i < len(chain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	gup.Tree.Trace(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc, newChain)
}

func (gup *Groups) Options(path string, handlerFunc HandlerFunc, chain ...MiddlewareChain) {
	var newChain MiddlewareChain
	if (gup.Middleware != nil && len(gup.Middleware) > 0) || (chain != nil && len(chain) > 0) {
		newChain = make(MiddlewareChain, 0)
		for i := 0; i < len(gup.Middleware); i++ {
			newChain = append(newChain, gup.Middleware[i]...)
		}
		for i := 0; i < len(chain); i++ {
			newChain = append(newChain, chain[i]...)
		}
	}
	gup.Tree.Options(gup.Path+"/"+strings.TrimPrefix(path, "/"), handlerFunc, newChain)
}

func (gup *Groups) Group(path string, fn func(Groups *Groups), chain ...MiddlewareChain) {
	g := new(Groups)
	g.Path = gup.Path + "/" + strings.TrimPrefix(path, "/")
	g.Tree = gup.Tree
	g.Middleware = chain
	fn(g)
}

func (t *tree) Group(path string, fn func(groups *Groups), chain ...MiddlewareChain) {
	g := new(Groups)
	g.Tree = t
	g.Path = path
	g.Middleware = chain
	fn(g)
}
