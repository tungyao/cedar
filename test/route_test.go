package test

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"reflect"
	"testing"
	"unsafe"

	"../../cedar"
	"./router/v1"
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
	r.Middleware("test", func(w http.ResponseWriter, r *http.Request, c *cedar.Core) bool {
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

var Data map[string]string

type TestPlugin struct {
	cedar.Plugin
}

func (tp *TestPlugin) AutoStart() *TestPlugin {
	fmt.Println("插件初始加载")
	Data = make(map[string]string)
	return &TestPlugin{}
}
func (tp *TestPlugin) AutoBefore(w http.ResponseWriter, r *http.Request, co *cedar.Core) {
	fmt.Println("插件运行前加载")
	fmt.Println(Data)
}
func (tp *TestPlugin) Set(key, value string) {
	Data[key] = value
}
func (tp *TestPlugin) Get(key string) string {
	return Data[key]
}
func PageAppIndex(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
	r.Plugin("TestPlugin").Call("Set", request.URL.Query().Get("key"), request.URL.Query().Get("key"))
	// r.View().Assign("name", "hello").Render("app/index")
}
func TestParam(t *testing.T) {
	r := cedar.NewRouter()
	r.SetDebug()
	r.SetLayout()
	r.Plugin(&TestPlugin{})
	r.Get("/", PageAppIndex, nil)
	r.Get("/get", func(writer http.ResponseWriter, request *http.Request, core *cedar.Core) {
		valye := core.Plugin("TestPlugin").Call("Get", request.URL.Query().Get("key"))
		writer.Write([]byte(valye[0].String()))
	}, nil)
	http.ListenAndServe(":8000", r)
}
func byt(s string) []byte {
	rs := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	return *(*[]byte)(unsafe.Pointer(&rs))
}
func TestAuto(t *testing.T) {
	r := cedar.NewRouter("localhost", "localhost")
	r.SetDebug()
	r.AutoRegister(&v1.Auto{})
	http.ListenAndServe(":8000", r)
}

func AppIndex(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
	r.Json().Success(map[string]string{"name": "cedar"})
}
func TestSession(t *testing.T) {
	r := cedar.NewRouter("localhost", "localhost")
	r.Get("/set", func(writer http.ResponseWriter, request *http.Request, core *cedar.Core) {
		core.Session.Set("a", rand.Intn(1000))
	}, nil)
	r.Get("/get", func(writer http.ResponseWriter, request *http.Request, core *cedar.Core) {

		fmt.Println(core.Session.Get("a"), 123)
	}, nil)
	http.ListenAndServe(":800", r)
}
