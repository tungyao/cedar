package test

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

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
		request.Context.Set("member", "hello")
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
	r.SetWebsocketModel(uc.OnlyPush)
	log.Println(uc.OnlyPush, uc.ReadPush)
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
	// r.Debug()
	r.Get("/ws", func(writer uc.ResponseWriter, request uc.Request) {
		n := rand.Intn(1000)
		log.Println("rand number", n)
		uc.WebsocketSwitchProtocol(writer, request, strconv.Itoa(n), func(value *uc.CedarWebSocketBuffReader, w *uc.CedarWebsocketWriter) {
			log.Println(string(value.Data))
		})
	})
	r.Post("/ws/push", func(writer uc.ResponseWriter, request uc.Request) {
		err := uc.WebsocketSwitchPush("123", 0x1, []byte(`{"key":"123","data":"hello world"}`))
		if err != nil {
			return
		}
	})
	http.ListenAndServe(":8080", r)
}

var mux sync.Mutex

func TestLock(t *testing.T) {
	var key uint32
	var wait sync.WaitGroup
	wait.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			mux.Lock()
			defer mux.Unlock()
			// atomic.AddUint32(&key, 1)
			time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
			key += 1
			fmt.Println(key)
			wait.Done()
		}()
	}
	wait.Wait()
}

// func TestHeapSort(t *testing.T) {
// 	var arr = []int{1, 2, 124, 3, 2, 124, 341, 412, 5555}
// 	heapsort(arr)
// }
// func heapsort(arr []int) {
// 	length := len(arr)
// 	buildMaxHeap(arr, length)
// 	for i := 0; i < len(arr)-1; i++ {
// 		swap(arr, 0, i)
// 		length--
// 		heapify(arr, 0, length)
// 	}
// 	log.Println(arr)
// }
// func buildMaxHeap(arr []int, arrLen int) {
// 	for i := arrLen / 2; i >= 0; i-- {
// 		heapify(arr, i, arrLen)
// 	}
// }
// func heapify(arr []int, i int, leng int) {
// 	left, right, largest := 2*i+1, 2*i+2, i
// 	if left < leng && arr[left] > arr[largest] {
// 		largest = left
// 	}
// 	if right < leng && arr[right] > arr[largest] {
// 		largest = right
// 	}
// 	if largest != i {
// 		arr[i], arr[largest] = arr[largest], arr[i]
// 		heapify(arr, largest, leng)
// 	}
// }
// func swap(arr []int, i int, j int) {
//
// }
func TestArr(t *testing.T) {
	var arr = []int{1, 2, 124, 3, 2, 124, 341, 412, 5555}
	arr = arr[2:]
	log.Println(arr)
}
