package test

import (
	"fmt"
	"net/http"
	"testing"

	uc "ultimate-cedar"
)

func TestRouter(t *testing.T) {
	r := uc.NewRouter()
	r.Get("ab/abc", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello world")
	})
	r.Get("ab/abc/:id", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, request.URL.Query().Get("id"))
	})
	r.Get("ccc", func(writer http.ResponseWriter, request *http.Request) {

	})
	r.Get("aaa/bbb/:id", func(writer http.ResponseWriter, request *http.Request) {

	})
	http.ListenAndServe(":8000", r)
}
