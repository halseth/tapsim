package main

import (
	"fmt"
	"log"
	"os"

	"github.com/halseth/bitcoinscriptsim/script"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "bitcoinscriptsim",
		Usage: "parse bitcoin script",
	}

	app.Commands = []*cli.Command{
		{
			Name:        "parse",
			Usage:       "",
			UsageText:   "",
			Description: "",
			ArgsUsage:   "",
			Action:      parse,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "script",
					Usage: "script to parse",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func parse(cCtx *cli.Context) error {

	var scriptStr string
	if cCtx.NArg() > 0 {
		scriptStr = cCtx.Args().Get(0)
	} else if cCtx.String("script") != "" {
		scriptStr = cCtx.String("script")
	}

	parsed, err := script.Parse(scriptStr)
	if err != nil {
		return err
	}

	fmt.Printf("Parsed: %x\n", parsed)
	return nil
}
