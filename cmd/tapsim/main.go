package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/wire"

	"github.com/halseth/tapsim/file"
	"github.com/halseth/tapsim/output"
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
					Name:  "scripts",
					Usage: "list of filenames with output scripts to assemble into a taptree",
				},
				&cli.IntFlag{
					Name:  "scriptindex",
					Usage: "index of script from \"scripts\" to execute",
				},
				&cli.StringFlag{
					Name:  "tx",
					Usage: "serialized transaction in hex",
				},
				&cli.StringFlag{
					Name:  "prevouts",
					Usage: "serialized prevouts comma seperated. Must be in same order as tx inputs",
				},
				&cli.IntFlag{
					Name:  "inputindex",
					Usage: "index of input from \"tx\" to execute",
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
				&cli.BoolFlag{
					Name:    "no-step",
					Aliases: []string{"ns"},
					Usage:   "don't show step by step, just validate",
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
					Name:  "outputs",
					Usage: "specify taproot outputs as \"<pubkey>:<value>\"",
				},

				&cli.StringFlag{
					Name:  "tagfile",
					Usage: "optional json file map from hex values to human-readable tags",
				},
				&cli.IntFlag{
					Name:  "colwidth",
					Usage: "output column width",
					Value: output.ColumnWidth,
				},
				&cli.IntFlag{
					Name:  "rows",
					Usage: "max rows to print in execution table",
					Value: output.MaxRows,
				},
				&cli.IntFlag{
					Name:  "skip",
					Usage: "skip ahead",
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

	colWidth := cCtx.Int("colwidth")
	if colWidth > 0 {
		output.ColumnWidth = colWidth
	}
	maxRows := cCtx.Int("rows")
	if maxRows > 0 {
		output.MaxRows = maxRows
	}

	skipAhead := cCtx.Int("skip")

	outputKeyStr := cCtx.String("outputkey")
	outputsStr := cCtx.String("outputs")

	if len(outputKeyStr) > 0 && len(outputsStr) > 0 {
		return fmt.Errorf("cannot set both outputkey and outputs")
	}

	if len(outputKeyStr) > 0 {
		outputsStr = fmt.Sprintf("%s:100000000", outputKeyStr)
	}

	outputs := strings.Split(outputsStr, ",")
	var txOutKeys []script.TxOutput
	for _, oStr := range outputs {
		if oStr == "" {
			continue
		}

		k := strings.Split(oStr, ":")
		pubKeyBytes, err := hex.DecodeString(k[0])
		if err != nil {
			return err
		}

		pubKey, err := schnorr.ParsePubKey(pubKeyBytes)
		if err != nil {
			return err
		}

		val, err := strconv.ParseInt(k[1], 10, 0)
		if err != nil {
			return err
		}

		txOutKeys = append(txOutKeys, script.TxOutput{
			OutputKey: pubKey,
			Value:     val,
		})
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

	nonInteractive := cCtx.Bool("non-interactive")
	noStep := cCtx.Bool("no-step")

	inputKeyStr := cCtx.String("inputkey")
	prevoutsStr := cCtx.String("prevouts")

	if inputKeyStr != "" && prevoutsStr != "" {
		return fmt.Errorf("cannot set both inputkey and prevouts")
	}

	inputKeyBytes, err := hex.DecodeString(inputKeyStr)
	if err != nil {
		return err
	}

	var prevOuts []*wire.TxOut
	for _, p := range strings.Split(prevoutsStr, ",") {
		if p == "" {
			prevOuts = append(prevOuts, nil)
			continue
		}

		b, err := hex.DecodeString(p)
		if err != nil {
			return err
		}

		txOut := wire.TxOut{}
		reader := bytes.NewReader(b)
		err = wire.ReadTxOut(reader, 0, 0, &txOut)
		if err != nil {
			return err
		}

		prevOuts = append(prevOuts, &txOut)
	}

	var scriptStr []string
	scriptFile := cCtx.String("script")
	scriptFiles := cCtx.String("scripts")
	txStr := cCtx.String("tx")

	nn := 0
	if scriptFile != "" {
		nn++
	}
	if scriptFiles != "" {
		nn++
	}
	if txStr != "" {
		nn++
	}
	if nn != 1 {
		return fmt.Errorf("must set single one of script, scripts or tx")
	}

	scriptIndex := cCtx.Int("scriptindex")
	inputIndex := cCtx.Int("inputindex")

	var tx *wire.MsgTx
	if scriptFile != "" {
		// Attempt to read the script from file.
		scriptBytes, err := file.Read(scriptFile)
		if err == nil {
			s, err := file.ParseScript(scriptBytes)
			if err != nil {
				return err
			}

			scriptStr = []string{s}
		} else {
			// If we failed reading the file, assume it's the
			// script directly.
			scriptStr = []string{scriptFile}
		}
	} else if scriptFiles != "" {
		for _, f := range strings.Split(scriptFiles, ",") {
			if f == "" {
				continue
			}

			scriptBytes, err := file.Read(f)
			if err != nil {
				return err
			}
			s, err := file.ParseScript(scriptBytes)
			if err != nil {
				return err
			}

			scriptStr = append(scriptStr, s)
		}
	} else if txStr != "" {
		b, err := hex.DecodeString(txStr)
		if err != nil {
			return err
		}
		reader := bytes.NewReader(b)
		tx = &wire.MsgTx{}
		err = tx.Deserialize(reader)
		if err != nil {
			return err
		}

		executeErr := script.ExecuteTx(
			tx, prevOuts, inputIndex, !nonInteractive,
			noStep, tags, skipAhead,
		)
		if executeErr != nil {
			fmt.Printf("script exection failed: %s\r\n", executeErr)
			return executeErr
		}

		fmt.Printf("tx execution verified\r\n")
		return nil
	} else {
		return fmt.Errorf("must specify tx or script")
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

	privKeys := strings.Split(cCtx.String("privkeys"), ",")
	keyMap := make(map[string][]byte)
	for _, privKeyStr := range privKeys {
		if privKeyStr == "" {
			continue
		}
		k := strings.Split(privKeyStr, ":")
		privKeyBytes, err := hex.DecodeString(k[1])
		if err != nil {
			return err
		}

		keyMap[k[0]] = privKeyBytes
	}

	fmt.Printf("Script: %s\r\n", scriptStr[scriptIndex])
	fmt.Printf("Witness: %s\r\n", witnessStr)

	var parsedScripts [][]byte
	for _, s := range scriptStr {
		parsedScript, err := script.Parse(s)
		if err != nil {
			return err
		}

		parsedScripts = append(parsedScripts, parsedScript)
	}

	parsedWitness, err := script.ParseWitness(witnessStr)
	if err != nil {
		return err
	}

	executeErr := script.Execute(
		keyMap, inputKeyBytes, txOutKeys, parsedScripts, scriptIndex,
		parsedWitness, !nonInteractive, noStep, tags, skipAhead,
	)
	if executeErr != nil {
		fmt.Printf("script exection failed: %s\r\n", executeErr)
		return executeErr
	}

	fmt.Printf("script execution verified\r\n")
	return nil
}
