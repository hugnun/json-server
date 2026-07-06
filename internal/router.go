package internal

import (
	"fmt"
	"net/http"
	"time"
)

type route struct {
	rp ResolvedPath
}

// Router matches incoming requests to ResolvedPaths in four priority
// buckets (exact, param, query, body) and dispatches to the response
// pipeline.
type Router struct {
	exact []route
	param []route
	query []route
	body  []route
}

// NewRouter returns an empty Router.
func NewRouter() *Router {
	return &Router{}
}

// Add registers rp in the bucket for its Priority class. The router
// only accepts the four known priorities.
func (r *Router) Add(rp ResolvedPath) error {
	switch rp.Priority {
	case PriorityExact:
		r.exact = append(r.exact, route{rp})
	case PriorityParam:
		r.param = append(r.param, route{rp})
	case PriorityQuery:
		r.query = append(r.query, route{rp})
	case PriorityBody:
		r.body = append(r.body, route{rp})
	default:
		return fmt.Errorf("unknown priority %v for %q", rp.Priority, rp.Name)
	}
	return nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	buckets := [][]route{r.exact, r.param, r.query, r.body}
	for _, bucket := range buckets {
		for _, rt := range bucket {
			data, res := Match(req, rt.rp)
			switch res {
			case Matched:
				status, hdr, body, deadline, err := Resolve(req, rt.rp, data)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				writeResponse(w, status, hdr, body, deadline)
				return
			case BodyInvalid:
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			case NoMatch:
				continue
			}
		}
	}
	http.NotFound(w, req)
}

func writeResponse(w http.ResponseWriter, status int, hdr http.Header, body []byte, deadline time.Time) {
	for k, vs := range hdr {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	if !deadline.IsZero() {
		remaining := time.Until(deadline)
		if remaining > 0 {
			time.Sleep(remaining)
		}
	}
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
