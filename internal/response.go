package internal

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Resolve produces the response for a matched request: status, headers,
// body, and the deadline at which the body should be written (used to
// honour rp.Delay). It renders the body through the template engine
// only if the template syntax is present.
// Resolve produces the response for a matched request: status, headers,
// body, and the deadline at which the body should be written (used to
// honour rp.Delay). It renders the body through the template engine
// only if the template syntax is present. The req argument is reserved
// for future per-request response shaping.
func Resolve(_ *http.Request, rp ResolvedPath, data TemplateData) (status int, headers http.Header, body []byte, deadline time.Time, err error) {
	headers = http.Header{}
	for k, v := range rp.Headers {
		headers.Set(k, v)
	}
	if headers.Get("Content-Type") == "" {
		headers.Set("Content-Type", "application/json")
	}

	tmplBody := rp.Template
	if strings.Contains(tmplBody, "{{") {
		rendered, rerr := RenderResponse(tmplBody, data)
		if rerr != nil {
			return 0, nil, nil, time.Time{}, fmt.Errorf("template: %w", rerr)
		}
		tmplBody = rendered
	}

	if rp.Delay != "" {
		d, perr := ParseDelay(rp.Delay)
		if perr != nil {
			return 0, nil, nil, time.Time{}, fmt.Errorf("parse delay: %w", perr)
		}
		if d > 0 {
			deadline = time.Now().Add(d)
		}
	}

	return rp.Status, headers, []byte(tmplBody), deadline, nil
}
