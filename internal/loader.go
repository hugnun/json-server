package internal

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Priority is the structural class of a ResolvedPath, used by the
// router to order match attempts. Most specific first.
type Priority int

// Priority classes, ordered most-specific to least-specific.
const (
	PriorityExact Priority = iota
	PriorityParam
	PriorityQuery
	PriorityBody
)

// String returns the lower-case name of the priority class.
func (p Priority) String() string {
	switch p {
	case PriorityExact:
		return "exact"
	case PriorityParam:
		return "param"
	case PriorityQuery:
		return "query"
	case PriorityBody:
		return "body"
	default:
		return "unknown"
	}
}

// ResolvedPath is the normalised form of a Path produced by the
// ConfigLoader. The Path/PathAdvanced ambiguity is resolved here and
// the priority class is computed from structural shape.
type ResolvedPath struct {
	Name     string
	Method   string
	Status   int
	Headers  map[string]string
	Delay    string
	Template string
	Query    map[string]string
	BodyRule *BodyMatch
	Priority Priority
}

// LoadedConfig is a parsed config together with the resolved routes
// derived from it.
type LoadedConfig struct {
	Server   Server         `yaml:"server"`
	Resolved []ResolvedPath `yaml:"-"`
}

// LoadConfig reads and parses a YAML config file from the given path,
// resolving each declared path into a ResolvedPath with its priority class
// assigned.
func LoadConfig(path string) (LoadedConfig, error) {
	// #nosec G304 -- path is a user-supplied CLI argument by design.
	data, err := os.ReadFile(path)
	if err != nil {
		return LoadedConfig{}, fmt.Errorf("read config: %w", err)
	}
	var raw Config
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return LoadedConfig{}, fmt.Errorf("parse YAML: %w", err)
	}

	resolved := make([]ResolvedPath, 0, len(raw.Paths))
	for _, p := range raw.Paths {
		rp, err := resolve(p)
		if err != nil {
			return LoadedConfig{}, fmt.Errorf("path %q: %w", p.Name, err)
		}
		rp.Priority = classify(rp)
		resolved = append(resolved, rp)
	}

	return LoadedConfig{Server: raw.Server, Resolved: resolved}, nil
}

func resolve(p Path) (ResolvedPath, error) {
	rp := ResolvedPath{
		Name:     p.Name,
		Method:   p.Method,
		Status:   http.StatusOK,
		Headers:  map[string]string{},
		Query:    p.Query,
		BodyRule: p.Body,
	}

	switch {
	case p.Response != "":
		if isInlineJSON(p.Response) {
			rp.Template = p.Response
		} else {
			data, err := os.ReadFile(p.Response)
			if err != nil {
				return rp, fmt.Errorf("read response file %q: %w", p.Response, err)
			}
			rp.Template = string(data)
		}
	case p.ResponseAdvanced != nil:
		adv := p.ResponseAdvanced
		rp.Template = adv.Body
		if adv.Status != 0 {
			rp.Status = adv.Status
		}
		if adv.Delay != "" {
			rp.Delay = adv.Delay
		}
		for k, v := range adv.Headers {
			rp.Headers[k] = v
		}
	}

	return rp, nil
}

func isInlineJSON(s string) bool {
	return strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[")
}

func classify(rp ResolvedPath) Priority {
	if rp.BodyRule != nil {
		return PriorityBody
	}
	if len(rp.Query) > 0 {
		return PriorityQuery
	}
	if hasParam(rp.Name) {
		return PriorityParam
	}
	return PriorityExact
}

func hasParam(name string) bool {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			return true
		}
	}
	return false
}
