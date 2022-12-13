package script

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/txscript"
)

func Parse(script string) ([]byte, error) {
	c := strings.Split(script, " ")

	builder := txscript.NewScriptBuilder()
	for _, o := range c {
		op, ok := txscript.OpcodeByName[o]
		if !ok {
			return nil, fmt.Errorf("invalid opcode: %s", o)
		}

		builder.AddOp(op)
	}

	return builder.Script()
}
