package jsonfmt

import (
	"strings"
	"testing"
)

func TestMarshalDoesNotEscapeHTML(t *testing.T) {
	data, err := Marshal(map[string]any{
		"html": "<div>&</div>",
	}, "  ")
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	got := string(data)
	if strings.Contains(got, "\\u003c") || strings.Contains(got, "\\u003e") || strings.Contains(got, "\\u0026") {
		t.Fatalf("Marshal escaped html-sensitive characters: %s", got)
	}

	want := "{\n  \"html\": \"<div>&</div>\"\n}"
	if got != want {
		t.Fatalf("unexpected formatted json:\nwant:\n%s\ngot:\n%s", want, got)
	}
}
