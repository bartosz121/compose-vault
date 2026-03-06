package parser

import (
	"errors"
	"fmt"
	"strings"
)

const PlaceholderPrefix = "vault:"

var ErrInvalidPlaceholder = errors.New("invalid vault placeholder")

type Placeholder struct {
	Path  string
	Field string
}

// ParsePlaceholder validates and parses an inline Vault placeholder in
// `vault:<path>#<field>` format into path and field components.
func ParsePlaceholder(raw string) (Placeholder, error) {
	value := strings.TrimSpace(raw)
	if !strings.HasPrefix(value, PlaceholderPrefix) {
		return Placeholder{}, fmt.Errorf("%w: missing 'vault:' prefix", ErrInvalidPlaceholder)
	}

	trimmed := strings.TrimPrefix(value, PlaceholderPrefix)
	parts := strings.SplitN(trimmed, "#", 2)
	if len(parts) != 2 {
		return Placeholder{}, fmt.Errorf("%w: missing field separator #", ErrInvalidPlaceholder)
	}

	path := strings.TrimSpace(parts[0])
	field := strings.TrimSpace(parts[1])
	if path == "" || field == "" {
		return Placeholder{}, fmt.Errorf("%w: empty path or field", ErrInvalidPlaceholder)
	}

	return Placeholder{
		Path:  path,
		Field: field,
	}, nil
}

func IsPlaceholderCandidate(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), PlaceholderPrefix)
}
