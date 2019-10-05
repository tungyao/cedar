package cedar

import (
	"fmt"
	"net/http"
)

type RestConfig struct {
	entryPath string
	apiName   string
}
type _rest struct {
	trie Trie
}

func NewRestRouter(rc RestConfig) *_rest {
	return &_rest{trie: Trie{
		num: 1,
		root: NewSon("GET", rc.entryPath+"?"+rc.apiName+"=", func(writer http.ResponseWriter, request *http.Request) {
			_, _ = fmt.Fprint(writer, "index")
		}, 1),
	}}
}
func (re *_rest) GetR(api string, fn http.HandlerFunc) {
	re.trie.Insert("GET", api+"g", fn)
}
func (re *_rest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	me, fun := re.trie.Find(r.URL.Path)
	if fun == nil || r.Method != me {
		w.Header().Set("Content-type", "text/html")
		w.WriteHeader(404)
		_, _ = w.Write([]byte("<p style=\"font-size=500px\">404</p>"))
		return
	}
	if fun != nil {
		fun(w, r)
	}
}
