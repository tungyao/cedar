package cedar

import (
	"fmt"
	"net/http"
)

type RestConfig struct {
	EntryPath string
	ApiName   string
}
type _rest struct {
	trie   Trie
	config RestConfig
}

func NewRestRouter(rc RestConfig) *_rest {
	return &_rest{trie: Trie{
		num: 1,
		root: NewSon("GET", rc.EntryPath+"?"+rc.ApiName+"=", func(writer http.ResponseWriter, request *http.Request) {
			_, _ = fmt.Fprint(writer, "index")
		}, 1),
	}, config: rc,
	}
}
func (re *_rest) GetR(api string, fn http.HandlerFunc) {
	re.trie.Insert("GET", api, fn)
}
func (re *_rest) PostR(api string, fn http.HandlerFunc) {
	re.trie.Insert("Post", api, fn)
}
func (re *_rest) PutR(api string, fn http.HandlerFunc) {
	re.trie.Insert("Put", api, fn)
}
func (re *_rest) DeleteR(api string, fn http.HandlerFunc) {
	re.trie.Insert("Delete", api, fn)
}
func (re *_rest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	me, fun := re.trie.Find(r.URL.Query().Get(re.config.ApiName))
	if fun == nil || r.Method != me || r.URL.Path != "/"+re.config.EntryPath {
		w.Header().Set("Content-type", "text/html")
		w.WriteHeader(404)
		_, _ = w.Write([]byte("<p style=\"font-size=500px\">404</p>"))
		return
	}
	if fun != nil {
		fun(w, r)
	}
}
