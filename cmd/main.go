package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/bartosz121/compose-vault/internal/cli"
)

var version = "dev"

func main() {
	cli := cli.CLI{
		Globals: cli.Globals{
			Version: cli.VersionFlag(version),
		},
	}

	ctx := kong.Parse(&cli,
		kong.Name("compose-vault"),
		kong.Description("#TODO:"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Exit(func(code int) { os.Exit(code) }),
		kong.Vars{"version": version},
	)
	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
