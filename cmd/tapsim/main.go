package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/halseth/tapsim/file"
	"github.com/halseth/tapsim/script"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "tapsim",
		Usage: "parse and debug bitcoin scripts",
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
		{
			Name:        "execute",
			Usage:       "",
			UsageText:   "",
			Description: "",
			ArgsUsage:   "",
			Action:      execute,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "script",
					Usage: "filename or output script as string",
				},
				&cli.StringFlag{
					Name:  "witness",
					Usage: "witness stack",
				},
				&cli.BoolFlag{
					Name:    "non-interactive",
					Aliases: []string{"ni"},
					Usage:   "disable interactive mode",
				},
				&cli.StringFlag{
					Name:  "inputkey",
					Usage: "use specified internal key for the input",
				},
				&cli.StringFlag{
					Name:  "outputkey",
					Usage: "use specified internal key for the output",
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

func execute(cCtx *cli.Context) error {
	var scriptFile, scriptStr, witnessStr string
	if cCtx.NArg() > 0 {
		scriptFile = cCtx.Args().Get(0)
	} else if cCtx.String("script") != "" {
		scriptFile = cCtx.String("script")
	}

	// Attempt to read the script from file.
	scriptBytes, err := file.Read(scriptFile)
	if err == nil {
		scriptStr, err = file.Parse(scriptBytes)
		if err != nil {
			return err
		}
	} else {
		// If we failed reading the file, assume it's the
		// script directly.
		scriptStr = scriptFile
	}

	if cCtx.NArg() > 1 {
		witnessStr = cCtx.Args().Get(1)
	} else if cCtx.String("witness") != "" {
		witnessStr = cCtx.String("witness")
	}

	nonInteractive := cCtx.Bool("non-interactive")
	inputKeyStr := cCtx.String("inputkey")
	inputKeyBytes, err := hex.DecodeString(inputKeyStr)
	if err != nil {
		return err
	}
	outputKeyStr := cCtx.String("outputkey")
	outputKeyBytes, err := hex.DecodeString(outputKeyStr)
	if err != nil {
		return err
	}

	fmt.Printf("Script: %s\r\n", scriptStr)
	fmt.Printf("Witness: %s\r\n", witnessStr)

	parsedScript, err := script.Parse(scriptStr)
	if err != nil {
		return err
	}

	parsedWitness, err := script.ParseWitness(witnessStr)
	if err != nil {
		return err
	}

	executeErr := script.Execute(
		inputKeyBytes, outputKeyBytes, parsedScript, parsedWitness, !nonInteractive,
	)
	if executeErr != nil {
		fmt.Printf("script exection failed: %s\r\n", executeErr)
		return nil
	}

	fmt.Printf("script verified\r\n")
	return nil
}
