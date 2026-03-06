package cli

import (
	"errors"
	"os"

	composeparser "github.com/bartosz121/compose-vault/internal/parser"
)

type CheckCmd struct {
	File string `arg:"" optional:"" name:"file" help:"Path to the YAML file to validate. Reads stdin when omitted or set to '-'."`
}

func (c *CheckCmd) Run(globals *Globals) error {
	file, source, err := parseYAMLInput(c.File, os.Stdin, stdinIsTerminal(os.Stdin))
	if err != nil {
		return err
	}

	_, matchErrors := composeparser.FindMatches(file)
	if len(matchErrors) == 0 {
		return nil
	}

	return errors.New(joinMatchErrors(source, matchErrors))
}
