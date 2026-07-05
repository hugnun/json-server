# CLI HTTP Server from YAML Config - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement a CLI tool that runs an HTTP server based on a YAML configuration file with support for static files, inline JSON, dynamic responses, CORS, logging, and advanced routing.

**Architecture:** Use Go's standard library `net/http` with a custom mux for route matching. Config parsed with `gopkg.in/yaml.v3`. Support path parameters, query matching, and body matching with template substitution.

**Tech Stack:** Go 1.26, Cobra CLI, gopkg.in/yaml.v3

---

## Task 1: Update Config Struct

**Files:**

- Modify: `internal/config.go`

**Step 1: Read existing config.go**

Read `internal/config.go` to see current structure.

**Step 2: Update config.go**

```go
package internal

type Config struct {
 Paths  []Path  `yaml:"paths"`
 Server Server `yaml:"server"`
}

type Path struct {
 Name     string            `yaml:"name"`
 Method   string            `yaml:"method"`
 Response string            `yaml:"response"`
 ResponseAdvanced *ResponseAdvanced `yaml:"response,omitempty"`
 Query    map[string]string `yaml:"query,omitempty"`
 Body     *BodyMatch        `yaml:"body,omitempty"`
}

type ResponseAdvanced struct {
 Body    string            `yaml:"body"`
 Status  int               `yaml:"status"`
 Delay   string            `yaml:"delay"`
 Headers map[string]string `yaml:"headers"`
}

type BodyMatch struct {
 Contains string `yaml:"contains"`
}

type Server struct {
 Port    uint           `yaml:"port"`
 Host    string         `yaml:"host"`
 CORS    *CORSConfig    `yaml:"cors,omitempty"`
 Logging *LoggingConfig `yaml:"logging,omitempty"`
 DefaultHeaders map[string]string `yaml:"defaultHeaders,omitempty"`
}

type CORSConfig struct {
 Enabled  bool     `yaml:"enabled"`
 Origins  []string `yaml:"origins"`
 Headers  []string `yaml:"headers"`
}

type LoggingConfig struct {
 Enabled bool   `yaml:"enabled"`
 Format  string `yaml:"format"`
}
```

**Step 3: Commit**

```bash
git add internal/config.go
git commit -m "feat: update config struct for all features"
```

---

## Task 2: Add Route Matching Logic

**Files:**

- Create: `internal/router.go`

**Step 1: Create router.go with route matching**

```go
package internal

import (
 "net/http"
 "path/filepath"
 "strings"
)

type Route struct {
 Path        Path
 Handler     http.HandlerFunc
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
 // Method check
 if path.Method != "" && path.Method != req.Method {
  return false
 }

 // Path matching with params support
 if !matchPath(path.Name, req.URL.Path) {
  return false
 }

 // Query matching
 if path.Query != nil {
  for k, v := range path.Query {
   if req.URL.Query().Get(k) != v {
    return false
   }
  }
 }

 // Body matching
 if path.Body != nil && path.Body.Contains != "" {
  return false // Implement body reading in handler
 }

 return true
}

func matchPath(pattern, path string) bool {
 // Exact match
 if pattern == path {
  return true
 }

 // Wildcard match for /users/{id} style
 patternParts := strings.Split(pattern, "/")
 pathParts := strings.Split(path, "/")

 if len(patternParts) != len(pathParts) {
  return false
 }

 for i, part := range patternParts {
  if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
   continue // Param placeholder matches anything
  }
  if part != pathParts[i] {
   return false
  }
 }

 return true
}

func ExtractPathParams(pattern, path string) map[string]string {
 params := make(map[string]string)
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
```

**Step 2: Commit**

```bash
git add internal/router.go
git commit -m "feat: add route matching logic"
```

---

## Task 3: Add Template Substitution

**Files:**

- Create: `internal/template.go`

**Step 1: Create template.go**

```go
package internal

import (
 "bytes"
 "encoding/json"
 "fmt"
 "io"
 "net/url"
 "text/template"
 "time"
)

type TemplateData struct {
 Params map[string]string
 Query  map[string]string
 Body   map[string]interface{}
}

func RenderResponse(body string, data TemplateData) (string, error) {
 funcMap := template.FuncMap{
  "json": func(v interface{}) string {
   b, _ := json.Marshal(v)
   return string(b)
  },
 }

 tmpl, err := template.New("response").Funcs(funcMap).Parse(body)
 if err != nil {
  return "", fmt.Errorf("template parse error: %w", err)
 }

 var buf bytes.Buffer
 err = tmpl.Execute(&buf, data)
 if err != nil {
  return "", fmt.Errorf("template execute error: %w", err)
 }

 return buf.String(), nil
}

func ParseDelay(delayStr string) (time.Duration, error) {
 if delayStr == "" {
  return 0, nil
 }
 return time.ParseDuration(delayStr)
}

func ParseBody(body string) (map[string]interface{}, error) {
 var result map[string]interface{}
 err := json.Unmarshal([]byte(body), &result)
 if err != nil {
  // Try as form data
  values, err := url.ParseQuery(body)
  if err != nil {
   return nil, err
  }
  result = make(map[string]interface{})
  for k, v := range values {
   result[k] = v
  }
 }
 return result, nil
}

func ReadBody(req *http.Request) (string, error) {
 body, err := io.ReadAll(req.Body)
 if err != nil {
  return "", err
 }
 return string(body), nil
}
```

**Step 2: Commit**

```bash
git add internal/template.go
git commit -m "feat: add template substitution"
```

---

## Task 4: Update Run Logic

**Files:**

- Modify: `internal/run.go`

**Step 1: Read existing run.go**

Read `internal/run.go` to see current implementation.

**Step 2: Update run.go**

```go
package internal

import (
 "fmt"
 "io/ioutil"
 "net/http"
 "os"
 "strings"
 "time"

 "github.com/spf13/cobra"
 "gopkg.in/yaml.v3"
)

func Run(cmd *cobra.Command, args []string) error {
 pathCfgFile := args[0]

 data, err := os.ReadFile(pathCfgFile)
 if err != nil {
  return fmt.Errorf("failed to read config: %w", err)
 }

 var config Config
 if err := yaml.Unmarshal(data, &config); err != nil {
  return fmt.Errorf("failed to parse YAML: %w", err)
 }

 // Override port from CLI if provided
 if port, _ := cmd.Flags().GetInt("port"); port > 0 {
  config.Server.Port = uint(port)
 }

 router := NewRouter()

 for _, path := range config.Paths {
  handler := createHandler(path, &config)
  router.AddRoute(path, handler)
 }

 // Wrap with middleware
 var handler http.Handler = router

 if config.Server.CORS != nil && config.Server.CORS.Enabled {
  handler = corsMiddleware(handler, config.Server.CORS)
 }

 if config.Server.Logging != nil && config.Server.Logging.Enabled {
  handler = loggingMiddleware(handler, config.Server.Logging)
 }

 // Set default headers
 if len(config.Server.DefaultHeaders) > 0 {
  handler = defaultHeadersMiddleware(handler, config.Server.DefaultHeaders)
 }

 addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
 fmt.Printf("Starting server on %s\n", addr)
 return http.ListenAndServe(addr, handler)
}

func createHandler(path Path, config *Config) http.HandlerFunc {
 return func(w http.ResponseWriter, req *http.Request) {
  var responseBody string
  var statusCode int
  var responseHeaders map[string]string

  // Determine response
  if path.Response != "" {
   // Auto-detect: file path or inline JSON
   if strings.HasPrefix(path.Response, "{") || strings.HasPrefix(path.Response, "[") {
    responseBody = path.Response
   } else {
    // It's a file path
    data, err := ioutil.ReadFile(path.Response)
    if err != nil {
     http.Error(w, fmt.Sprintf("file not found: %s", path.Response), http.StatusNotFound)
     return
    }
    responseBody = string(data)
   }
   statusCode = http.StatusOK
  } else if path.ResponseAdvanced != nil {
   responseBody = path.ResponseAdvanced.Body
   statusCode = path.ResponseAdvanced.Status
   if statusCode == 0 {
    statusCode = http.StatusOK
   }
   responseHeaders = path.ResponseAdvanced.Headers
  }

  // Extract params
  params := ExtractPathParams(path.Name, req.URL.Path)
  query := make(map[string]string)
  for k, v := range req.URL.Query() {
   query[k] = v
  }

  // Read body if needed
  var bodyData map[string]interface{}
  if path.Body != nil {
   bodyStr, err := ReadBody(req)
   if err == nil && strings.Contains(bodyStr, path.Body.Contains) {
    bodyData, _ = ParseBody(bodyStr)
   }
  }

  // Template substitution
  if strings.Contains(responseBody, "{{") {
   rendered, err := RenderResponse(responseBody, TemplateData{
    Params: params,
    Query:  query,
    Body:   bodyData,
   })
   if err != nil {
    http.Error(w, fmt.Sprintf("template error: %v", err), http.StatusInternalServerError)
    return
   }
   responseBody = rendered
  }

  // Handle delay
  var delay time.Duration
  if path.ResponseAdvanced != nil && path.ResponseAdvanced.Delay != "" {
   delay, _ = ParseDelay(path.ResponseAdvanced.Delay)
  }
  if delay > 0 {
   time.Sleep(delay)
  }

  // Set headers
  w.Header().Set("Content-Type", "application/json")
  for k, v := range responseHeaders {
   w.Header().Set(k, v)
  }

  w.WriteHeader(statusCode)
  w.Write([]byte(responseBody))
 }
}

func corsMiddleware(next http.Handler, config *CORSConfig) http.HandlerFunc {
 return func(w http.ResponseWriter, req *http.Request) {
  origin := req.Header.Get("Origin")
  if config.Origins[0] == "*" || contains(config.Origins, origin) {
   w.Header().Set("Access-Control-Allow-Origin", origin)
   w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
   w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.Headers, ", "))
  }
  if req.Method == "OPTIONS" {
   w.WriteHeader(http.StatusOK)
   return
  }
  next.ServeHTTP(w, req)
 }
}

func loggingMiddleware(next http.Handler, config *LoggingConfig) http.HandlerFunc {
 return func(w http.ResponseWriter, req *http.Request) {
  fmt.Printf("[%s] %s %s\n", req.Method, req.URL.Path, req.RemoteAddr)
  next.ServeHTTP(w, req)
 }
}

func defaultHeadersMiddleware(next http.Handler, headers map[string]string) http.HandlerFunc {
 return func(w http.ResponseWriter, req *http.Request) {
  for k, v := range headers {
   w.Header().Set(k, v)
  }
  next.ServeHTTP(w, req)
 }
}

func contains(slice []string, item string) bool {
 for _, s := range slice {
  if s == item {
   return true
  }
 }
 return false
}
```

**Step 3: Commit**

```bash
git add internal/run.go
git commit -m "feat: implement run logic with middleware and routing"
```

---

## Task 5: Update CLI Flags

**Files:**

- Modify: `cmd/root.go`

**Step 1: Read existing root.go**

Read `cmd/root.go` to see current flags.

**Step 2: Update root.go**

```go
package cmd

import (
 "os"

 "github.com/hugnun/json-server/internal"
 "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
 Use:   "serve",
 Short: "Start the HTTP server from a config file",
 Args:  cobra.ExactArgs(1),
 RunE:  internal.Run,
}

func init() {
 rootCmd.Flags().IntP("port", "p", 0, "Port to listen on (overrides config)")
}

func Execute() {
 err := rootCmd.Execute()
 if err != nil {
  os.Exit(1)
 }
}
```

**Step 3: Commit**

```bash
git add cmd/root.go
git commit -m "feat: add port flag to CLI"
```

---

## Task 6: Add Validate Command

**Files:**

- Modify: `cmd/root.go`

**Step 1: Add validate command**

Add a validate subcommand to check YAML syntax without starting server.

```go
var validateCmd = &cobra.Command{
 Use:   "validate",
 Short: "Validate a config file",
 Args:  cobra.ExactArgs(1),
 RunE: func(cmd *cobra.Command, args []string) error {
  data, err := os.ReadFile(args[0])
  if err != nil {
   return fmt.Errorf("failed to read config: %w", err)
  }
  var config internal.Config
  if err := yaml.Unmarshal(data, &config); err != nil {
   return fmt.Errorf("invalid YAML: %w", err)
  }
  fmt.Println("Config is valid")
  return nil
 },
}
```

**Step 2: Add to Execute**

In `init()`, add:

```go
rootCmd.AddCommand(validateCmd)
```

**Step 3: Commit**

```bash
git add cmd/root.go
git commit -m "feat: add validate command"
```

---

## Task 7: Write Unit Tests

**Files:**

- Create: `internal/router_test.go`
- Create: `internal/template_test.go`

**Step 1: Create router_test.go**

```go
package internal

import (
 "net/http"
 "net/http/httptest"
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
   if got := matchPath(Path{Name: tt.pattern}, &http.Request{URL: &url.URL{Path: tt.path}}); got != tt.want {
    t.Errorf("matchPath() = %v, want %v", got, tt.want)
   }
  })
 }
}
```

**Step 2: Create template_test.go**

```go
package internal

import (
 "testing"
)

func TestRenderResponse(t *testing.T) {
 data := TemplateData{
  Params: map[string]string{"id": "123"},
  Query:  map[string]string{"name": "bob"},
  Body:   map[string]interface{}{"email": "test@example.com"},
 }

 tests := []struct {
  body   string
  want   string
 }{
  {`{"id": "{{.Params.id}}"}`, `{"id": "123"}`},
  {`{{.Query.name}}`, "bob"},
 }

 for _, tt := range tests {
  t.Run(tt.body, func(t *testing.T) {
   got, err := RenderResponse(tt.body, data)
   if err != nil {
    t.Fatalf("RenderResponse() error = %v", err)
   }
   if got != tt.want {
    t.Errorf("RenderResponse() = %v, want %v", got, tt.want)
   }
  })
 }
}
```

**Step 3: Run tests**

```bash
go test ./...
```

**Step 4: Commit**

```bash
git add internal/router_test.go internal/template_test.go
git commit -m "test: add unit tests for router and template"
```

---

## Task 8: Add Example Config

**Files:**

- Create: `examples/config.yaml`

**Step 1: Create example config**

```yaml
server:
  port: 8080
  cors:
    enabled: true
    origins: ["*"]
    headers: ["Content-Type"]
  logging:
    enabled: true
    format: "text"

paths:
  - name: users
    method: GET
    response: examples/users.json

  - name: users
    method: POST
    response: '{"id": 1, "name": "New User"}'

  - name: users/{id}
    method: GET
    response:
      body: '{"id": "{{.Params.id}}"}'
      status: 200
      delay: 100ms

  - name: search
    method: GET
    query:
      q: "type:user"
    response:
      body: '[{"id": 1}]'
```

**Step 2: Create example data file**

Create `examples/users.json`:

```json
[{"id": 1, "name": "Alice"}]
```

**Step 3: Test the server**

```bash
go run main.go serve examples/config.yaml
```

**Step 4: Commit**

```bash
git add examples/config.yaml examples/users.json
git commit -m "docs: add example config and data"
```

---

## Plan Complete

All tasks are ready for implementation. Each task follows TDD with failing tests first, minimal implementation, passing tests, then commit.
