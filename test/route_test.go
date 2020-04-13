package test

import (
	"../../cedar"
	"encoding/json"
	"fmt"
	"golang.org/x/net/webdav"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"
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
func ToJSon(x interface{}, args ...interface{}) []byte {
	tt := reflect.TypeOf(x).Elem()
	vv := reflect.ValueOf(x).Elem()
	for i := 0; i < tt.NumField(); i++ {
		switch vv.Field(i).Kind() {
		case reflect.String:
			vv.Field(i).SetString(args[tt.Field(i).Index[0]].(string))
		case reflect.Int:
			vv.Field(i).SetInt(int64(args[tt.Field(i).Index[0]].(int)))
		case reflect.Int64:
			vv.Field(i).SetInt(args[tt.Field(i).Index[0]].(int64))
		case reflect.Bool:
			vv.Field(i).SetBool(args[tt.Field(i).Index[0]].(bool))
		case reflect.Slice:
			vv.Field(i).SetBytes(args[tt.Field(i).Index[0]].([]byte))
		case reflect.Float64:
			vv.Field(i).SetFloat(args[tt.Field(i).Index[0]].(float64))
		case reflect.Float32:
			vv.Field(i).SetFloat(args[tt.Field(i).Index[0]].(float64))

		}
	}
	d, err := json.Marshal(x)
	if err != nil {
		return nil
	}
	return d
}

type AuthJson struct {
	Status int
	Msg    string
	Data   List
	Time   int64
	Op     int
}
type List struct {
	Id             int
	Name           string
	Appid          string
	Secret         string
	Status         int
	LastUpdateTime int64
	LastUpdateIp   string
	Domino         string
	RequestTimes   int
}
type ListJson struct {
	Data []List
	Time int64
}
type Delete struct {
	Id     string
	Time   int64
	Status int
}

func TestAPI(t *testing.T) {
	r := cedar.NewRouter()
	r.Group("/api", func(g *cedar.Groups) {
		g.Group("/v1", func(gg *cedar.Groups) {
			gg.Get("/auth", func(writer http.ResponseWriter, request *http.Request) {
				l := new(AuthJson)
				l.Data.Name = "helloworld"
				l.Data.Id = 1
				l.Data.Status = 1
				l.Data.Appid = "1231kjasjdkljkl1j23"
				l.Data.Secret = "asdokajksdjkjwqkjklasjkd123klljis"
				l.Data.Domino = "localhost"
				l.Data.LastUpdateIp = "127.0.0.1"
				l.Data.LastUpdateTime = time.Now().Unix()
				l.Data.RequestTimes = 4
				l.Status = 1
				l.Time = time.Now().Unix()
				l.Msg = "not ok"
				l.Op = 1
				l.Time = time.Now().Unix()
				d, _ := json.Marshal(l)
				writer.Write(d)
			}, nil)
			gg.Group("/backend", func(ggg *cedar.Groups) {
				ggg.Group("/list", func(groups *cedar.Groups) {
					groups.Get("/all", func(writer http.ResponseWriter, request *http.Request) {
						l := new(ListJson)
						for i := 0; i < 10; i++ {
							l.Data = append(l.Data, List{Id: i, Name: "Name" + strconv.Itoa(i), Status: i % 2, LastUpdateTime: time.Now().Unix(), LastUpdateIp: "123.123.123.123", Domino: "test.domino", RequestTimes: i, Appid: "123123123123", Secret: "sadjajsdiasjodas"})
						}
						l.Time = time.Now().Unix()
						d, _ := json.Marshal(l)
						writer.Write(d)
					}, nil)
					groups.Get("/one", func(writer http.ResponseWriter, request *http.Request) {
						writer.Write(ToJSon(&List{}, 1, "Name2", 2%2, time.Now().Unix(), "123.123.123.123", "test.domino", 3))
					}, nil)
				})
				ggg.Delete("/operation", func(writer http.ResponseWriter, request *http.Request) {
					fmt.Println(request.URL.RawQuery)
					data := make([]byte, 1024)
					request.Body.Read(data)
					fmt.Println(string(data))
					writer.Write(ToJSon(&Delete{}, "1", time.Now().Unix(), 1))
				}, nil)
				ggg.Put("/operation", func(writer http.ResponseWriter, request *http.Request) {
					fmt.Println(request.URL.RawQuery)
					data := make([]byte, 1024)
					request.Body.Read(data)
					fmt.Println(string(data))
					writer.Write(ToJSon(&Delete{}, "1", time.Now().Unix(), 1))
				}, nil)
				ggg.Post("/operation", func(writer http.ResponseWriter, request *http.Request) {
					fmt.Println(request.URL.RawQuery)
					data := make([]byte, 1024)
					request.Body.Read(data)
					fmt.Println(string(data))
					writer.Write(ToJSon(&Delete{}, "1", time.Now().Unix(), 1))
				}, nil)
			})
		})
	})

	http.ListenAndServe(":800", r)
}

type User struct {
	Name string
	Pass string
}

var checkUser [2]User

func TestFTP(t *testing.T) {
	checkUser[0] = User{
		Name: "abcd",
		Pass: "1121",
	}
	checkUser[1] = User{
		Name: "feng",
		Pass: "1121331",
	}
	fs := &webdav.Handler{
		FileSystem: webdav.Dir("D:\\phpstudy_pro\\WWW\\test"),
		LockSystem: webdav.NewMemLS(),
	}
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var p bool
		for _, v := range checkUser {
			if v.Name != username && v.Pass != password {
				continue
			}
			p = true
		}
		if !p {
			http.Error(w, "WebDAV: need authorized!", http.StatusUnauthorized)
			return
		}
		fs.ServeHTTP(w, req)
	})
	http.ListenAndServe(":4444", nil)
}
