package render

import (
	"context"
	"fmt"

	composeparser "github.com/bartosz121/compose-vault/internal/parser"
	"github.com/bartosz121/compose-vault/internal/vault"
	"github.com/goccy/go-yaml/ast"
)

type Replacement struct {
	Node  *ast.StringNode
	Value string
}

type secretReadFunc func(context.Context, string) (map[string]any, error)

func Resolve(ctx context.Context, client *vault.Client, matches []composeparser.Match) ([]Replacement, error) {
	return resolveWithReadSecret(ctx, client.ReadSecret, matches)
}

func resolveWithReadSecret(ctx context.Context, readSecret secretReadFunc, matches []composeparser.Match) ([]Replacement, error) {
	cache := make(map[string]map[string]any)
	replacements := make([]Replacement, 0, len(matches))

	for _, match := range matches {
		data, ok := cache[match.Placeholder.Path]
		if !ok {
			var err error
			data, err = readSecret(ctx, match.Placeholder.Path)
			if err != nil {
				return nil, err
			}
			cache[match.Placeholder.Path] = data
		}

		value, ok := data[match.Placeholder.Field]
		if !ok {
			return nil, fmt.Errorf("%w: %s#%s", vault.ErrFieldNotFound, match.Placeholder.Path, match.Placeholder.Field)
		}

		stringValue, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s#%s", vault.ErrFieldNotString, match.Placeholder.Path, match.Placeholder.Field)
		}

		replacements = append(replacements, Replacement{
			Node:  match.Node,
			Value: stringValue,
		})
	}

	return replacements, nil
}

func Apply(replacements []Replacement) {
	for _, replacement := range replacements {
		replacement.Node.Value = replacement.Value
	}
}
