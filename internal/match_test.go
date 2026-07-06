package internal

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestMatchPath(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"users", "/users", true},
		{"users", "/posts", false},
		{"users/{id}", "/users/123", true},
		{"users/{id}", "/users", false},
		{"users/{id}/posts", "/users/123/posts", true},
		{"users", "users", true},
	}
	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.path, func(t *testing.T) {
			if got := matchPath(tt.pattern, tt.path); got != tt.want {
				t.Errorf("matchPath(%q,%q) = %v want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    map[string]string
	}{
		{"users/{id}", "/users/123", map[string]string{"id": "123"}},
		{"users/{id}/posts/{postId}", "/users/456/posts/789", map[string]string{"id": "456", "postId": "789"}},
		{"users", "/users", map[string]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := ExtractPathParams(tt.pattern, tt.path)
			if len(got) != len(tt.want) {
				t.Errorf("len = %d want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("[%q] = %q want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestMatch(t *testing.T) {
	mkReq := func(method, target, body string) *http.Request {
		r, _ := http.NewRequest(method, target, strings.NewReader(body))
		return r
	}

	tests := []struct {
		name       string
		rp         ResolvedPath
		req        *http.Request
		wantRes    MatchResult
		wantParams map[string]string
		wantQuery  map[string]string
		wantBody   map[string]interface{}
	}{
		{
			name:    "exact match",
			rp:      ResolvedPath{Name: "users", Method: "GET"},
			req:     mkReq("GET", "/users", ""),
			wantRes: Matched,
		},
		{
			name:       "param match",
			rp:         ResolvedPath{Name: "users/{id}", Method: "GET"},
			req:        mkReq("GET", "/users/42", ""),
			wantRes:    Matched,
			wantParams: map[string]string{"id": "42"},
		},
		{
			name:    "method mismatch",
			rp:      ResolvedPath{Name: "users", Method: "POST"},
			req:     mkReq("GET", "/users", ""),
			wantRes: NoMatch,
		},
		{
			name:    "empty method matches any",
			rp:      ResolvedPath{Name: "users"},
			req:     mkReq("DELETE", "/users", ""),
			wantRes: Matched,
		},
		{
			name:      "query match",
			rp:        ResolvedPath{Name: "search", Method: "GET", Query: map[string]string{"q": "x"}},
			req:       mkReq("GET", "/search?q=x", ""),
			wantRes:   Matched,
			wantQuery: map[string]string{"q": "x"},
		},
		{
			name:    "query mismatch",
			rp:      ResolvedPath{Name: "search", Method: "GET", Query: map[string]string{"q": "x"}},
			req:     mkReq("GET", "/search?q=y", ""),
			wantRes: NoMatch,
		},
		{
			name:     "body contains match",
			rp:       ResolvedPath{Name: "login", Method: "POST", BodyRule: &BodyMatch{Contains: "username"}},
			req:      mkReq("POST", "/login", `{"username":"bob","password":"x"}`),
			wantRes:  Matched,
			wantBody: map[string]interface{}{"username": "bob", "password": "x"},
		},
		{
			name:    "body contains mismatch",
			rp:      ResolvedPath{Name: "login", Method: "POST", BodyRule: &BodyMatch{Contains: "username"}},
			req:     mkReq("POST", "/login", `{"email":"x"}`),
			wantRes: NoMatch,
		},
		{
			name:    "no body rule, body ignored",
			rp:      ResolvedPath{Name: "users", Method: "GET"},
			req:     mkReq("GET", "/users", "anything"),
			wantRes: Matched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData, gotRes := Match(tt.req, tt.rp)
			if gotRes != tt.wantRes {
				t.Errorf("result = %v, want %v", gotRes, tt.wantRes)
			}
			if gotRes != Matched {
				return
			}
			for k, v := range tt.wantParams {
				if gotData.Params[k] != v {
					t.Errorf("Params[%q] = %q want %q", k, gotData.Params[k], v)
				}
			}
			for k, v := range tt.wantQuery {
				if gotData.Query[k] != v {
					t.Errorf("Query[%q] = %q want %q", k, gotData.Query[k], v)
				}
			}
			for k, v := range tt.wantBody {
				if gotData.Body[k] != v {
					t.Errorf("Body[%q] = %v want %v", k, gotData.Body[k], v)
				}
			}
		})
	}
}

func TestGatherQuery(t *testing.T) {
	v := url.Values{}
	v.Add("a", "1")
	v.Add("b", "2")
	v.Add("c", "3")
	got := gatherQuery(v)
	want := map[string]string{"a": "1", "b": "2", "c": "3"}
	for k, val := range want {
		if got[k] != val {
			t.Errorf("[%q] = %q want %q", k, got[k], val)
		}
	}
}

func TestMatchQuery(t *testing.T) {
	v := url.Values{}
	v.Set("q", "x")
	if !matchQuery(map[string]string{"q": "x"}, v) {
		t.Error("expected match")
	}
	if matchQuery(map[string]string{"q": "y"}, v) {
		t.Error("expected no match")
	}
	if !matchQuery(nil, v) {
		t.Error("nil want matches all")
	}
}
