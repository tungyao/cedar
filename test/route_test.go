package test

import (
	"../../cedar"
	"fmt"
	"net/http"
	"testing"
)

func TestR(t *testing.T) {
	r := cedar.NewRestRouter(cedar.RestConfig{
		EntryPath: "wechat",
		ApiName:   "api",
		Pattern:   ".",
	})
	r.Static("./static/")
	r.GetR("user.add", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintln(writer, "hello")
	})
	r.GroupR("test", func(groups *cedar.GroupR) {
		groups.GetR("one", func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprintln(writer, "test.one")
		})
		groups.PutR("two", func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprintln(writer, "test.two")
		})
	})
	http.ListenAndServe(":80", r)
}
