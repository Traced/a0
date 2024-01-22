package utils

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	Url "net/url"
	"time"

	"golang.org/x/net/publicsuffix"
)

var (
	DefaultClient  = NewClient()
	UseSystemProxy = true
)

func PostWithHeader(client *http.Client, url string, data io.Reader, header StringMap, cookie []*http.Cookie) (resp *http.Response, err error) {
	return RequestWithHeader(client, "POST", url, data, header, cookie)
}

func GetWithHeader(client *http.Client, url string, header StringMap, cookie []*http.Cookie) (resp *http.Response, err error) {
	return RequestWithHeader(client, "GET", url, nil, header, cookie)
}

func Post(url string, data io.Reader) (resp *http.Response, err error) {
	return DefaultClient.Post(url, `application/x-www-form-urlencoded`, data)
}

func Get(url string) (resp *http.Response, err error) {
	return DefaultClient.Get(url)
}

func RequestWithHeader(client *http.Client, method, url string, data io.Reader, header StringMap, cookies []*http.Cookie) (resp *http.Response, err error) {
	req, _ := http.NewRequest(method, url, data)
	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}
	if client == nil {
		client = DefaultClient
	}
	if cookies != nil {
		if client.Jar != nil {
			cookieURL, _ := Url.Parse(url)
			client.Jar.SetCookies(cookieURL, cookies)
		} else {
			for _, cookie := range cookies {
				req.AddCookie(cookie)
			}
		}
	}
	return client.Do(req)
}

func NewProxyClient(proxyAddr string) (client *http.Client) {
	client = NewClient()
	if proxyAddr != "" {
		proxy, _ := Url.Parse(proxyAddr)
		client.Transport.(*http.Transport).Proxy = http.ProxyURL(proxy)
	}
	return client
}

func SetClientProxy(client *http.Client, proxyIP string) *http.Client {
	if client == nil {
		client = NewClient()
	}
	if proxyIP != "" {
		setClientProxy(client, proxyIP)
		proxy, _ := Url.Parse(proxyIP)
		client.Transport.(*http.Transport).Proxy = http.ProxyURL(proxy)
	}
	return client
}

func setClientProxy(client *http.Client, proxyIP string) *http.Client {
	proxy, _ := Url.Parse(proxyIP)
	client.Transport.(*http.Transport).Proxy = http.ProxyURL(proxy)
	return client
}

// NewClient create new http client
func NewClient() *http.Client {
	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	return &http.Client{
		Jar:     cookieJar,
		Timeout: time.Second * 60,
		// skip cert verify
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*Url.URL, error) {
				if UseSystemProxy {
					// 使用终端代理
					return http.ProxyFromEnvironment(req)
				}
				return nil, nil
			},
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 20 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 8,
			IdleConnTimeout:     3 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		},
	}
}
