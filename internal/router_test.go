package internal

import (
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
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := matchPath(tt.pattern, tt.path); got != tt.want {
				t.Errorf("matchPath(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
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
		t.Run(tt.path, func(t *testing.T) {
			got := ExtractPathParams(tt.pattern, tt.path)
			if len(got) != len(tt.want) {
				t.Errorf("ExtractPathParams() = %v, want %v", got, tt.want)
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("ExtractPathParams()[%q] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}