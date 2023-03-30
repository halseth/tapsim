package script

import (
	"encoding/hex"
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
			return nil, err
		}

		builder.AddData(data)
	}

	return builder.Script()
}
