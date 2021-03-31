package cedar

import (
	"bufio"
	"context"
	"fmt"
	json2 "github.com/json-iterator/go"
	"html/template"
	"io"
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
	Name      string `dynamic:"name"`
	Path      string `dynamic:"path"`
	View      string `dynamic:"view"`
	Method    string `dynamic:"method"`
	Type      string `dynamic:"type"`
	ProxyPass string `dynamic:"proxy_pass"`
}

const (
	TypeJson = "application/json"
)

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
	_, hand, handler, middle, p, ok := mux.Find(r.URL.Path, r.Method)
	if !ok {
		w.Header().Set("Content-type", "text/html")
		w.Header().Set("charset", "UTF-8")
		w.WriteHeader(404)
		_, _ = w.Write([]byte("404"))
		return
	}
	if handler != nil {
		handler.ServeHTTP(w, r)
		return
	}
	r.URL.Fragment = p
	name := ""
	if mux.sessions != nil {
		c, err := r.Cookie("session")
		if err == http.ErrNoCookie {
			x := Sha1(mux.sessions.CreateUUID([]byte(r.RemoteAddr)))
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    string(x),
				HttpOnly: true, Secure: false, Path: "/",
				Expires: time.Now().Add(1 * time.Hour), Domain: mux.sessions.Domino,
			})
			name = string(x)
		} else {
			if c != nil {
				name = c.Value
			}
		}
	}
	co := &Core{
		writer: w,
		resp:   r,
		Session: Session{
			Cookie: name,
		},
	}
	go func() {
		for k, v := range mux.globalFunc {
			if err := v.Fn(w, r, co); err != nil {
				log.Panicln(k, err)
			}
		}
	}()
	if m := mux.middle[middle]; m != nil {
		if !m(w, r, co) {
			return
		}
	}
	for _, v := range autosome {
		v.MethodByName("AutoBefore").Call([]reflect.Value{reflect.ValueOf(w),
			reflect.ValueOf(r),
			reflect.ValueOf(co)})
	}
	if hand != nil {
		hand(w, r, co)
	}

}

// handFunc
func (mux *Groups) Get(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.Tree.Get(mux.Path+path, handlerFunc, middleName...)
}
func (mux *Groups) Head(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.Tree.Head(mux.Path+path, handlerFunc, middleName...)
}
func (mux *Groups) Post(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.Tree.Post(mux.Path+path, handlerFunc, middleName...)
}
func (mux *Groups) Put(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.Tree.Put(mux.Path+path, handlerFunc, middleName...)
}
func (mux *Groups) Patch(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.Tree.Patch(mux.Path+path, handlerFunc, middleName...)
}
func (mux *Groups) Delete(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.Tree.Delete(mux.Path+path, handlerFunc, middleName...)
}
func (mux *Groups) Connect(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.Tree.Connect(mux.Path+path, handlerFunc, middleName...)
}
func (mux *Groups) Trace(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.Tree.Trace(mux.Path+path, handlerFunc, middleName...)
}
func (mux *Groups) Options(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.Tree.Options(mux.Path+path, handlerFunc, middleName...)
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

func filter(s []byte) string {
	sr := make([]byte, 0)
	// 清除前面的空格符号
	for k, v := range s {
		if v != ' ' {
			s = s[k:]
			break
		}
	}
	// 清除后面的换行 符号
	for _, v := range s {
		if v != '\n' && v != '\r' {
			sr = append(sr, v)
		}
	}
	return string(sr)
}
func (mux *Trie) Get(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.insert(http.MethodGet, path, handlerFunc, nil, middleName)
}
func (mux *Trie) Head(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.insert(http.MethodHead, path, handlerFunc, nil, middleName)
}
func (mux *Trie) Post(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.insert(http.MethodPost, path, handlerFunc, nil, middleName)
}
func (mux *Trie) Put(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.insert(http.MethodPut, path, handlerFunc, nil, middleName)
}
func (mux *Trie) Patch(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.insert(http.MethodPatch, path, handlerFunc, nil, middleName)
}
func (mux *Trie) Delete(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.insert(http.MethodDelete, path, handlerFunc, nil, middleName)
}
func (mux *Trie) Connect(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.insert(http.MethodConnect, path, handlerFunc, nil, middleName)
}
func (mux *Trie) Trace(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.insert(http.MethodTrace, path, handlerFunc, nil, middleName)
}
func (mux *Trie) Options(path string, handlerFunc HandlerFunc, middleName ...string) {
	mux.insert(http.MethodOptions, path, handlerFunc, nil, middleName)
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
func (mux *Trie) Dynamic(ymlPath string) {
	f, err := os.Open(ymlPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	point := false
	var dy = make(map[string]*DynamicRoute, 0)
	var single = &DynamicRoute{}
	// 得弄个排序出来
	var sortTag = map[string]int{}
	typ := reflect.TypeOf(single)
	v := reflect.ValueOf(single).Elem()
	for i := 0; i < v.NumField(); i++ {
		tag := typ.Elem().Field(i).Tag.Get("dynamic")
		sortTag[tag] = i
	}
	for {
		line, err := rd.ReadBytes('\n')
		if err != nil || len(line) == 0 {
			break
		}
		if line[0] == '#' {
			continue
		}
		if point {
			if line[0] == '-' {
				// fmt.Println(single)
				point = false
				goto ct
			}
			kp := 0
			for k, v := range line {
				if v == ':' {
					kp = k
					break
				}
			}
			reflect.ValueOf(single).Elem().Field(sortTag[string(line[2:kp])]).SetString(filter(line[kp+1:]))
			if single.Name != "" {
				dy[single.Name] = single
			}
			continue
		}
		goto ct
	ct:
		if line[0] == '-' {
			point = true // 开始计算
			single = &DynamicRoute{}
			// 解析开始
			kp := 0
			for k, v := range line {
				if v == ':' {
					kp = k
					break
				}
			}
			reflect.ValueOf(single).Elem().Field(sortTag[string(line[2:kp])]).SetString(filter(line[kp+1:]))
		}
	}
	// if route != nil {
	//
	// }
	for _, v := range dy {
		go func(v *DynamicRoute) {
			switch strings.ToUpper(v.Method) {
			case "GET":
				mux.Get(v.Path, mux.HttpProxy(v))
			case "POST":
				mux.Post(v.Path, mux.HttpProxy(v))
			case "PUT":
				mux.Put(v.Path, mux.HttpProxy(v))
			case "DELETE":
				mux.Delete(v.Path, mux.HttpProxy(v))
			case "CONNECT":
				mux.Connect(v.Path, mux.HttpProxy(v))
			case "OPTIONS":
				mux.Options(v.Path, mux.HttpProxy(v))
			case "TRACE":
				mux.Trace(v.Path, mux.HttpProxy(v))
			}
		}(v)

	}
}
func (mux *Trie) HttpProxy(dr *DynamicRoute) func(w http.ResponseWriter, r *http.Request, core *Core) {
	switch dr.Type {
	case "proxy":
		return func(w http.ResponseWriter, r *http.Request, core *Core) {
			client := &http.Client{}
			req, err := http.NewRequest(dr.Method, dr.ProxyPass, r.Body)
			if err != nil {
				log.Panicln(err)
			}
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
			io.Copy(w, resp.Body)
		}
	case "static":
		return func(w http.ResponseWriter, r *http.Request, core *Core) {
			tp, err := template.ParseFiles(includeTemp(dr.View, Layout)...)
			if err != nil {
				log.Println(err)
				return
			}
			tp.Execute(w, nil)
		}
	}
	return nil
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
			thrownErr(err, v.w)
			return
		}
		err = tp.Execute(v.w, v.data)
		thrownErr(err, v.w)
		return
	}
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	s := getTemplatePath(runtime.FuncForPC(pc[0]).Name())
	tp, err := template.ParseFiles(includeTemp(s, Layout)...)
	if err != nil {
		thrownErr(err, v.w)
		return
	}
	err = tp.Execute(v.w, v.data)
	thrownErr(err, v.w)
}

func (co *Core) View() *view {
	return &view{
		data: make(map[string]interface{}),
		w:    co.writer,
	}
}

// data
// http status 1
// map or struct 2
// header 3
func (co *Core) Json(data ...interface{}) *json {
	if len(data) > 1 {
		co.writer.WriteHeader(data[0].(int))
		b, err := json2.Marshal(data[1])
		if err != nil {
			thrownErr(err, co.writer)
			return nil
		}
		co.writer.Header().Set("Content-Type", TypeJson)
		co.writer.Write(b)
	}
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

// 2020/8/31
// add basic auth component
// arg doubles include name and pass like this
// BasicAuth("user","pass","user2","pass2") , it's true
// BasicAuth("user","pass","user2") , it's false
func (co *Core) BasicAuth(args ...string) bool {
	if len(args)%2 != 0 {
		http.Error(co.writer, " args failed", 503)
		return true
	}
	user, pass, ok := co.resp.BasicAuth()
	if !ok {
		co.writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		co.writer.WriteHeader(http.StatusUnauthorized)
		return true
	}
	check := false
	for i := 0; i < len(args)-2; i += 2 {
		if user != args[i] && pass != args[i+1] {
			check = true
		} else {
			check = false
			break
		}
	}
	if check {
		http.Error(co.writer, " need authorized!", http.StatusUnauthorized)
		return true
	}
	return false
}

func byt(s string) []byte {
	rs := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	return *(*[]byte)(unsafe.Pointer(&rs))
}
func str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
func thrownErr(err error, w http.ResponseWriter) {
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
	i := 0
	if s[p:4] == "Page" {
		i = 4
		p = 4
	}
	for k, v := range s[i:] {
		if v > 64 && v < 91 && p != k+i {
			o += strings.ToLower(s[p:k+i]) + "/"
			p = k + i
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
		thrownErr(err, j.w)
		return
	}
	j.w.Header().Set("content-type", "application/json")
	j.w.Write(b)
}
func (j *json) Error(err string) {
	j.w.WriteHeader(403)
	j.w.Header().Set("content-type", "application/json")
	j.w.Write(byt(`{"msg":"` + err + `"}`))
}

// 自动注册路由
type AutoRegister struct {
	ctx context.Context
}

func (mux *Trie) AutoRegister(auto interface{}) *AutoRegister {
	spPkg := strings.Split(reflect.TypeOf(auto).Elem().PkgPath(), "/")
	pkgName := spPkg[len(spPkg)-1] + "/"
	for i := 0; i < reflect.ValueOf(auto).NumMethod(); i++ {
		mName := reflect.TypeOf(auto).Method(i).Name
		fuc := reflect.ValueOf(auto).MethodByName(mName)
		// 判断是不是中间件构成
		if mName[:6] == "Middle" {
			reflect.ValueOf(mux).MethodByName("Middleware").Call([]reflect.Value{reflect.ValueOf(strings.ToLower(mName[6:])), reflect.ValueOf(fuc.Interface().(func(w http.ResponseWriter, r *http.Request, co *Core) bool))})
			continue
		}
		x := fuc.Interface().(func(writer http.ResponseWriter, request *http.Request, core *Core))
		// x := *(*func(writer http.ResponseWriter, request *http.Request, core *Core))(unsafe.Pointer(fuc.Pointer()))
		ma := strings.Split(mName, "_")
		if len(ma) == 2 {
			mName = ma[1]
			mName = getRouterPath(mName)
		} else if len(ma) > 2 {
			mName = ma[len(ma)-1]
			if mName == "" {
				mName = ma[len(ma)-2]
			}
			mName = getRouterPath(mName)
		}
		// in := make([]reflect.Value, 0)
		// in = append(in, reflect.ValueOf("/"+pkgName+getRouterPath(mName)))
		// in = append(in, reflect.ValueOf(mName))
		// in = append(in, reflect.ValueOf(x))
		// in = append(in, reflect.ValueOf(reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())))
		in := make([]string, 0)
		for i := 0; i < len(ma)-2; i++ {
			in = append(in, strings.ToLower(ma[1+i]))
		}
		// 在这里可能会出现意外卡住 【已解决】
		go func() {
			switch ma[0] {
			case "Get":
				mux.Get("/"+pkgName+getRouterPath(mName), x, in...)
			case "Post":
				mux.Post("/"+pkgName+getRouterPath(mName), x, in...)
			case "Put":
				mux.Put("/"+pkgName+getRouterPath(mName), x, in...)
			case "Delete":
				mux.Delete("/"+pkgName+getRouterPath(mName), x, in...)
			case "Connect":
				mux.Connect("/"+pkgName+getRouterPath(mName), x, in...)
			case "Options":
				mux.Options("/"+pkgName+getRouterPath(mName), x, in...)
			case "Trace":
				mux.Trace("/"+pkgName+getRouterPath(mName), x, in...)
			}
		}()
		// go reflect.ValueOf(mux).MethodByName(ma[0]).Call(in)
	}
	t := &AutoRegister{}
	return t
}

var (
	pluginArr = make(map[string]reflect.Value)
	autosome  = make([]reflect.Value, 0)
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
	autosome = append(autosome, x[0])
	pluginArr[st.Elem().Name()] = rv

}

// func (pl *Plugin)AutoStart() interface{}  {
// 	return nil
// }
func (pl *Plugin) AutoBefore(http.ResponseWriter, *http.Request, *Core) {}
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
