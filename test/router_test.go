package test

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	uc "github.com/tungyao/ultimate-cedar"
)

func TestRouter(t *testing.T) {
	r := uc.NewRouter()

	r.Get("ab/:id/abc", func(writer uc.ResponseWriter, request uc.Request) {
		fmt.Fprintln(writer, request.Data.Get("id"))
	})
	r.Get("m/:id/:number", func(writer uc.ResponseWriter, request uc.Request) {
		writer.Write([]byte(request.Data.Get("id") + request.Data.Get("number")))
	})

	r.Get("ccc", func(writer uc.ResponseWriter, request uc.Request) {
		writer.Json.
			ContentType("application/json").
			AddHeader("time", "unix").
			Data(map[string]string{"a": "b"}).
			Status(403).
			Encode("123123").
			Send()
	})

	r.Group("a", func(groups *uc.Groups) {
		groups.Get("b", func(writer uc.ResponseWriter, request uc.Request) {
			writer.Write([]byte("Get"))
		})
		groups.Patch("b", func(writer uc.ResponseWriter, request uc.Request) {
			writer.Write([]byte("Patch"))
		})
	})
	if err := http.ListenAndServe(":9000", r); err != nil {
		log.Fatalln(err)
	}
}
