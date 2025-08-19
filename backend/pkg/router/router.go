package router

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
)

type Route struct {
	Handler    http.HandlerFunc
	Middleware []func(http.Handler) http.Handler
}

type Router struct {
	DB     *sql.DB
	Routes map[string]map[string]*Route
}

func NewRouter(db *sql.DB) *Router {
	return &Router{
		DB:     db,
		Routes: make(map[string]map[string]*Route),
	}
}

func (r *Router) Handle(method, path string, handler http.HandlerFunc, middleware ...func(http.Handler) http.Handler) {
	if r.Routes[method] == nil {
		r.Routes[method] = make(map[string]*Route)
	}
	r.Routes[method][path] = &Route{
		Handler:    handler,
		Middleware: middleware,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// CORS headers
	origin := req.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Serve static files
	if strings.HasPrefix(req.URL.Path, "/uploads/") {
		http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))).ServeHTTP(w, req)
		return
	}
	if strings.HasPrefix(req.URL.Path, "/assets/") || strings.HasPrefix(req.URL.Path, "/vite.svg") {
		http.ServeFile(w, req, "../frontend/dist"+req.URL.Path)
		return
	}

	// Find handler
	var handler http.Handler
	var route *Route
	var params map[string]string

	if methodRoutes, ok := r.Routes[req.Method]; ok {
		// Exact match first
		if rt, ok := methodRoutes[req.URL.Path]; ok {
			route = rt
			handler = rt.Handler
		} else {
			// Parameterized route matching
			for pathTemplate, rt := range methodRoutes {
				templateParts := strings.Split(pathTemplate, "/")
				requestParts := strings.Split(req.URL.Path, "/")

				if len(templateParts) != len(requestParts) {
					continue
				}

				match := true
				params = make(map[string]string)
				for i, part := range templateParts {
					if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
						paramName := strings.Trim(part, "{}")
						params[paramName] = requestParts[i]
					} else if part != requestParts[i] {
						match = false
						break
					}
				}

				if match {
					route = rt
					handler = rt.Handler
					// Add params to request context
					ctx := req.Context()
					for k, v := range params {
						ctx = context.WithValue(ctx, k, v)
					}
					req = req.WithContext(ctx)
					break
				}
			}
		}
	}

	if handler != nil {
		// Apply middleware
		if route != nil {
			for i := len(route.Middleware) - 1; i >= 0; i-- {
				handler = route.Middleware[i](handler)
			}
		}
		handler.ServeHTTP(w, req)
		return
	}

	// Serve frontend index.html for all other GET requests that are not API calls
	if req.Method == "GET" && !strings.HasPrefix(req.URL.Path, "/api/") {
		http.ServeFile(w, req, "../frontend/dist/index.html")
		return
	}

	http.NotFound(w, req)
}

func ForContext(ctx context.Context, key string) string {
	value, _ := ctx.Value(key).(string)
	return value
}
