package test

import (
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"testing"

	"../../cedar"
	"./router"
)

func TestWebsocket(t *testing.T) {
	r := cedar.NewRouter()
	r.Get("/static/", nil, http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	// r.Get("/websocket", nil, websocket.Handler(upper))
	r.Get("/", func(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
		t, _ := template.ParseFiles("./static/socket.html")
		t.Execute(writer, nil)
	}, nil)
	http.ListenAndServe(":80", r)
}
func TestDynamic(t *testing.T) {
	r := cedar.NewRouter()
	r.Dynamic("dynamic.yml")
	r.Get("/reset", func(writer http.ResponseWriter, request *http.Request, co *cedar.Core) {
		r.Dynamic("dynamic.yml")
		writer.Write([]byte("refused success"))
	}, nil)
	http.ListenAndServe(":80", r)

}
func TestNormalGlobal(t *testing.T) {
	r := cedar.NewRouter()
	r.Get("/k", func(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
		writer.Write([]byte("helloxxx"))
	}, nil)
	r.Get("/kx", func(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
		writer.Write([]byte("helloxxxkk"))
	}, nil)
	http.ListenAndServe(":80", r)
}
func TestGroup(t *testing.T) {
	r := cedar.NewRouter()
	r.Middleware("test", func(w http.ResponseWriter, r *http.Request) bool {
		http.Redirect(w, r, "/a/b/c", 302)
		return false
	})
	r.Get("/", func(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
		writer.Write([]byte("hello"))
	}, nil, "test")
	r.Group("/a", func(groups *cedar.Groups) {
		groups.Group("/b", func(groups *cedar.Groups) {
			groups.Get("/c", func(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
				writer.Write([]byte("hellocc"))
			}, nil)
			groups.Get("/d", func(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
				writer.Write([]byte("hellodd"))
			}, nil)
		})
		groups.Get("/d", func(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
		}, nil, "test")
	})
	http.ListenAndServe(":82", r)
}

func PageAppIndex(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
	r.View().Assign("name", "hello").Render("app/index")
}
func AppIndex(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
	r.Json().Success(map[string]string{"name": "cedar"})
}
func TestParam(t *testing.T) {
	r := cedar.NewRouter()
	r.SetDebug()
	r.SetLayout()
	r.Get("/", PageAppIndex, nil)
	r.Get("/json", AppIndex, nil)
	http.ListenAndServe(":8000", r)
}
func TestAuto(t *testing.T) {
	r := cedar.NewRouter()
	r.SetDebug()
	r.AutoRegister(&router.Auto{})
	http.ListenAndServe(":8000", r)
}

type TestX struct {
}

func (tx *TestX) Name(x http.Handler) {
	fmt.Println(123)
}
func TestAutoMethod(t *testing.T) {
	x := &TestX{}
	m := reflect.ValueOf(x).MethodByName("Name")
	m.Call([]reflect.Value{reflect.New(m.Type().In(0)).Elem()})
	// p := unsafe.Pointer(x)
	// fmt.Println(uintptr(p) + unsafe.Sizeof(p) + uintptr(16))
}
