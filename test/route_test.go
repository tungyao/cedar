package test

import (
	"../../cedar"
	"fmt"
	"golang.org/x/net/websocket"
	"html/template"
	"net/http"
	"strings"
	"testing"
)

func upper(ws *websocket.Conn) {
	var err error
	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println(err)
			continue
		}

		if err = websocket.Message.Send(ws, strings.ToUpper(reply)); err != nil {
			fmt.Println(err)
			continue
		}
	}
}
func TestR(t *testing.T) {
	r := cedar.NewRouter()
	//r.Get("/static/", nil, http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.Get("/websocket", nil, websocket.Handler(upper))
	r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		t, _ := template.ParseFiles("./static/socket.html")
		t.Execute(writer, nil)
	}, nil)
	http.ListenAndServe(":80", r)
	//r := cedar.NewRestRouter(cedar.RestConfig{
	//	EntryPath: "blog",
	//	ApiName:   "api",
	//	Pattern:   ".",
	//})
	//r.Static("./static/")
	//r.Index("user")
	//
	//r.Get("user", func(writer http.ResponseWriter, request *http.Request) {
	//	r.Template(writer, "/index")
	//})
	//r.Group("test", func(groups *cedar.GroupR) {
	//	groups.Get("one", func(writer http.ResponseWriter, request *http.Request) {
	//		fmt.Fprintln(writer, "test.one")
	//	})
	//	groups.Post("two", func(writer http.ResponseWriter, request *http.Request) {
	//		fmt.Fprintln(writer, "test.two")
	//	})
	//})
	//http.ListenAndServe(":80", r)
}
