package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"text/template"
	"time"
)

// TemplateData is the data passed to the response template: path
// params, query values, and (if the matched route required it) the
// parsed request body.
type TemplateData struct {
	Params map[string]string
	Query  map[string]string
	Body   map[string]any
}

// RenderResponse parses body as a Go template with the {{json ...}}
// helper, then executes it with data. Parse and execute errors are
// wrapped with positional context.
func RenderResponse(body string, data TemplateData) (string, error) {
	funcMap := template.FuncMap{
		"json": func(v any) string {
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

// ParseDelay parses a Go duration string (e.g. "500ms", "2s"). The
// empty string yields a zero duration with no error.
func ParseDelay(delayStr string) (time.Duration, error) {
	if delayStr == "" {
		return 0, nil
	}
	return time.ParseDuration(delayStr)
}

// ParseBody decodes body as JSON into a map. If the JSON decode fails,
// it falls back to URL-encoded form parsing, returning the first key's
// values when a key has multiple entries.
func ParseBody(body string) (map[string]any, error) {
	var result map[string]any
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		values, perr := url.ParseQuery(body)
		if perr != nil {
			return nil, perr
		}
		result = make(map[string]any)
		for k, v := range values {
			result[k] = v
		}
	}
	return result, nil
}
