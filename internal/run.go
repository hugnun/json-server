package internal

import (
	"fmt"
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
				data, err := os.ReadFile(path.Response)
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
			if len(v) > 0 {
				query[k] = v[0]
			}
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