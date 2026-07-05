package internal

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestResolve_PlainBody(t *testing.T) {
	rp := ResolvedPath{
		Name:     "users",
		Status:   200,
		Template: `[{"id":1}]`,
		Headers:  map[string]string{},
	}
	req := httptest.NewRequest("GET", "/users", nil)
	status, hdr, body, deadline, err := Resolve(req, rp, TemplateData{})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if status != 200 {
		t.Errorf("status = %d", status)
	}
	if string(body) != `[{"id":1}]` {
		t.Errorf("body = %q", body)
	}
	if !deadline.IsZero() {
		t.Errorf("deadline = %v, want zero", deadline)
	}
	if hdr.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q", hdr.Get("Content-Type"))
	}
}

func TestResolve_DefaultStatus(t *testing.T) {
	rp := ResolvedPath{
		Status:   0,
		Template: "ok",
		Headers:  map[string]string{},
	}
	_, _, _, _, err := Resolve(httptest.NewRequest("GET", "/", nil), rp, TemplateData{})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestResolve_TemplatedBody(t *testing.T) {
	rp := ResolvedPath{
		Name:     "users/{id}",
		Status:   200,
		Template: `{"id":"{{.Params.id}}"}`,
		Headers:  map[string]string{},
	}
	data := TemplateData{Params: map[string]string{"id": "42"}}
	_, _, body, _, err := Resolve(httptest.NewRequest("GET", "/users/42", nil), rp, data)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if string(body) != `{"id":"42"}` {
		t.Errorf("body = %q", body)
	}
}

func TestResolve_TemplateError(t *testing.T) {
	rp := ResolvedPath{
		Status:   200,
		Template: `{{.NoSuchField.deep}}`,
		Headers:  map[string]string{},
	}
	_, _, _, _, err := Resolve(httptest.NewRequest("GET", "/", nil), rp, TemplateData{})
	if err == nil {
		t.Fatal("expected error for bad template")
	}
}

func TestResolve_DeadlineFromDelay(t *testing.T) {
	rp := ResolvedPath{
		Status:   200,
		Template: "ok",
		Delay:    "100ms",
		Headers:  map[string]string{},
	}
	before := time.Now()
	_, _, _, deadline, err := Resolve(httptest.NewRequest("GET", "/", nil), rp, TemplateData{})
	after := time.Now()
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if deadline.IsZero() {
		t.Fatal("deadline not set")
	}
	if deadline.Before(before.Add(50*time.Millisecond)) || deadline.After(after.Add(200*time.Millisecond)) {
		t.Errorf("deadline %v out of expected window [%v, %v]", deadline, before, after)
	}
}

func TestResolve_BadDelay(t *testing.T) {
	rp := ResolvedPath{
		Status:   200,
		Template: "ok",
		Delay:    "not-a-duration",
		Headers:  map[string]string{},
	}
	_, _, _, _, err := Resolve(httptest.NewRequest("GET", "/", nil), rp, TemplateData{})
	if err == nil {
		t.Fatal("expected error for bad delay")
	}
}

func TestResolve_Headers(t *testing.T) {
	rp := ResolvedPath{
		Status:   200,
		Template: "ok",
		Headers:  map[string]string{"X-Custom": "y", "Content-Type": "text/plain"},
	}
	_, hdr, _, _, err := Resolve(httptest.NewRequest("GET", "/", nil), rp, TemplateData{})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if hdr.Get("X-Custom") != "y" {
		t.Errorf("X-Custom = %q", hdr.Get("X-Custom"))
	}
	if hdr.Get("Content-Type") != "text/plain" {
		t.Errorf("Content-Type = %q want overridden", hdr.Get("Content-Type"))
	}
}

func TestResolve_EmptyBodyNoTemplate(t *testing.T) {
	rp := ResolvedPath{
		Status:   200,
		Template: "static text",
		Headers:  map[string]string{},
	}
	_, _, body, _, err := Resolve(httptest.NewRequest("GET", "/", nil), rp, TemplateData{})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(string(body), "static text") {
		t.Errorf("body = %q", body)
	}
}
