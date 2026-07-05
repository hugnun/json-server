package internal

import (
	"testing"
)

func TestRenderResponse(t *testing.T) {
	data := TemplateData{
		Params: map[string]string{"id": "123"},
		Query:  map[string]string{"name": "bob"},
		Body:   map[string]interface{}{"email": "test@example.com"},
	}

	tests := []struct {
		body string
		want string
	}{
		{`{"id": "{{.Params.id}}"}`, `{"id": "123"}`},
		{`{{.Query.name}}`, "bob"},
		{`{"email": "{{index .Body "email"}}"}`, `{"email": "test@example.com"}`},
	}

	for _, tt := range tests {
		t.Run(tt.body, func(t *testing.T) {
			got, err := RenderResponse(tt.body, data)
			if err != nil {
				t.Fatalf("RenderResponse() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("RenderResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDelay(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"100ms", false},
		{"1s", false},
		{"", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseDelay(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDelay() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
