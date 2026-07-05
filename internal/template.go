package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
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
