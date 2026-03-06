package cli

import (
	"fmt"
	"strings"

	composeparser "github.com/bartosz121/compose-vault/internal/parser"
)

func formatMatchError(path string, matchErr composeparser.MatchError) string {
	position := matchErr.Node.GetToken().Position
	return fmt.Sprintf("%s:%d:%d: %s", path, position.Line, position.Column, strings.TrimSpace(matchErr.Err.Error()))
}

func joinMatchErrors(path string, matchErrors []composeparser.MatchError) string {
	var builder strings.Builder
	for idx, matchErr := range matchErrors {
		if idx > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(formatMatchError(path, matchErr))
	}

	return builder.String()
}
