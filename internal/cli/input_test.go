package cli

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestParseYAMLInputFromFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := dir + "/compose.yaml"
	content := "services:\n  app:\n    image: nginx\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	file, source, err := parseYAMLInput(path, strings.NewReader(""), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if source != path {
		t.Fatalf("expected source %q, got %q", path, source)
	}

	if got := file.String(); got != content {
		t.Fatalf("unexpected file content: %q", got)
	}
}

func TestParseYAMLInputFromStdinWhenOmitted(t *testing.T) {
	t.Parallel()

	content := "services:\n  app:\n    image: nginx\n"

	file, source, err := parseYAMLInput("", strings.NewReader(content), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if source != "stdin" {
		t.Fatalf("expected stdin source, got %q", source)
	}

	if got := file.String(); got != content {
		t.Fatalf("unexpected file content: %q", got)
	}
}

func TestParseYAMLInputFromStdinWhenDash(t *testing.T) {
	t.Parallel()

	content := "services:\n  app:\n    image: nginx\n"

	file, source, err := parseYAMLInput("-", strings.NewReader(content), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if source != "stdin" {
		t.Fatalf("expected stdin source, got %q", source)
	}

	if got := file.String(); got != content {
		t.Fatalf("unexpected file content: %q", got)
	}
}

func TestParseYAMLInputRejectsMissingInteractiveInput(t *testing.T) {
	t.Parallel()

	_, _, err := parseYAMLInput("", strings.NewReader(""), true)
	if !errors.Is(err, ErrNoInput) {
		t.Fatalf("expected ErrNoInput, got %v", err)
	}
}

func TestParseYAMLInputRejectsEmptyPipedInput(t *testing.T) {
	t.Parallel()

	_, _, err := parseYAMLInput("", strings.NewReader("   \n"), false)
	if !errors.Is(err, ErrNoInput) {
		t.Fatalf("expected ErrNoInput, got %v", err)
	}
}
