package ultimate_cedar

import (
	"context"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// 反向代理模块
// 主要是用来出来在大量请求的时候的 使用缓存

type ProxyItem struct {
	Path     string
	Url      string
	Cache    bool
	MaxLimit int64 // 每秒多少次请求 触发缓存的使用
}

// go sdk 源码
func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

// go sdk 源码
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	urls, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(urls)
	proxy.Director = func(request *http.Request) {
		targetQuery := urls.RawQuery
		request.URL.Scheme = urls.Scheme
		request.URL.Host = urls.Host
		request.Host = urls.Host
		request.URL.Path, request.URL.RawPath = joinURLPath(urls, request.URL)

		if targetQuery == "" || request.URL.RawQuery == "" {
			request.URL.RawQuery = targetQuery + request.URL.RawQuery
		} else {
			request.URL.RawQuery = targetQuery + "&" + request.URL.RawQuery
		}
		if _, ok := request.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36")
		}
	}
	return proxy, nil
}

func proxyFn(proxy *ProxyItem, limiter *rate.Limiter) HandlerFunc {
	p, err := NewProxy(proxy.Url)
	if err != nil {
		log.Panic(err)
	}
	return func(writer ResponseWriter, request Request) {
		err := limiter.Wait(context.Background())
		if err != nil {
			log.Println(err)
			return
		}
		p.ServeHTTP(writer.ResponseWriter, request.Request)
		limiter.Allow()
	}
}
