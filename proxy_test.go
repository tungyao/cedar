package ultimate_cedar

import (
	"net/http"
	"testing"
)

func TestProxy(t *testing.T) {
	r := NewRouter()
	r.Proxy(50, &ProxyItem{
		Path:  "/login/wechat/echo",
		Url:   "http://login-plat-api-fdezez.huique.cn",
		Cache: true,
	})
	http.ListenAndServe(":8000", r)
}
