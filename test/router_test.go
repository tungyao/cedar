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

	// test url params
	r.Get("ab/:id/abc", func(writer uc.ResponseWriter, request uc.Request) {
		log.Println(request.Data.Get("id"))
	})
	r.Get("m/:id/:number", func(writer uc.ResponseWriter, request uc.Request) {
		log.Println(request.Data.Get("id"))
		log.Println(request.Data.Get("number"))
	})

	// test return chain
	r.Get("ccc", func(writer uc.ResponseWriter, request uc.Request) {
		writer.Json.
			ContentType("application/json").
			AddHeader("time", "unix").
			Data(map[string]string{"a": "b"}).
			Status(403).
			Encode("123123").
			Send()
	})

	// test group
	r.Group("a", func(groups *uc.Groups) {
		groups.Get("b", func(writer uc.ResponseWriter, request uc.Request) {
			writer.Write([]byte("get"))
		})
		groups.Patch("b", func(writer uc.ResponseWriter, request uc.Request) {
			writer.Write([]byte("trace"))
		})
	})

	// test check query params
	r.Get("test_query_check", func(writer uc.ResponseWriter, request uc.Request) {
		var err error
		if d, err := request.Query.Check("id"); err == nil {
			log.Println(d)
			return
		}
		log.Println(err)
	})

	// test middleware
	echoMiddleware := uc.MiddlewareInterceptor(func(writer uc.ResponseWriter, request uc.Request, handlerFunc uc.HandlerFunc) {
		log.Println(request.URL.Query().Get("echo"))
		writer.Data("runner middle").Send()

		handlerFunc(writer, request)
	})
	middleware := uc.MiddlewareChain{
		echoMiddleware,
	}
	r.Get("test_middle", middleware.Handler(func(writer uc.ResponseWriter, request uc.Request) {
		writer.Data("hello world").Send()
	}))

	if err := http.ListenAndServe(":9000", r); err != nil {
		log.Fatalln(err)
	}
}
