package script

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/txscript"
)

func Parse(script string) ([]byte, error) {
	c := strings.Split(script, " ")

	var (
		// We'll not use the script builder for the actual script, as
		// it will automatically change data pushes to be minimal
		// (which we don't always want). We use it to sanity check the
		// parsed script
		parsed  []byte
		builder = txscript.NewScriptBuilder()
	)
	for _, o := range c {
		// Trim any leftover whitespace.
		o := strings.TrimSpace(o)
		if o == "" {
			continue
		}

		// If valid opcode, simply push it to the script.
		if op, ok := txscript.OpcodeByName[o]; ok {
			builder.AddOp(op)
			parsed = append(parsed, op)
			continue
		}

		// Empty element.
		if o == "<>" {
			builder.AddData([]byte{})
			parsed = append(parsed, txscript.OP_0)
			continue
		}

		// Otherwise, try to interpret it as data.
		data, err := hex.DecodeString(o)
		if err != nil {
			return nil, fmt.Errorf("parsing '%s': %w", o, err)
		}

		dataLen := len(data)
		if dataLen < txscript.OP_PUSHDATA1 {
			parsed = append(parsed, byte((txscript.OP_DATA_1-1)+dataLen))
		} else if dataLen <= 0xff {
			parsed = append(parsed, txscript.OP_PUSHDATA1, byte(dataLen))
		} else if dataLen <= 0xffff {
			buf := make([]byte, 2)
			binary.LittleEndian.PutUint16(buf, uint16(dataLen))
			parsed = append(parsed, txscript.OP_PUSHDATA2)
			parsed = append(parsed, buf...)
		} else {
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, uint32(dataLen))
			parsed = append(parsed, txscript.OP_PUSHDATA4)
			parsed = append(parsed, buf...)
		}

		// Append the actual data.
		parsed = append(parsed, data...)
		builder.AddData(data)
	}

	_, err := builder.Script()
	return parsed, err
}

// SignFunc should return a signature for the current input given the private
// key ID given as an argument.
type SignFunc func(string) ([]byte, error)

// WitnessGen returns an element to place on the witness stack.
type WitnessGen func(SignFunc) ([]byte, error)

// ParseWitness parses the given witness string and returns a slice of
// WitnessGen functions. Each function should provide the witness element at
// its index, given a function to optionally obtain a signature.
//
// Signatures can be created by <sig:privkeyid> in the witness string, which
// will attempt to produce a signature from the key with name privkeyid.
func ParseWitness(witness string) ([]WitnessGen, error) {
	if witness == "" {
		return nil, nil
	}

	c := strings.Split(witness, " ")

	var witnessGen []WitnessGen
	for _, o := range c {
		var (
			gen WitnessGen
		)
		switch {
		// Empty element.
		case o == "<>":
			gen = func(SignFunc) ([]byte, error) {
				return []byte{}, nil
			}

		// Signature.
		case strings.HasPrefix(o, "<sig:") &&
			strings.HasSuffix(o, ">"):
			suf, _ := strings.CutPrefix(o, "<sig:")
			key, _ := strings.CutSuffix(suf, ">")

			gen = func(sign SignFunc) ([]byte, error) {
				return sign(key)
			}

		default:
			data, err := hex.DecodeString(o)
			if err != nil {
				return nil, fmt.Errorf("parsing %s: %w", o, err)
			}

			gen = func(SignFunc) ([]byte, error) {
				return data, nil
			}
		}

		witnessGen = append(witnessGen, gen)
	}

	return witnessGen, nil
}
