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
	r.Get("ab/:id/abc", func(writer uc.ResponseWriter, request uc.Request) {
		fmt.Fprintln(writer, request.URL.Fragment)
	})
	r.Get("ccc", func(writer uc.ResponseWriter, request uc.Request) {
		writer.Json.ContentType("application/json").
			AddHeader("time", "unix").
			Data(map[string]string{"a": "b"}).Decode("123123").Send()
	})
	r.Get("aaa/bbb/:id", func(writer uc.ResponseWriter, request uc.Request) {
		log.Println(request.URL.Fragment)
	})
	r.Group("a", func(groups *uc.Groups) {
		groups.Get("b", func(writer uc.ResponseWriter, request uc.Request) {
			writer.Write([]byte("get"))
		})
		groups.Patch("b", func(writer uc.ResponseWriter, request uc.Request) {
			writer.Write([]byte("trace"))
		})
	})
	if err := http.ListenAndServe(":9000", r); err != nil {
		log.Fatalln(err)
	}
}
