package router

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is a type for context keys.
type contextKey string

// paramsKey is the key for the path parameters in the context.
const paramsKey = contextKey("params")

// Vars returns the route variables for the current request, if any.
func Vars(r *http.Request) map[string]string {
	if rv := r.Context().Value(paramsKey); rv != nil {
		return rv.(map[string]string)
	}
	return nil
}

// Route represents a route.
type route struct {
	method  string
	path    []string
	handler http.Handler
}

// Router is a simple HTTP router.
type Router struct {
	routes []*route
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{routes: []*route{}}
}

// Handle adds a new route with a handler.
func (r *Router) Handle(path string, handler http.Handler) *route {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	route := &route{path: pathParts, handler: handler}
	r.routes = append(r.routes, route)
	return route
}

// HandleFunc adds a new route with a handler function.
func (r *Router) HandleFunc(path string, handler http.HandlerFunc) *route {
	return r.Handle(path, handler)
}

// Methods sets the HTTP methods for the route.
func (ro *route) Methods(methods ...string) {
	if len(methods) > 0 {
		ro.method = methods[0]
	}
}

// ServeHTTP dispatches the request to the handler whose pattern matches the request URL.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	reqPath := strings.Trim(req.URL.Path, "/")
	reqPathParts := strings.Split(reqPath, "/")

	for _, route := range r.routes {
		if route.method != "" && req.Method != route.method && req.Method != "OPTIONS" {
			continue
		}

		if len(route.path) != len(reqPathParts) {
			continue
		}

		params := make(map[string]string)
		match := true
		for i, part := range route.path {
			if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
				params[part[1:len(part)-1]] = reqPathParts[i]
			} else if part != reqPathParts[i] {
				match = false
				break
			}
		}

		if match {
			ctx := context.WithValue(req.Context(), paramsKey, params)
			route.handler.ServeHTTP(w, req.WithContext(ctx))
			return
		}
	}

	http.NotFound(w, req)
}
