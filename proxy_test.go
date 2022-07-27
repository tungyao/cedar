package ultimate_cedar

import (
	"net/http"
	"testing"
)

func TestProxy(t *testing.T) {
	r := NewRouter()
	r.Proxy(&ProxyItem{
		Path:     "/ccc",
		Url:      "http://192.168.77.12:9000",
		Cache:    true,
		MaxLimit: 1000,
	})
	http.ListenAndServe(":8000", r)
}
