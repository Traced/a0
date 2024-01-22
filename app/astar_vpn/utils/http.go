package utils

import (
	"crypto/tls"
	"io"
	"net/http"
)

var (
	DefaultClient = NewClient()
)

func PostWithHeader(url string, data io.Reader, header map[string]string) (resp *http.Response, err error) {
	return RequestWithHeader("POST", url, data, header)
}

func GetWithHeader(url string, header map[string]string) (resp *http.Response, err error) {
	return RequestWithHeader("GET", url, nil, header)
}

func RequestWithHeader(method, url string, data io.Reader, header map[string]string) (resp *http.Response, err error) {
	req, _ := http.NewRequest(method, url, data)
	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}
	return DefaultClient.Do(req)
}

func Post(url string, data io.Reader) (resp *http.Response, err error) {
	return DefaultClient.Post(url, `application/x-www-form-urlencoded`, data)
}

func Get(url string) (resp *http.Response, err error) {
	return DefaultClient.Get(url)
}

// NewClient create new http client
func NewClient() *http.Client {
	return &http.Client{
		// skip cert verify
		Transport: &http.Transport{
			// Proxy: func(_ *http.Request)(*url.URL, error)  {
			// 	return url.Parse("http://127.0.0.1:2021")
			// },
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
