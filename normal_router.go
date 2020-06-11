package cedar

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	go func() {
		for k, v := range mux.globalFunc {
			if err := v.Fn(w, r); err != nil {
				log.Panicln(k, err)
			}
		}
	}()
	me, handf, hand, midle, p := mux.Find(r.URL.Path + "/" + r.Method)
	r.URL.Fragment = p
	if r.Method != me {
		w.Header().Set("Content-type", "text/html")
		w.Header().Set("charset", "UTF-8")
		w.WriteHeader(404)
		_, _ = w.Write([]byte("<h1>404</h1>"))
		return
	}
	if m := mux.middle[midle]; m != nil {
		if !m(w, r) {
			return
		}
	}
	if hand != nil {
		hand.ServeHTTP(w, r)
	}
	if handf != nil {
		handf(w, r)
	}

}

func (mux *Groups) Get(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Get(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) HEAD(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Head(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) Post(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Post(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) Put(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Put(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) PATCH(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Patch(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) Delete(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Delete(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) CONNECT(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Connect(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) TRACE(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Trace(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) OPTIONS(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Tree.Options(mux.Path+path, handlerFunc, handler, middleName...)
}
func (mux *Groups) Middleware(name string, fn func(w http.ResponseWriter, r *http.Request) bool) {
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
	var dy map[string]*DynamicRoute = make(map[string]*DynamicRoute, 0)
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
	var path string = ""
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
		mux.Get(v.Path, func(writer http.ResponseWriter, request *http.Request) {
			mux.Template(writer, dy[request.URL.Path].View)
		}, nil)
	}

}
func (mux *Trie) Get(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodGet, path+"/GET", handlerFunc, handler, middleName)
}
func (mux *Trie) Head(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodHead, path+"/HEAD", handlerFunc, handler, middleName)
}
func (mux *Trie) Post(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodPost, path+"/POST", handlerFunc, handler, middleName)
}
func (mux *Trie) Put(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodPut, path+"/PUT", handlerFunc, handler, middleName)
}
func (mux *Trie) Patch(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodPatch, path+"/PATCH", handlerFunc, handler, middleName)
}
func (mux *Trie) Delete(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodDelete, path+"/DELETE", handlerFunc, handler, middleName)
}
func (mux *Trie) Connect(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodConnect, path+"/CONNECT", handlerFunc, handler, middleName)
}
func (mux *Trie) Trace(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodTrace, path+"/TRACE", handlerFunc, handler, middleName)
}
func (mux *Trie) Options(path string, handlerFunc http.HandlerFunc, handler http.Handler, middleName ...string) {
	mux.Insert(http.MethodOptions, path+"/OPTIONS", handlerFunc, handler, middleName)
}
func (mux *Trie) GlobalFunc(name string, fn func(w http.ResponseWriter, r *http.Request) error) {
	mux.globalFunc = append(mux.globalFunc, &GlobalFunc{
		Name: name,
		Fn:   fn,
	})
}
func (mux *Trie) Middleware(name string, fn func(w http.ResponseWriter, r *http.Request) bool) {
	mux.Middle(name, fn)
}
func (mux *Trie) Group(path string, fn func(groups *Groups)) {
	g := new(Groups)
	g.Tree = mux
	g.Path = path
	fn(g)
}
func (mux *Trie) Template(w http.ResponseWriter, path string) {
	writeStaticFile(path, []string{"", "html"}, w)
}
