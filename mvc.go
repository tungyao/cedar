package cedar

import (
	json2 "encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
	"unsafe"
)

var FileType = map[string]string{"html": "text/html", "json": "application/json", "css": "text/css", "txt": "text/plain", "zip": "application/x-zip-compressed", "png": "image/png", "jpg": "image/jpeg"}

type Groups struct {
	Tree *Trie
	Path string
}
type DynamicRoute struct {
	Path string
	View string
}

func writeStaticFile(path string, filename []string, w http.ResponseWriter) {
	if pusher, ok := w.(http.Pusher); ok {
		// Push is supported.
		options := &http.PushOptions{
			Header: http.Header{
				"Accept-Encoding": {"Content-Type:" + FileType[filename[1]]},
			},
		}
		if err := pusher.Push("."+path, options); err != nil {
			goto end
		}
	} else {
		goto end
	}
end:
	w.Header().Set("Content-Type", FileType[filename[1]])
	fs, err := os.OpenFile("."+path, os.O_RDONLY, 0666)
	if err != nil {
		log.Println(err)
	}
	defer fs.Close()
	data, err := ioutil.ReadAll(fs)
	if err != nil {
		log.Println(err)
	}
	_, err = w.Write(data)
}
func (mux *Trie) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) > 7 && r.URL.Path[1:7] == "static" {
		filename := SplitString([]byte(r.URL.Path[8:]), []byte("."))
		writeStaticFile(r.URL.Path, filename, w)
		return
	}

	me, handf, _, midle, p := mux.Find(r.URL.Path + "/" + r.Method)
	r.URL.Fragment = p
	if r.Method != me {
		w.Header().Set("Content-type", "text/html")
		w.Header().Set("charset", "UTF-8")
		w.WriteHeader(404)
		_, _ = w.Write([]byte("<h1>404</h1>"))
		return
	}
	sname := ""
	if mux.sessionx != nil {
		c, err := r.Cookie("session")
		if err == http.ErrNoCookie {
			x := Sha1(mux.sessionx.CreateUUID([]byte(r.RemoteAddr)))
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    string(x),
				HttpOnly: true, Secure: false, Path: "/",
				Expires: time.Now().Add(1 * time.Hour), Domain: mux.sessionx.Domino,
			})
			sname = string(x)
		} else {
			sname = c.Value
		}
	}
	co := &Core{
		writer: w,
		resp:   r,
		Session: Session{
			Cookie: sname,
		},
	}
	go func() {
		for k, v := range mux.globalFunc {
			if err := v.Fn(w, r, co); err != nil {
				log.Panicln(k, err)
			}
		}
	}()
	if m := mux.middle[midle]; m != nil {
		if !m(w, r, co) {
			return
		}
	}
	// if hand != nil {
	// 	hand.ServeHTTP(w, r)
	// }

	for _, v := range autobefore {
		v.MethodByName("AutoBefore").Call([]reflect.Value{reflect.ValueOf(w),
			reflect.ValueOf(r),
			reflect.ValueOf(co)})
	}
	if handf != nil {
		handf(w, r, co)
	}

}

func (mux *Groups) Get(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Get(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) HEAD(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Head(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) Post(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Post(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) Put(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Put(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) PATCH(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Patch(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) Delete(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Delete(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) CONNECT(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Connect(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) TRACE(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Trace(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) OPTIONS(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Options(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) Middleware(name string, fn func(w http.ResponseWriter, r *http.Request, co *Core) bool) {
	mux.Tree.Middleware(name, fn)
}
func (mux *Groups) Group(path string, fn func(groups *Groups)) {
	g := new(Groups)
	g.Path = mux.Path + path
	g.Tree = mux.Tree
	fn(g)
}

func (mux *Trie) Dynamic(ymlPath string) {
	defer func() {
		x := recover()
		if x != nil {
			log.Panic(x)
		}
	}()
	fs, err := os.Open(ymlPath)
	if err != nil {
		log.Panic(91, err)
	}
	defer fs.Close()
	all, _ := ioutil.ReadAll(fs)
	var enterStyle = 0
	var lastChar = 0
	var dy = make(map[string]*DynamicRoute, 0)
	for k, v := range all {
		if v == 13 {
			if all[k+1] == 10 {
				enterStyle = 0 // windows
			} else {
				enterStyle = 1 // unix
			}
			break
		}
	}
	var path = ""
	for k, v := range all {
		if v == 13 {
			if enterStyle == 0 {
				for j, c := range all[lastChar:k] {
					if c == 58 {
						if path == "" {
							path = string(all[lastChar:k][j+2:])
							break
						} else if path != "" {
							dy[path] = &DynamicRoute{
								path,
								string(all[lastChar:k][j+2:]),
							}
							path = ""
							break
						}
					}
				}
				lastChar = k + 2
				continue
			} else {
				lastChar = k
			}
		}
	}
	dy[path] = &DynamicRoute{
		path,
		string(all[lastChar+8 : len(all)]),
	}
	for _, v := range dy {
		mux.Get(v.Path, func(writer http.ResponseWriter, request *http.Request, r *Core) {
			writeStaticFile(dy[request.URL.Path].View, []string{"", "html"}, writer)
		}, nil)
	}
}
func (mux *Trie) Get(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodGet, path+"/GET", handlerFunc, handler, middleName)
}
func (mux *Trie) Head(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodHead, path+"/HEAD", handlerFunc, handler, middleName)
}
func (mux *Trie) Post(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodPost, path+"/POST", handlerFunc, handler, middleName)
}
func (mux *Trie) Put(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodPut, path+"/PUT", handlerFunc, handler, middleName)
}
func (mux *Trie) Patch(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodPatch, path+"/PATCH", handlerFunc, handler, middleName)
}
func (mux *Trie) Delete(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodDelete, path+"/DELETE", handlerFunc, handler, middleName)
}
func (mux *Trie) Connect(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodConnect, path+"/CONNECT", handlerFunc, handler, middleName)
}
func (mux *Trie) Trace(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodTrace, path+"/TRACE", handlerFunc, handler, middleName)
}
func (mux *Trie) Options(path string, handlerFunc HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodOptions, path+"/OPTIONS", handlerFunc, handler, middleName)
}
func (mux *Trie) GlobalFunc(name string, fn func(w http.ResponseWriter, r *http.Request, co *Core) error) {
	mux.globalFunc = append(mux.globalFunc, &GlobalFunc{
		Name: name,
		Fn:   fn,
	})
}
func (mux *Trie) Middleware(name string, fn func(w http.ResponseWriter, r *http.Request, co *Core) bool) {
	fmt.Println("middle =>", name)
	mux.Middle(name, fn)
}
func (mux *Trie) Group(path string, fn func(groups *Groups)) {
	g := new(Groups)
	g.Tree = mux
	g.Path = path
	fn(g)
}

type Core struct {
	writer  http.ResponseWriter
	resp    *http.Request
	PL      *Plugin
	Session Session
}

// add new component view render
var (
	Debug  = false
	Layout []string
	OUT    = "./view"
)

type view struct {
	data map[string]interface{}
	w    http.ResponseWriter
}
type json struct {
	w http.ResponseWriter
}

func (v *view) Assign(name string, value interface{}) *view {
	v.data[name] = value
	return v
}
func (v *view) Render(view_page ...string) {
	if len(view_page) > 0 {
		tp, err := template.ParseFiles(includeTemp(OUT+"/"+view_page[0]+".html", Layout)...)
		if err != nil {
			thownErr(err, v.w)
			return
		}
		err = tp.Execute(v.w, v.data)
		thownErr(err, v.w)
		return
	}
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	s := getTemplatePath(runtime.FuncForPC(pc[0]).Name())
	tp, err := template.ParseFiles(includeTemp(s, Layout)...)
	if err != nil {
		thownErr(err, v.w)
		return
	}
	err = tp.Execute(v.w, v.data)
	thownErr(err, v.w)
}

func (co *Core) View() *view {
	return &view{
		data: make(map[string]interface{}),
		w:    co.writer,
	}
}
func (co *Core) Json() *json {
	return &json{
		w: co.writer,
	}
}
func (co *Core) Byte(s string) []byte {
	rs := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	return *(*[]byte)(unsafe.Pointer(&rs))
}
func (co *Core) String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
func byt(s string) []byte {
	rs := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	return *(*[]byte)(unsafe.Pointer(&rs))
}
func str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
func thownErr(err error, w http.ResponseWriter) {
	if err == nil {
		return
	}
	if Debug {
		w.Write(byt(err.Error()))
	} else {
		http.NotFound(w, nil)
		// w.WriteHeader(404)
		// w.Write(byt("404 NOT FOUND"))
	}
}

// 提取模板路径
func getTemplatePath(s string) string {
	sr := strings.Split(s, ".")
	p := 0
	o := OUT
	if strings.Index(sr[len(sr)-1], "_") != -1 {
		sx := strings.Split(sr[len(sr)-1], "_")
		sr[1] = sx[len(sx)-1]
		if sr[1] == "" {
			sr[1] = sx[len(sx)-2] + "/index"
		}
	}
	for k, v := range sr[1][1:] {
		if v > 64 && v < 91 {
			if p == 0 {
				if sr[1][p:k+1] != "Page" {
					return ""
				}
				p = k + 1
			}
			o += strings.ToLower(sr[1][p:k+1]) + "/"
			p = k + 1
		}
	}
	o += strings.ToLower(sr[1][p:]) + ".html"
	return o
}
func getRouterPath(s string) string {
	p := 0
	o := ""
	for k, v := range s[1:] {
		if v > 64 && v < 91 {
			if p == 0 {
				if s[p:k+1] != "Page" {
					return ""
				}
				p = k + 1
			}
			o += strings.ToLower(s[p:k+1]) + "/"
			p = k + 1
		}
	}
	o += strings.ToLower(s[p:])
	return o
}
func (mux *Trie) SetDebug() {
	Debug = true
}
func (mux *Trie) SetLayout(path ...string) {
	Layout = path
}
func (mux *Trie) SetView(path string) {
	OUT = path
}
func includeTemp(s string, ss []string) []string {
	if len(ss) == 0 {
		return []string{s}
	}
	n := make([]string, 0)
	n = append(n, s)
	n = append(n, ss...)
	return n
}

func (j *json) Success(data interface{}) {
	b, err := json2.Marshal(data)
	if err != nil {
		thownErr(err, j.w)
		return
	}
	j.w.Header().Set("content-type", "application/json")
	j.w.Write(b)
}
func (j *json) Error(err string) {
	j.w.WriteHeader(503)
	j.w.Header().Set("content-type", "application/json")
	j.w.Write(byt(`{"msg":"` + err + `"}`))
}

// 自动注册路由
type AutoRegister struct {
}

func (mux *Trie) AutoRegister(auto interface{}, middleware ...string) *AutoRegister {
	// spPkg := strings.Split(reflect.TypeOf(auto).Elem().PkgPath(), "/")
	// pkgName := spPkg[len(spPkg)-1]
	for i := 0; i < reflect.ValueOf(auto).NumMethod(); i++ {
		mName := reflect.TypeOf(auto).Method(i).Name
		fuc := reflect.ValueOf(auto).MethodByName(mName)
		// 判断是不是中间件构成
		if mName[:6] == "Middle" {
			reflect.ValueOf(mux).MethodByName("Middleware").Call([]reflect.Value{reflect.ValueOf(strings.ToLower(mName[6:])), reflect.ValueOf(fuc.Interface().(func(w http.ResponseWriter, r *http.Request) bool))})
			continue
		}
		x := fuc.Interface().(func(writer http.ResponseWriter, request *http.Request, core *Core))
		// x := *(*func(writer http.ResponseWriter, request *http.Request, core *Core))(unsafe.Pointer(fuc.Pointer()))
		ma := strings.Split(mName, "_")
		if len(ma) == 2 {
			mName = ma[1]
		} else if len(ma) > 2 {
			mName = ma[len(ma)-1]
			if mName == "" {
				mName = ma[len(ma)-2]
			}
		}
		in := make([]reflect.Value, 0)
		// in = append(in, reflect.ValueOf("/"+pkgName+getRouterPath(mName)))
		in = append(in, reflect.ValueOf(getRouterPath(mName)))
		in = append(in, reflect.ValueOf(x))
		in = append(in, reflect.ValueOf(http.Handler(mux)))
		for i := 0; i < len(ma)-2; i++ {
			in = append(in, reflect.ValueOf(strings.ToLower(ma[1+i])))
		}
		reflect.ValueOf(mux).MethodByName(ma[0]).Call(in)
	}
	t := &AutoRegister{}
	return t
}

var (
	pluginArr  = make(map[string]reflect.Value)
	autobefore = make([]reflect.Value, 0)
)

// 要求有自动注册插件到Core中去
// 插件名字 插件结构体
// 插件应该要继承 Core 结构体 并且重写 AutoStart 和 AutoBefore 方法
// 插件执行时间 应该分为2个阶段 一个是 系统启动 一个是 执行HandlerFunc之前
type Plugin struct {
}
type plugin struct {
	structs reflect.Value
}

// 直接从这里加载进插件池
func (mux *Trie) Plugin(pluginStruct interface{}) {
	st := reflect.TypeOf(pluginStruct)
	rv := reflect.ValueOf(pluginStruct)
	x := rv.MethodByName("AutoStart").Call(nil)
	autobefore = append(autobefore, x[0])
	pluginArr[st.Elem().Name()] = rv

}

// func (pl *Plugin)AutoStart() interface{}  {
// 	return nil
// }
func (pl *Plugin) AutoBefore(w http.ResponseWriter, r *http.Request, co *Core) {

}
func (pl plugin) Call(funcName string, args ...interface{}) []reflect.Value {
	arr := make([]reflect.Value, len(args))
	if len(args) != 0 {
		for k, v := range args {
			arr[k] = reflect.ValueOf(v)
		}
	}
	return pl.structs.MethodByName(funcName).Call(arr)
}

// 在方法之中使用 插件
func (co *Core) Plugin(name string) plugin {
	return plugin{
		structs: pluginArr[name],
	}
}
