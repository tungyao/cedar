package test

import (
	"fmt"
	"net/http"
	"testing"

	uc "ultimate-cedar"
)

func TestRouter(t *testing.T) {
	r := uc.NewRouter()
	r.Get("ab/abc", func(w uc.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello world")
	})
	r.Get("ab/:id/abc", func(writer uc.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, request.URL.Fragment)
	})
	r.Get("/ccc", func(writer uc.ResponseWriter, request *http.Request) {

	})
	r.Get("aaa/bbb/:id", func(writer uc.ResponseWriter, request *http.Request) {

	})
	http.ListenAndServe(":8000", r)
}
