package internal

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		rp   ResolvedPath
		want Priority
	}{
		{"exact", ResolvedPath{Name: "users"}, PriorityExact},
		{"param", ResolvedPath{Name: "users/{id}"}, PriorityParam},
		{"query", ResolvedPath{Name: "search", Query: map[string]string{"q": "x"}}, PriorityQuery},
		{"body", ResolvedPath{Name: "login", BodyRule: &BodyMatch{Contains: "u"}}, PriorityBody},
		{"param-and-query → body wins not query? body nil so query", ResolvedPath{Name: "users/{id}", Query: map[string]string{"x": "y"}}, PriorityQuery},
		{"param-and-body → body", ResolvedPath{Name: "users/{id}", BodyRule: &BodyMatch{}}, PriorityBody},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classify(tt.rp); got != tt.want {
				t.Errorf("classify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasParam(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"users", false},
		{"users/{id}", true},
		{"users/{id}/posts", true},
		{"{single}", true},
		{"", false},
		{"users{id}", false},
		{"/{leading}", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasParam(tt.name); got != tt.want {
				t.Errorf("hasParam(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsInlineJSON(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{`{"a":1}`, true},
		{"[1,2,3]", true},
		{"users.json", false},
		{"", false},
		{"  {trimmed}", false},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := isInlineJSON(tt.s); got != tt.want {
				t.Errorf("isInlineJSON(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestResolve_Simple(t *testing.T) {
	rp, err := resolve(Path{
		Name:     "users",
		Response: `{"id":1}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rp.Template != `{"id":1}` {
		t.Errorf("Template = %q, want %q", rp.Template, `{"id":1}`)
	}
	if rp.Status != http.StatusOK {
		t.Errorf("Status = %d, want %d", rp.Status, http.StatusOK)
	}
}

func TestResolve_Advanced(t *testing.T) {
	rp, err := resolve(Path{
		Name: "users/{id}",
		ResponseAdvanced: &ResponseAdvanced{
			Body:    `{"id":"{{.Params.id}}"}`,
			Status:  201,
			Delay:   "100ms",
			Headers: map[string]string{"X-Test": "y"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rp.Template != `{"id":"{{.Params.id}}"}` {
		t.Errorf("Template = %q", rp.Template)
	}
	if rp.Status != 201 {
		t.Errorf("Status = %d, want 201", rp.Status)
	}
	if rp.Delay != "100ms" {
		t.Errorf("Delay = %q, want 100ms", rp.Delay)
	}
	if rp.Headers["X-Test"] != "y" {
		t.Errorf("Headers[X-Test] = %q", rp.Headers["X-Test"])
	}
}

func TestResolve_AdvancedDefaultStatus(t *testing.T) {
	rp, err := resolve(Path{
		Name:             "x",
		ResponseAdvanced: &ResponseAdvanced{Body: "ok"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rp.Status != http.StatusOK {
		t.Errorf("Status = %d, want default 200", rp.Status)
	}
}

func TestResolve_FileSource(t *testing.T) {
	dir := t.TempDir()
	body := `[{"id":1,"name":"a"}]`
	path := filepath.Join(dir, "users.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	rp, err := resolve(Path{Name: "users", Response: path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rp.Template != body {
		t.Errorf("Template = %q, want %q", rp.Template, body)
	}
}

func TestResolve_MissingFile(t *testing.T) {
	_, err := resolve(Path{Name: "users", Response: "/nonexistent/path.json"})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadConfig_Full(t *testing.T) {
	dir := t.TempDir()
	body := `[{"id":1}]`
	rf := filepath.Join(dir, "users.json")
	if err := os.WriteFile(rf, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	yamlData := []byte(`
server:
  port: 8080
  host: "0.0.0.0"
paths:
  - name: users
    method: GET
    response: ` + rf + `
  - name: users/{id}
    method: GET
    responseConfig:
      body: '{"id":"{{.Params.id}}"}'
      status: 200
  - name: search
    method: GET
    query:
      q: "type:user"
    response: '[]'
  - name: login
    method: POST
    body:
      contains: "username"
    response: '{"token":"abc"}'
`)

	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, yamlData, 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", got.Server.Port)
	}
	if len(got.Resolved) != 4 {
		t.Fatalf("Resolved count = %d, want 4", len(got.Resolved))
	}

	wantPriorities := []Priority{PriorityExact, PriorityParam, PriorityQuery, PriorityBody}
	for i, want := range wantPriorities {
		if got.Resolved[i].Priority != want {
			t.Errorf("Resolved[%d].Priority = %v, want %v", i, got.Resolved[i].Priority, want)
		}
	}
}

func TestLoadConfig_MissingFile_FailsFast(t *testing.T) {
	dir := t.TempDir()
	yamlData := []byte(`
server:
  port: 8080
paths:
  - name: users
    method: GET
    response: /does/not/exist.json
`)
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, yamlData, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Fatal("expected error for missing response file")
	}
}

func TestLoadConfig_BadYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("not: [valid: yaml"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Fatal("expected error for bad YAML")
	}
}

func TestLoadConfig_MissingConfigFile(t *testing.T) {
	_, err := LoadConfig("/no/such/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}
