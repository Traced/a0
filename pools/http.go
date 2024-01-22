package pools

import (
	"io"
	"net/http"
	"sync"

	"github.com/Traced/a0/utils"
)

var (
	clients = sync.Pool{
		New: func() interface{} {
			return utils.NewClient()
		},
	}
	beforeRequestHooks = make([]func(*http.Client), 0, 300)
)

func GetHttpClientPool() *sync.Pool {
	return &clients
}

func AddRequestHook(h func(*http.Client)) {
	beforeRequestHooks = append(beforeRequestHooks, h)
}

func getRequestHook() func(*http.Client) {
	if len(beforeRequestHooks) == 0 {
		return nil
	}
	defer func() {
		if len(beforeRequestHooks) > 1 {
			beforeRequestHooks = beforeRequestHooks[1:]
		} else {
			beforeRequestHooks = beforeRequestHooks[:0]
		}
	}()
	return beforeRequestHooks[0]
}

type (
	StringMap           map[string]string
	HttpResponseHandler func(r *http.Response, e error)
)

func HttpRequest(method, url string, data io.Reader, headers StringMap, cookie []*http.Cookie, handle HttpResponseHandler) {
	c := GetHttpClient()
	defer clients.Put(c)
	if h := getRequestHook(); h != nil {
		h(c)
	}
	r, err := utils.RequestWithHeader(c, method, url, data, headers, cookie)
	if handle != nil {
		handle(r, err)
	}
}

func HttpHead(url string, headers StringMap, cookie []*http.Cookie, handle HttpResponseHandler) {
	HttpRequest("HEAD", url, nil, headers, cookie, handle)
}

func HttpGet(url string, headers StringMap, cookie []*http.Cookie, handle HttpResponseHandler) {
	HttpRequest("GET", url, nil, headers, cookie, handle)
}

func HttpPost(url string, data io.Reader, headers StringMap, cookie []*http.Cookie, handle HttpResponseHandler) {
	HttpRequest("POST", url, data, headers, cookie, handle)
}

func GetHttpClient() *http.Client {
	return clients.Get().(*http.Client)
}
