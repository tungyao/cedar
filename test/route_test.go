package test

import (
	"../../cedar"
	"fmt"
	"net/http"
	"testing"
)

func TestR(t *testing.T) {
	r := cedar.NewRestRouter(cedar.RestConfig{
		EntryPath: "blog",
		ApiName:   "api",
		Pattern:   ".",
	})
	r.Static("./static/")
	r.Index("user")
	r.Get("user", func(writer http.ResponseWriter, request *http.Request) {
		r.Template(writer, "/index")
	})
	r.Group("test", func(groups *cedar.GroupR) {
		groups.Get("one", func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprintln(writer, "test.one")
		})
		groups.Post("two", func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprintln(writer, "test.two")
		})
	})
	http.ListenAndServe(":80", r)
}
