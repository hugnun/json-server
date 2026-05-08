package internal

import (
	"net/http"
	"strings"
)

type Route struct {
	Path    Path
	Handler http.HandlerFunc
}

type Router struct {
	routes []Route
}

func NewRouter() *Router {
	return &Router{routes: make([]Route, 0)}
}

func (r *Router) AddRoute(path Path, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{Path: path, Handler: handler})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		if matchRoute(route.Path, req) {
			route.Handler(w, req)
			return
		}
	}
	http.NotFound(w, req)
}

func matchRoute(path Path, req *http.Request) bool {
	if path.Method != "" && path.Method != req.Method {
		return false
	}

	if !matchPath(path.Name, req.URL.Path) {
		return false
	}

	if path.Query != nil {
		for k, v := range path.Query {
			if req.URL.Query().Get(k) != v {
				return false
			}
		}
	}

	// Body matching is handled in the handler since we need to read the request body
	// Return true here to allow the route to be matched; actual body check happens in handler

	return true
}

func matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}

	if strings.HasPrefix(path, "/") && !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		if part != pathParts[i] {
			return false
		}
	}

	return true
}

func ExtractPathParams(pattern, path string) map[string]string {
	params := make(map[string]string)

	if strings.HasPrefix(path, "/") && !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			key := strings.Trim(part, "{}")
			if i < len(pathParts) {
				params[key] = pathParts[i]
			}
		}
	}

	return params
}