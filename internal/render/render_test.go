package render

import (
	"context"
	"errors"
	"testing"

	composeparser "github.com/bartosz121/compose-vault/internal/parser"
	"github.com/bartosz121/compose-vault/internal/vault"
	"github.com/goccy/go-yaml/ast"
)

type fakeSecretReader struct {
	data      map[string]map[string]any
	err       error
	readPaths []string
}

func (f *fakeSecretReader) ReadSecret(_ context.Context, path string) (map[string]any, error) {
	f.readPaths = append(f.readPaths, path)
	if f.err != nil {
		return nil, f.err
	}
	return f.data[path], nil
}

func TestResolveCachesSecretReadsByPath(t *testing.T) {
	t.Parallel()

	reader := &fakeSecretReader{
		data: map[string]map[string]any{
			"secret/data/app": {
				"username": "alice",
				"password": "secret",
			},
		},
	}

	matches := []composeparser.Match{
		{
			Node:        &ast.StringNode{Value: "vault:secret/data/app#username"},
			Placeholder: composeparser.Placeholder{Path: "secret/data/app", Field: "username"},
		},
		{
			Node:        &ast.StringNode{Value: "vault:secret/data/app#password"},
			Placeholder: composeparser.Placeholder{Path: "secret/data/app", Field: "password"},
		},
	}

	replacements, err := resolveWithReadSecret(context.Background(), reader.ReadSecret, matches)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reader.readPaths) != 1 || reader.readPaths[0] != "secret/data/app" {
		t.Fatalf("expected one read for secret/data/app, got %#v", reader.readPaths)
	}

	if len(replacements) != 2 {
		t.Fatalf("expected 2 replacements, got %d", len(replacements))
	}

	if replacements[0].Value != "alice" || replacements[1].Value != "secret" {
		t.Fatalf("unexpected replacements: %#v", replacements)
	}
}

func TestResolveReturnsFieldNotFound(t *testing.T) {
	t.Parallel()

	reader := &fakeSecretReader{
		data: map[string]map[string]any{
			"secret/data/app": {"username": "alice"},
		},
	}

	matches := []composeparser.Match{
		{
			Node:        &ast.StringNode{Value: "vault:secret/data/app#password"},
			Placeholder: composeparser.Placeholder{Path: "secret/data/app", Field: "password"},
		},
	}

	_, err := resolveWithReadSecret(context.Background(), reader.ReadSecret, matches)
	if !errors.Is(err, vault.ErrFieldNotFound) {
		t.Fatalf("expected ErrFieldNotFound, got %v", err)
	}
}

func TestResolveReturnsFieldNotString(t *testing.T) {
	t.Parallel()

	reader := &fakeSecretReader{
		data: map[string]map[string]any{
			"secret/data/app": {"password": 123},
		},
	}

	matches := []composeparser.Match{
		{
			Node:        &ast.StringNode{Value: "vault:secret/data/app#password"},
			Placeholder: composeparser.Placeholder{Path: "secret/data/app", Field: "password"},
		},
	}

	_, err := resolveWithReadSecret(context.Background(), reader.ReadSecret, matches)
	if !errors.Is(err, vault.ErrFieldNotString) {
		t.Fatalf("expected ErrFieldNotString, got %v", err)
	}
}

func TestResolvePropagatesReadError(t *testing.T) {
	t.Parallel()

	readErr := errors.New("vault unavailable")
	reader := &fakeSecretReader{err: readErr}

	matches := []composeparser.Match{
		{
			Node:        &ast.StringNode{Value: "vault:secret/data/app#password"},
			Placeholder: composeparser.Placeholder{Path: "secret/data/app", Field: "password"},
		},
	}

	_, err := resolveWithReadSecret(context.Background(), reader.ReadSecret, matches)
	if !errors.Is(err, readErr) {
		t.Fatalf("expected read error, got %v", err)
	}
}

func TestApplyMutatesNodes(t *testing.T) {
	t.Parallel()

	node := &ast.StringNode{Value: "vault:secret/data/app#password"}

	Apply([]Replacement{
		{Node: node, Value: "secret"},
	})

	if node.Value != "secret" {
		t.Fatalf("expected node value to be replaced, got %q", node.Value)
	}
}
