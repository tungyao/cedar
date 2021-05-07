package test

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	uc "ultimate-cedar"
)

func TestRouter(t *testing.T) {
	r := uc.NewRouter()
	r.Get("ab/:id/abc", func(writer uc.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, request.URL.Fragment)
	})
	r.Get("ccc", func(writer uc.ResponseWriter, request *http.Request) {

	})
	r.Get("aaa/bbb/:id", func(writer uc.ResponseWriter, request *http.Request) {
		log.Println(request.URL.Fragment)
	})
	r.Group("a", func(groups *uc.Groups) {
		groups.Get("b", func(writer uc.ResponseWriter, request *http.Request) {
			writer.Write([]byte("get"))
		})
		groups.Patch("b", func(writer uc.ResponseWriter, request *http.Request) {
			writer.Write([]byte("trace"))
		})
	})
	http.ListenAndServe(":8000", r)
}
