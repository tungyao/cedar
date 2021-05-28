package test

import (
	"log"
	"net/http"
	"testing"

	uc "github.com/tungyao/ultimate-cedar"
)

func TestRouter(t *testing.T) {
	r := uc.NewRouter()
	r.ErrorTemplate(func(err error) []byte {
		return []byte(err.Error() + "12312")
	})
	r.Get("ab/:id/abc", func(writer uc.ResponseWriter, request uc.Request) {
		log.Println(request.Data.Get("id"))
		if d, err := request.Query.Check("id"); err == nil {
			log.Println(d.Get("id"), err)
			return
		} else {
			writer.Json.Status(403).Data(err).Send()
		}
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
	r.Get("test", func(writer uc.ResponseWriter, request uc.Request) {
		writer.Data(123).Send()
	})
	r.Get("test_query_check", func(writer uc.ResponseWriter, request uc.Request) {
		var err error
		if d, err := request.Query.Check("id"); err == nil {
			log.Println(d)
			return
		}
		log.Println(err)

	})
	if err := http.ListenAndServe(":9000", r); err != nil {
		log.Fatalln(err)
	}
}
