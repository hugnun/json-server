package internal

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// Build constructs the HTTP handler for cfg: a Router that owns all
// resolved routes, wrapped in the configured CORS, logging, and
// default-headers middleware.
func Build(cfg LoadedConfig) (http.Handler, error) {
	router := NewRouter()
	for _, rp := range cfg.Resolved {
		if err := router.Add(rp); err != nil {
			return nil, fmt.Errorf("add route %q: %w", rp.Name, err)
		}
	}

	var handler http.Handler = router

	if cfg.Server.CORS != nil && cfg.Server.CORS.Enabled {
		handler = corsMiddleware(handler, cfg.Server.CORS)
	}
	if cfg.Server.Logging != nil && cfg.Server.Logging.Enabled {
		handler = loggingMiddleware(handler, cfg.Server.Logging)
	}
	if len(cfg.Server.DefaultHeaders) > 0 {
		handler = defaultHeadersMiddleware(handler, cfg.Server.DefaultHeaders)
	}

	return handler, nil
}

// Run is the cobra RunE for the serve command. It loads the config at
// args[0], applies the --port override if set, builds the handler, and
// starts the HTTP server. ListenAndServe failures are returned.
func Run(cmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig(args[0])
	if err != nil {
		return err
	}
	if port, _ := cmd.Flags().GetInt("port"); port > 0 {
		cfg.Server.Port = uint(port)
	}

	handler, err := Build(cfg)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Starting server on %s\n", addr)
	// #nosec G114 -- dev tool: no Read/Write timeouts needed; controlled shutdown is out of scope.
	return http.ListenAndServe(addr, handler)
}

func corsMiddleware(next http.Handler, config *CORSConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		origin := req.Header.Get("Origin")
		if len(config.Origins) > 0 && (config.Origins[0] == "*" || contains(config.Origins, origin)) {
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

func loggingMiddleware(next http.Handler, _ *LoggingConfig) http.HandlerFunc {
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
