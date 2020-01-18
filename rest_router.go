package cedar

import (
	"fmt"
	"net/http"
)

type RestConfig struct {
	EntryPath string
	ApiName   string
	Pattern   string
}
type _rest struct {
	trie   Trie
	config RestConfig
	static string
	index  string
}
type GroupR struct {
	tree *_rest
	path string
}

func NewRestRouter(rc RestConfig) *_rest {
	return &_rest{trie: Trie{
		num: 1,
		root: NewSon("GET", rc.EntryPath+"?"+rc.ApiName+"=", func(writer http.ResponseWriter, request *http.Request) {
			_, _ = fmt.Fprint(writer, "index")
		}, nil, 1),
		pattern: rc.Pattern,
	}, config: rc,
	}
}
func (re *_rest) Index(api string) {
	re.index = api
}
func (re *_rest) Get(api string, fn http.HandlerFunc, fnd http.Handler) {
	re.trie.Insert("GET", api, fn, fnd)
}
func (re *_rest) Post(api string, fn http.HandlerFunc, fnd http.Handler) {
	re.trie.Insert("POST", api, fn, fnd)
}
func (re *_rest) Put(api string, fn http.HandlerFunc, fnd http.Handler) {
	re.trie.Insert("PUT", api, fn, fnd)
}
func (re *_rest) Delete(api string, fn http.HandlerFunc, fnd http.Handler) {
	re.trie.Insert("DELETE", api, fn, fnd)
}
func (re *_rest) Group(path string, fn func(groups *GroupR)) {
	g := new(GroupR)
	g.tree = re
	g.path = path
	fn(g)
}
func (re *_rest) Template(w http.ResponseWriter, path string) {
	writeStaticFile(path+".html", []string{"", "html"}, w)
}
func (re *_rest) Static(filepath string) {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath))))
}
func (mux *GroupR) Get(path string, fun http.HandlerFunc, fnd http.Handler) {
	mux.tree.trie.Get(mux.path+mux.tree.config.Pattern+path, fun, fnd)
}
func (mux *GroupR) Post(path string, fun http.HandlerFunc, fnd http.Handler) {
	mux.tree.trie.Post(mux.path+mux.tree.config.Pattern+path, fun, fnd)
}
func (mux *GroupR) Put(path string, fun http.HandlerFunc, fnd http.Handler) {
	mux.tree.trie.Put(mux.path+mux.tree.config.Pattern+path, fun, fnd)
}
func (mux *GroupR) Delete(path string, fun http.HandlerFunc, fnd http.Handler) {
	mux.tree.trie.Delete(mux.path+mux.tree.config.Pattern+path, fun, fnd)
}
func (re *_rest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) > 7 && r.URL.Path[1:7] == "static" {
		filename := SplitString([]byte(r.URL.Path[8:]), []byte("."))
		writeStaticFile(r.URL.Path, filename, w)
		return
	}

	me, fun, _ := re.trie.Find(r.URL.Query().Get(re.config.ApiName))
	if r.URL.Path == "/" {
		me, fun, _ = re.trie.Find(re.index)
	}
	if fun == nil || r.Method != me {
		w.Header().Set("Content-type", "text/html")
		w.WriteHeader(404)
		_, _ = w.Write([]byte("<p style=\"font-size=500px\">404</p>"))
		return
	}
	fun(w, r)

}
