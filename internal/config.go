package internal

type Config struct {
	Paths  []Path  `yaml:"paths"`
	Server Server `yaml:"server"`
}

type Path struct {
	Name              string            `yaml:"name"`
	Method            string            `yaml:"method"`
	Response          string            `yaml:"response"`
	ResponseAdvanced  *ResponseAdvanced `yaml:"responseConfig,omitempty"`
	Query             map[string]string `yaml:"query,omitempty"`
	Body              *BodyMatch        `yaml:"body,omitempty"`
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