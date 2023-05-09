package output

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/txscript"
)

func StackToString(stack [][]byte) []string {
	var str []string
	for i := len(stack) - 1; i >= 0; i-- {
		b := stack[i]
		s := hex.EncodeToString(b)
		str = append(str, s)
	}

	return str
}

func ScriptToString(script []byte) []string {
	str, _ := txscript.DisasmString(script)
	return strings.Split(str, " ")
}

func VmScriptToString(vm *txscript.Engine, scriptIdx int) []string {
	str, _ := vm.DisasmScript(scriptIdx)

	var ss []string
	// Trim prefix.
	l := len("01:0000: ")
	for _, s := range strings.Split(str, "\n") {
		if s == "" {
			continue
		}

		ss = append(ss, s[l:])
	}

	return ss
}

func WitnessToString(witness [][]byte) []string {
	// TODO: reverse order?
	var str []string
	for _, b := range witness {
		s, _ := txscript.DisasmString(b)
		str = append(str, s)
	}

	return str
}
