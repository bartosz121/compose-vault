package parser

import (
	"errors"
	"strings"
	"testing"
)

func TestParsePlaceholder(t *testing.T) {
	cases := []struct {
		name        string
		value       string
		want        Placeholder
		wantErr     bool
		errContains string
	}{
		{
			"parses long nested path",
			"vault:a/b/c/d/e/f/g#h",
			Placeholder{
				Path:  "a/b/c/d/e/f/g",
				Field: "h",
			},
			false,
			"",
		},
		{
			"trims trailing whitespace",
			"vault:a/b/c/d/e/f/g#h         ",
			Placeholder{
				Path:  "a/b/c/d/e/f/g",
				Field: "h",
			},
			false,
			"",
		},
		{
			"parses kv v2 style path",
			"vault:secret/data#token",
			Placeholder{
				Path:  "secret/data",
				Field: "token",
			},
			false,
			"",
		},
		{
			"allows hash in field name",
			"vault:secret/data#token#abc", // TODO: is this allowed in vault?
			Placeholder{
				Path:  "secret/data",
				Field: "token#abc",
			},
			false,
			"",
		},
		{
			"rejects missing prefix",
			"secret/data#token",
			Placeholder{},
			true,
			"missing 'vault:' prefix",
		},
		{
			"rejects missing field separator",
			"vault:secret/data",
			Placeholder{},
			true,
			"missing field separator #",
		},
		{
			"rejects empty path",
			"vault:#password",
			Placeholder{},
			true,
			"empty path or field",
		},
		{
			"rejects empty field",
			"vault:secret/data#",
			Placeholder{},
			true,
			"empty path or field",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParsePlaceholder(tc.value)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, ErrInvalidPlaceholder) {
					t.Fatalf("expected ErrInvalidPlaceholder, got %v", err)
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Field != tc.want.Field || got.Path != tc.want.Path {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
