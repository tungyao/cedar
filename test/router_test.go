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
	r.Get("ab/abc", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello world")
	})
	r.Get("ab/:id/abc", func(writer http.ResponseWriter, request *http.Request) {
		log.Println(request.URL.Fragment)
		fmt.Fprintln(writer, request.URL.Fragment)
	})
	r.Get("/ccc", func(writer http.ResponseWriter, request *http.Request) {

	})
	r.Get("aaa/bbb/:id", func(writer http.ResponseWriter, request *http.Request) {

	})
	http.ListenAndServe(":8000", r)
}
