package ultimate_cedar

import (
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	json "github.com/json-iterator/go"
)

// 在想能不能借助数组来存放路由
// /ab/bc
// /ab/bc/:id
// /ab/bc/:id/nv
// 只需要做到这个三种匹配就能完成绝大部分的路由匹配

// Handler 对原来的方法进行重写
type Handler func(ResponseWriter, Request)
type ResponseWriter struct {
	http.ResponseWriter
	*Json
}
type Request struct {
	*http.Request
	*en
	Data *Data
}
type Data struct {
	m map[string]string
}

func (d *Data) Get(key string) string {
	if v, ok := d.m[key]; ok {
		return v
	}
	return ""
}
func (d *Data) set(key, value string) {
	d.m[key] = value
}

type en struct {
	r *http.Request
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

type Json struct {
	writer http.ResponseWriter
	header map[string]string
	status int
	data   []byte
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
		j.writer.Header().Add(k, v)
	}
	j.writer.WriteHeader(j.status)
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
