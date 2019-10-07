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
	})
	r.GetR("user", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintln(writer, "hello")
	})
	http.ListenAndServe(":80", r)
}
