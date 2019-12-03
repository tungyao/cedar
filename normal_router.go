package cedar

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var FileType = map[string]string{"css": "text/css", "txt": "text/plain", "zip": "application/x-zip-compressed", "png": "image/png", "jpg": "image/jpeg"}

type Groups struct {
	tree *Trie
	path string
}

func writeStaticFile(path string, filename []string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", FileType[filename[1]])
	w.Header().Set("Charset", "UTF-8")
	fs, err := os.OpenFile("."+path, os.O_RDONLY, 0666)
	if err != nil {
		log.Println(err)
	}
	data, err := ioutil.ReadAll(fs)
	if err != nil {
		log.Println(err)
	}
	if pusher, ok := w.(http.Pusher); ok {
		// Push is supported.
		//options := &http.PushOptions{
		//	Header: http.Header{
		//		"Accept-Encoding": ,
		//	},
		//}
		if err := pusher.Push("/app.js", nil); err != nil {
			_, err = w.Write(data)
		}
	}
	if err != nil {
		log.Println(err)
	}
}
func (mux *Trie) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//reg := regexp.MustCompile(`^/static[/\w-]*\.\w+$`)
	//file := reg.FindStringSubmatch(r.URL.String())
	if len(r.URL.Path) < 7 {
		goto bacall
	}
	if r.URL.Path[1:7] == "static" {
		filename := SplitString([]byte(r.URL.Path[9:]), []byte("."))
		writeStaticFile(r.URL.Path, filename, w)
		return
	}
	//if len(file) != 0 {
	//	filename := strings.Split(file[0], ".")
	//	writeStaticFile(r.URL.Path, filename, w)
	//	return
	//}
bacall:
	me, fun := mux.Find(r.URL.Path)
	if fun == nil || r.Method != me {
		w.Header().Set("Content-type", "text/html")
		w.Header().Set("charset", "UTF-8")
		w.WriteHeader(404)
		_, _ = w.Write([]byte("<p style=\"font-size=500px\">404</p>"))
		return
	}
	if fun != nil {
		fun(w, r)
	}
}
func (mux *Trie) Group(path string, fn func(groups *Groups)) {
	g := new(Groups)
	g.tree = mux
	g.path = path
	fn(g)
}
func (mux *Groups) Get(path string, fun http.HandlerFunc) {
	mux.tree.Get(mux.path+path, fun)
}
func (mux *Groups) Post(path string, fun http.HandlerFunc) {
	mux.tree.Post(mux.path+path, fun)
}
func (mux *Groups) Put(path string, fun http.HandlerFunc) {
	mux.tree.Put(mux.path+path, fun)
}
func (mux *Groups) Delete(path string, fun http.HandlerFunc) {
	mux.tree.Delete(mux.path+path, fun)
}

func (mux *Trie) Get(path string, fun http.HandlerFunc) {
	mux.Insert(http.MethodGet, path, fun)
}
func (mux *Trie) Post(path string, fun http.HandlerFunc) {
	mux.Insert(http.MethodPost, path, fun)
}
func (mux *Trie) Put(path string, fun http.HandlerFunc) {
	mux.Insert(http.MethodPut, path, fun)
}
func (mux *Trie) Delete(path string, fun http.HandlerFunc) {
	mux.Insert(http.MethodDelete, path, fun)
}
func (mux *Trie) Static(filepath string) {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath))))
}
func (mux *Trie) Listening(parameter ...interface{}) error {
	if len(parameter) != 2 && len(parameter) != 4 {
		return errors.New("parameter length must is 2 or 4")
	}
	if len(parameter) == 2 {
		if ok := http.ListenAndServe(parameter[0].(string), parameter[1].(http.Handler)); ok != nil {
			return ok
		}
	} else {
		if ok := http.ListenAndServeTLS(parameter[0].(string), parameter[1].(string), parameter[2].(string), parameter[3].(http.Handler)); ok != nil {
			return ok
		}
	}
	return nil
}
