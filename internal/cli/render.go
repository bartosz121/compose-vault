package cli

import (
	"context"
	"errors"
	"io"
	"os"

	composeparser "github.com/bartosz121/compose-vault/internal/parser"
	"github.com/bartosz121/compose-vault/internal/render"
	"github.com/bartosz121/compose-vault/internal/vault"
)

type RenderCmd struct {
	File       string `arg:"" optional:"" name:"file" help:"Path to the YAML file to render. Reads stdin when omitted or set to '-'."`
	VaultAddr  string `name:"vault-addr" env:"VAULT_ADDR" help:"Vault address. Defaults to VAULT_ADDR when set."`
	VaultToken string `name:"vault-token" env:"VAULT_TOKEN" help:"Vault token. Defaults to VAULT_TOKEN when set."`
}

func (c *RenderCmd) Run(globals *Globals) error {
	file, source, err := parseYAMLInput(c.File, os.Stdin, stdinIsTerminal(os.Stdin))
	if err != nil {
		return err
	}

	matches, matchErrors := composeparser.FindMatches(file)
	if len(matchErrors) > 0 {
		return errors.New(joinMatchErrors(source, matchErrors))
	}

	if len(matches) == 0 {
		if err := writeRenderedYAML(os.Stdout, file); err != nil {
			return err
		}
		return nil
	}

	client, err := vault.NewClient(c.VaultAddr, c.VaultToken)
	if err != nil {
		return err
	}

	replacements, err := render.Resolve(context.Background(), client, matches)
	if err != nil {
		return err
	}

	render.Apply(replacements)

	if err := writeRenderedYAML(os.Stdout, file); err != nil {
		return err
	}

	return nil
}

func writeRenderedYAML(w io.Writer, file interface{ String() string }) error {
	_, err := io.WriteString(w, file.String())
	return err
}
