package test

import (
	"fmt"
	"net/http"
	"src/github.com/tungyao/cedar"
	"testing"
)

func TestR(t *testing.T) {
	r := cedar.NewRouter()
	// r.Get("/static/", nil, http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	// r.Get("/websocket", nil, websocket.Handler(upper))
	// r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
	//	t, _ := template.ParseFiles("./static/socket.html")
	//	t.Execute(writer, nil)
	// }, nil)
	r.Get("/k", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("helloxxx"))
	}, nil)
	r.Get("/kx", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("helloxxxkk"))
	}, nil)
	r.GlobalFunc("test", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Println("123213")
		return nil
	})
	r.Group("/a", func(groups *cedar.Groups) {
		groups.Group("/b", func(groups *cedar.Groups) {
			groups.Get("/c", func(writer http.ResponseWriter, request *http.Request) {
				writer.Write([]byte("hellocc"))
			}, nil)
			groups.Get("/d", func(writer http.ResponseWriter, request *http.Request) {
				writer.Write([]byte("hellodd"))
			}, nil)
		})
		groups.Get("/d", func(writer http.ResponseWriter, request *http.Request) {
			r.Template(writer, "/index")
		}, nil)
	})
	http.ListenAndServe(":80", r)
}
