package internal

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

type MatchResult int

const (
	Matched MatchResult = iota
	NoMatch
	BodyInvalid
)

func (r MatchResult) String() string {
	switch r {
	case Matched:
		return "matched"
	case NoMatch:
		return "no-match"
	case BodyInvalid:
		return "body-invalid"
	default:
		return "unknown"
	}
}

func Match(req *http.Request, rp ResolvedPath) (TemplateData, MatchResult) {
	if !matchMethod(rp.Method, req.Method) {
		return TemplateData{}, NoMatch
	}
	if !matchPath(rp.Name, req.URL.Path) {
		return TemplateData{}, NoMatch
	}
	if !matchQuery(rp.Query, req.URL.Query()) {
		return TemplateData{}, NoMatch
	}

	var bodyData map[string]interface{}
	if rp.BodyRule != nil {
		bodyStr, err := readBody(req)
		if err != nil {
			return TemplateData{}, BodyInvalid
		}
		if !strings.Contains(bodyStr, rp.BodyRule.Contains) {
			return TemplateData{}, NoMatch
		}
		bodyData, _ = ParseBody(bodyStr)
	}

	return TemplateData{
		Params: ExtractPathParams(rp.Name, req.URL.Path),
		Query:  gatherQuery(req.URL.Query()),
		Body:   bodyData,
	}, Matched
}

func matchMethod(want, got string) bool {
	return want == "" || want == got
}

func matchQuery(want map[string]string, got url.Values) bool {
	for k, v := range want {
		if got.Get(k) != v {
			return false
		}
	}
	return true
}

func gatherQuery(values url.Values) map[string]string {
	q := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			q[k] = v[0]
		}
	}
	return q
}

func readBody(req *http.Request) (string, error) {
	if req.Body == nil {
		return "", nil
	}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}
	if strings.HasPrefix(path, "/") && !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")
	if len(patternParts) != len(pathParts) {
		return false
	}
	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		if part != pathParts[i] {
			return false
		}
	}
	return true
}

func ExtractPathParams(pattern, path string) map[string]string {
	params := make(map[string]string)
	if strings.HasPrefix(path, "/") && !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
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
