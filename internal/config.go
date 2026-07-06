// Package internal implements the core of json-server: config loading,
// request matching, template rendering, and HTTP routing.
package internal

// Config is the raw, on-disk form of a server config: server settings
// and the declared paths as they appear in YAML.
type Config struct {
	Paths  []Path `yaml:"paths"`
	Server Server `yaml:"server"`
}

// Path is one declared route, in its raw YAML form. The ConfigLoader
// resolves it into a ResolvedPath before it reaches the router.
type Path struct {
	Name             string            `yaml:"name"`
	Method           string            `yaml:"method"`
	Response         string            `yaml:"response"`
	ResponseAdvanced *ResponseAdvanced `yaml:"responseConfig,omitempty"`
	Query            map[string]string `yaml:"query,omitempty"`
	Body             *BodyMatch        `yaml:"body,omitempty"`
}

// ResponseAdvanced describes a path whose response is configured
// explicitly: body, status, delay, and headers.
type ResponseAdvanced struct {
	Body    string            `yaml:"body"`
	Status  int               `yaml:"status"`
	Delay   string            `yaml:"delay"`
	Headers map[string]string `yaml:"headers"`
}

// BodyMatch constrains a path to requests whose body contains a
// substring.
type BodyMatch struct {
	Contains string `yaml:"contains"`
}

// Server holds the per-server settings: port, host, optional CORS,
// optional logging, and default response headers.
type Server struct {
	Port           uint              `yaml:"port"`
	Host           string            `yaml:"host"`
	CORS           *CORSConfig       `yaml:"cors,omitempty"`
	Logging        *LoggingConfig    `yaml:"logging,omitempty"`
	DefaultHeaders map[string]string `yaml:"defaultHeaders,omitempty"`
}

// CORSConfig configures the CORS middleware.
type CORSConfig struct {
	Enabled bool     `yaml:"enabled"`
	Origins []string `yaml:"origins"`
	Headers []string `yaml:"headers"`
}

// LoggingConfig configures the access-log middleware.
type LoggingConfig struct {
	Enabled bool   `yaml:"enabled"`
	Format  string `yaml:"format"`
}
