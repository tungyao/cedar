package test

import (
	"fmt"
	"net/http"
	"src/github.com/tungyao/cedar"
	"testing"
)

func TestR(t *testing.T) {
	r := cedar.NewRouter()
	r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello"))
	}, nil)
	r.Post("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello_post"))
	}, nil)
	r.Group("/test", func(groups *cedar.Groups) {
		groups.Get("/a", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte("test_a"))
		}, nil)
	})
	http.ListenAndServe(":80", r)
	// r.Static("./static/")
	// r.Index("user")
	//
	// r.Get("user", func(writer http.ResponseWriter, request *http.Request) {
	// 	w
	// },nil)
	// r.Group("test", func(groups *cedar.GroupR) {
	// 	groups.Get("one", func(writer http.ResponseWriter, request *http.Request) {
	// 		fmt.Fprintln(writer, "test.one")
	// 	},nil)
	// 	groups.Post("two", func(writer http.ResponseWriter, request *http.Request) {
	// 		fmt.Fprintln(writer, "test.two")
	// 	},nil)
	// })
	// http.ListenAndServe(":80", r)
}
func TestOther(t *testing.T) {
	pattern := "ge_/index"
	fmt.Println(pattern[2:])
}
