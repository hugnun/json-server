package internal

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuild_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	body := `[{"id":1,"name":"a"}]`
	rf := filepath.Join(dir, "users.json")
	if err := os.WriteFile(rf, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(dir, "config.yaml")
	yamlData := []byte(`
server:
  port: 0
  cors:
    enabled: true
    origins: ["*"]
    headers: ["Content-Type"]
  logging:
    enabled: true
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
	if err := os.WriteFile(cfgPath, yamlData, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	handler, err := Build(cfg)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	srv := httptest.NewServer(handler)
	defer srv.Close()

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{"exact", "GET", "/users", "", 200, body},
		{"param", "GET", "/users/42", "", 200, `{"id":"42"}`},
		{"query match", "GET", "/search?q=type:user", "", 200, `[]`},
		{"query mismatch", "GET", "/search?q=other", "", 404, ""},
		{"body match", "POST", "/login", `{"username":"bob","password":"x"}`, 200, `{"token":"abc"}`},
		{"body mismatch", "POST", "/login", `{"email":"x"}`, 404, ""},
		{"not found", "GET", "/nope", "", 404, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req, _ = http.NewRequest(tt.method, srv.URL+tt.path, strings.NewReader(tt.body))
			} else {
				req, _ = http.NewRequest(tt.method, srv.URL+tt.path, nil)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			bodyBytes := readAndClose(t, resp)
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d want %d", resp.StatusCode, tt.wantStatus)
			}
			if tt.wantBody != "" {
				if string(bodyBytes) != tt.wantBody {
					t.Errorf("body = %q want %q", bodyBytes, tt.wantBody)
				}
			}
		})
	}
}

func TestBuild_Priority_ExactBeatsParam(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	yamlData := []byte(`
server:
  port: 0
paths:
  - name: users/{id}
    method: GET
    responseConfig:
      body: '{"id":"{{.Params.id}}"}'
  - name: users/me
    method: GET
    response: '{"id":"me"}'
`)
	if err := os.WriteFile(cfgPath, yamlData, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	h, err := Build(cfg)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/users/me")
	if err != nil {
		t.Fatal(err)
	}
	body := readAndClose(t, resp)
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if string(body) != `{"id":"me"}` {
		t.Errorf("body = %q want exact-match body", body)
	}
}

func TestBuild_DeadlineAppliesDelay(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	yamlData := []byte(`
server:
  port: 0
paths:
  - name: slow
    method: GET
    responseConfig:
      body: 'ok'
      delay: 50ms
`)
	if err := os.WriteFile(cfgPath, yamlData, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	h, err := Build(cfg)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	start := time.Now()
	resp, err := http.Get(srv.URL + "/slow")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	defer func() { _ = resp.Body.Close() }()
	elapsed := time.Since(start)
	if elapsed < 40*time.Millisecond {
		t.Errorf("elapsed = %v want >= 40ms", elapsed)
	}
}

func TestBuild_TemplateErrorReturns500(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	yamlData := []byte(`
server:
  port: 0
paths:
  - name: bad
    method: GET
    responseConfig:
      body: '{{noSuchFunc}}'
`)
	if err := os.WriteFile(cfgPath, yamlData, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	h, err := Build(cfg)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/bad")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 500 {
		t.Errorf("status = %d want 500", resp.StatusCode)
	}
}

func readAndClose(_ *testing.T, resp *http.Response) []byte {
	defer func() { _ = resp.Body.Close() }()
	buf, _ := io.ReadAll(resp.Body)
	return buf
}
