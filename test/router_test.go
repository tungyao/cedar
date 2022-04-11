package test

import (
	"context"
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
		writer.
			ContentType("application/json").
			AddHeader("time", "unix").
			Data(map[string]string{"a": "b"}).
			Status(200).
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

		// new func
		request.Query.Get("key")
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
	logMiddleware := uc.MiddlewareInterceptor(func(writer uc.ResponseWriter, request uc.Request, handlerFunc uc.HandlerFunc) {
		log.Println("log", request.URL.String())
		// add context
		request.Context = context.WithValue(request.Context, "member", "hello")
		handlerFunc(writer, request)
	})
	middleware := uc.MiddlewareChain{
		echoMiddleware,
	}
	logMiddlewareGroup := uc.MiddlewareChain{
		logMiddleware,
	}
	r.Get("test_middle", middleware.Handler(func(writer uc.ResponseWriter, request uc.Request) {
		request.Query.Check()
		writer.Data("hello world").Send()
	}))

	// test new middleware
	r.Get("test_new_middle", func(writer uc.ResponseWriter, request uc.Request) {
		writer.Data("hello new world").Send()
	}, middleware)
	// test new middleware for group
	r.Group("new_middle", func(groups *uc.Groups) {
		groups.Get("echo", func(writer uc.ResponseWriter, request uc.Request) {

			// add context
			log.Println(request.Context.Value("member"))
			writer.Data("hello new_middle echo").Send()
		}, logMiddlewareGroup)
		groups.Post("echo", func(writer uc.ResponseWriter, request uc.Request) {

		})
		groups.Patch("echo", func(writer uc.ResponseWriter, request uc.Request) {

		})
		groups.Put("echo", func(writer uc.ResponseWriter, request uc.Request) {

		})
		groups.Options("echo", func(writer uc.ResponseWriter, request uc.Request) {

		})
		groups.Connect("echo", func(writer uc.ResponseWriter, request uc.Request) {

		})
		groups.Head("echo", func(writer uc.ResponseWriter, request uc.Request) {

		})
	}, middleware)
	if err := http.ListenAndServe(":9000", r); err != nil {
		log.Fatalln(err)
	}
}

func TestEncryption(t *testing.T) {
	r := uc.NewRouter()
	r.Get("en", func(writer uc.ResponseWriter, request uc.Request) {
		writer.Data("hello world").Encode("F431jiyr3e0ag3wiAygjjTur0fh84sLr").Send()
	})
	r.Post("de", func(writer uc.ResponseWriter, request uc.Request) {
		t.Log(request.Decode("", nil))
	})
	http.ListenAndServe(":9000", r)

}

func TestWebsocket(t *testing.T) {
	r := uc.NewRouter()
	// r.Debug()
	r.Get("/ws", func(writer uc.ResponseWriter, request uc.Request) {
		uc.WebsocketSwitchProtocol(writer, request, "123", func(value *uc.CedarWebSocketBuffReader) {
			log.Println(value)
		})
	})
	r.Post("/ws/push", func(writer uc.ResponseWriter, request uc.Request) {
		uc.WebsocketSwitchPush("123", 0x1, []byte("hello world"))
	})
	http.ListenAndServe(":8080", r)
}
