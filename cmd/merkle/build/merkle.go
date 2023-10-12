package build

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func Merkle(leaves []string) ([]string, error) {

	var level [][32]byte
	for _, l := range leaves {
		if l == "" {
			continue
		}

		var bs []byte

		// Grouped leaves
		if strings.HasPrefix(l, "(") &&
			strings.HasSuffix(l, ")") {

			l = strings.TrimPrefix(l, "(")
			l = strings.TrimSuffix(l, ")")

			els := strings.Split(l, ",")

			group := bytes.Buffer{}
			for _, el := range els {
				h, err := hex.DecodeString(el)
				if err != nil {
					return nil, fmt.Errorf("unable to parse '%s': %w", l, err)
				}

				hash := sha256.Sum256(h)
				group.Write(hash[:])
			}

			bs = group.Bytes()
		} else if l == "<>" {
			bs = []byte{}
		} else {
			var err error
			bs, err = hex.DecodeString(l)
			if err != nil {
				return nil, fmt.Errorf("unable to parse '%s': %w", l, err)
			}

		}

		hash := sha256.Sum256(bs)
		level = append(level, hash)
	}

	var tree []string
	addLevel := func(level [][32]byte) {
		s := ""
		for i, l := range level {
			s += fmt.Sprintf("%x", l)
			if i < len(level)-1 {
				s += " "
			}
		}

		tree = append(tree, s)

	}

	addLevel(level)
	for len(level) > 1 {
		if len(level)%2 != 0 {
			return nil, fmt.Errorf("invalid number of leaves")
		}

		var nextLevel [][32]byte
		for i := 1; i < len(level); i += 2 {
			item := make([]byte, 64)
			copy(item[:32], level[i-1][:])
			copy(item[32:], level[i][:])
			hash := sha256.Sum256(item)

			nextLevel = append(nextLevel, hash)
		}

		level = nextLevel
		addLevel(level)
	}

	return tree, nil
}
