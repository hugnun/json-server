package internal

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func Resolve(req *http.Request, rp ResolvedPath, data TemplateData) (int, http.Header, []byte, time.Time, error) {
	headers := http.Header{}
	for k, v := range rp.Headers {
		headers.Set(k, v)
	}
	if headers.Get("Content-Type") == "" {
		headers.Set("Content-Type", "application/json")
	}

	body := rp.Template
	if strings.Contains(body, "{{") {
		rendered, err := RenderResponse(body, data)
		if err != nil {
			return 0, nil, nil, time.Time{}, fmt.Errorf("template: %w", err)
		}
		body = rendered
	}

	var deadline time.Time
	if rp.Delay != "" {
		d, err := ParseDelay(rp.Delay)
		if err != nil {
			return 0, nil, nil, time.Time{}, fmt.Errorf("parse delay: %w", err)
		}
		if d > 0 {
			deadline = time.Now().Add(d)
		}
	}

	return rp.Status, headers, []byte(body), deadline, nil
}
