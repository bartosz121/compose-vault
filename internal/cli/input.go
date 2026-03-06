package cli

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/goccy/go-yaml/ast"
	yamlparser "github.com/goccy/go-yaml/parser"
)

var ErrNoInput = errors.New("no input provided: pass a file path or pipe YAML on stdin")

func parseYAMLInput(path string, stdin io.Reader, stdinIsTerminal bool) (*ast.File, string, error) {
	path = strings.TrimSpace(path)
	if path != "" && path != "-" {
		file, err := yamlparser.ParseFile(path, 0)
		if err != nil {
			return nil, "", err
		}
		return file, path, nil
	}

	if stdinIsTerminal {
		return nil, "", ErrNoInput
	}

	input, err := io.ReadAll(stdin)
	if err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(string(input)) == "" {
		return nil, "", ErrNoInput
	}

	file, err := yamlparser.ParseBytes(input, 0)
	if err != nil {
		return nil, "", err
	}

	return file, "stdin", nil
}

// stdinIsTerminal reports whether stdin is attached to an interactive terminal
// so commands can fail fast instead of blocking while waiting for piped input
func stdinIsTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}
