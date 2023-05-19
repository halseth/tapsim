package script

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/txscript"
)

func Parse(script string) ([]byte, error) {
	c := strings.Split(script, " ")

	builder := txscript.NewScriptBuilder()
	for _, o := range c {

		// If valid opcode, simply push it to the script.
		if op, ok := txscript.OpcodeByName[o]; ok {
			builder.AddOp(op)
			continue
		}

		// Otherwise, try to interpret it as data.
		data, err := hex.DecodeString(o)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", o, err)
		}

		builder.AddData(data)
	}

	return builder.Script()
}

func ParseWitness(witness string) ([][]byte, error) {
	c := strings.Split(witness, " ")

	var witnessStack [][]byte
	for _, o := range c {
		var (
			data []byte
			err  error
		)
		switch o {
		// Empty element.
		case "<>":
			data = []byte{}

		default:
			data, err = hex.DecodeString(o)
			if err != nil {
				return nil, fmt.Errorf("parsing %s: %w", o, err)
			}
		}

		witnessStack = append(witnessStack, data)
	}

	return witnessStack, nil
}
