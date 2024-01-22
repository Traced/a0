package web

import (
	"log"
	"net/http"
	"strings"
)

type Route struct {
	Method      string
	Path        string
	Handler     http.Handler
	HandlerFunc http.HandlerFunc
}

func (route *Route) AddToServeMux(serveMux *http.ServeMux) {
	if route.HandlerFunc != nil {
		// 默认 GET
		if route.Method == "" {
			route.Method = "GET"
		}
		serveMux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			// 请求方法与配置的方法相同或者方法为ANY
			byPassMethod := strings.ToUpper(route.Method) == "ANY" ||
				r.Method == "OPTIONS" ||
				r.Method == strings.ToUpper(route.Method)
			if byPassMethod {
				log.Println(r.Method, r.RequestURI, 200)
				route.HandlerFunc(w, r)
				return
			}
			log.Println(r.Method, r.RequestURI, 404)
		})
	} else if route.Handler != nil {
		serveMux.Handle(route.Path, route.Handler)
	}
}

func NewRouter(serveMux *http.ServeMux) *Router {
	return &Router{
		Routes:   make(RouteGroup, 0, 20),
		ServeMux: serveMux,
	}
}

type (
	RouteGroup []*Route
	Router     struct {
		Routes   RouteGroup
		ServeMux *http.ServeMux
	}
)

func (r *Router) AddRoute(route *Route) *Router {
	if r.ServeMux == nil {
		return r
	}
	route.AddToServeMux(r.ServeMux)
	r.Routes = append(r.Routes, route)
	return r
}

func (r *Router) AddRouteGroup(routeGroup RouteGroup) *Router {
	for _, route := range routeGroup {
		route.AddToServeMux(r.ServeMux)
		r.Routes = append(r.Routes, route)
	}
	return r
}
