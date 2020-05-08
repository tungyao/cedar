package test

import (
	"fmt"
	"html/template"
	"net/http"
	"testing"

	"../../cedar"
)

func TestWebsocket(t *testing.T) {
	r := cedar.NewRouter()
	r.Get("/static/", nil, http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	// r.Get("/websocket", nil, websocket.Handler(upper))
	r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		t, _ := template.ParseFiles("./static/socket.html")
		t.Execute(writer, nil)
	}, nil)
	http.ListenAndServe(":80", r)
}
func TestDynamic(t *testing.T) {
	r := cedar.NewRouter()
	r.Dynamic("dynamic.yml")
	r.Get("/reset", func(writer http.ResponseWriter, request *http.Request) {
		r.Dynamic("dynamic.yml")
		writer.Write([]byte("refused success"))
	}, nil)
	http.ListenAndServe(":80", r)

}
func TestNormalGlobal(t *testing.T) {
	r := cedar.NewRouter()
	r.Get("/k", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("helloxxx"))
	}, nil)
	r.Get("/kx", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("helloxxxkk"))
	}, nil)
	r.GlobalFunc("test", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Println("global func run")
		return nil
	})
	http.ListenAndServe(":80", r)
}
func TestGroup(t *testing.T) {
	r := cedar.NewRouter()
	r.Middleware("test", func(w http.ResponseWriter, r *http.Request) bool {
		http.Redirect(w, r, "/a/b/c", 304)
		return false
	})
	r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello"))
	}, nil, "test")
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
		}, nil, "test")
	})
	http.ListenAndServe(":82", r)
}
func TestParam(t *testing.T) {
	r := cedar.NewRouter()
	r.Get("/index/abd/:id", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println(request.URL.Fragment)
		writer.Write([]byte("hello"))
	}, nil)
	http.ListenAndServe(":8000", r)
}
