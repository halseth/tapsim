package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

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
					Usage: "filename or witness stack as string",
				},
				&cli.BoolFlag{
					Name:    "non-interactive",
					Aliases: []string{"ni"},
					Usage:   "disable interactive mode",
				},
				&cli.StringFlag{
					Name:  "privkeys",
					Usage: "specify private keys as \"key1:<hex>,key2:<hex>\" to sign the transaction. Set <hex> empty to generate a random key with the given ID.",
				},

				&cli.StringFlag{
					Name:  "inputkey",
					Usage: "use specified internal key for the input",
				},
				&cli.StringFlag{
					Name:  "outputkey",
					Usage: "use specified internal key for the output",
				},
				&cli.StringFlag{
					Name:  "tagfile",
					Usage: "optional json file map from hex values to human-readable tags",
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
	var scriptFile, scriptStr string
	if cCtx.NArg() > 0 {
		scriptFile = cCtx.Args().Get(0)
	} else if cCtx.String("script") != "" {
		scriptFile = cCtx.String("script")
	}

	// Attempt to read the script from file.
	scriptBytes, err := file.Read(scriptFile)
	if err == nil {
		scriptStr, err = file.ParseScript(scriptBytes)
		if err != nil {
			return err
		}
	} else {
		// If we failed reading the file, assume it's the
		// script directly.
		scriptStr = scriptFile
	}

	var witnessFile, witnessStr string
	if cCtx.NArg() > 1 {
		witnessFile = cCtx.Args().Get(1)
	} else if cCtx.String("witness") != "" {
		witnessFile = cCtx.String("witness")
	}

	// Attempt to read the witness from file.
	witnessBytes, err := file.Read(witnessFile)
	if err == nil {
		witnessStr, err = file.ParseScript(witnessBytes)
		if err != nil {
			return err
		}
	} else {
		// If we failed reading the file, assume it's the
		// witness directly.
		witnessStr = witnessFile
	}

	nonInteractive := cCtx.Bool("non-interactive")
	privKeys := strings.Split(cCtx.String("privkeys"), ",")
	keyMap := make(map[string][]byte)
	for _, privKeyStr := range privKeys {
		k := strings.Split(privKeyStr, ":")
		privKeyBytes, err := hex.DecodeString(k[1])
		if err != nil {
			return err
		}

		keyMap[k[0]] = privKeyBytes
	}

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

	tagFile := cCtx.String("tagfile")
	var tags map[string]string
	if tagFile != "" {
		tagBytes, err := file.Read(tagFile)
		if err != nil {
			return err
		}

		tags, err = file.ParseTagMap(tagBytes)
		if err != nil {
			return err
		}
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
		keyMap, inputKeyBytes, outputKeyBytes, parsedScript,
		parsedWitness, !nonInteractive, tags,
	)
	if executeErr != nil {
		fmt.Printf("script exection failed: %s\r\n", executeErr)
		return nil
	}

	fmt.Printf("script verified\r\n")
	return nil
}
