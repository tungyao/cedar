package ultimate_cedar

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// 反向代理模块
// 主要是用来出来在大量请求的时候的 使用缓存

type ProxyItem struct {
	Path     string
	Url      string
	Cache    bool
	MaxLimit int64 // 每秒多少次请求 触发缓存的使用
}

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	urls, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(urls)
	proxy.ModifyResponse = modifyResponse()
	return proxy, nil
}

func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		resp.Header.Set("X-Proxy", "Magical")
		return nil
	}
}

func proxyFn(proxy *ProxyItem) HandlerFunc {
	p, err := NewProxy(proxy.Url)
	if err != nil {
		log.Panic(err)
	}
	return func(writer ResponseWriter, request Request) {

		p.ServeHTTP(writer.ResponseWriter, request.Request)
	}
}
