package test

import (
	"../../cedar"
	"fmt"
	"net/http"
	"src/github.com/tungyao/tpool"
	"testing"
)

func TestR(t *testing.T) {
	r := cedar.NewRestRouter(cedar.RestConfig{
		EntryPath: "wechat",
		ApiName:   "api",
		Pattern:   ".",
	})
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
func TestRR(t *testing.T) {
	r := cedar.NewRouter()
	r.Get("/a", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "hello")
	})
	r.Listening(":80", r)
	e := tpool.NewTask(func() error {
		return nil
	})
	p := tpool.NewPool(1000)
	go func() {
		for {
			p.EntryChannel <- e
		}
	}()
	p.Run()
}
