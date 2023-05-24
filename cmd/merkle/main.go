package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type config struct {
	Leaves string `short:"l" long:"leaves" description:"space separated string of hex values to commit to (must be power of 2)"`
}

var cfg = config{}

func main() {
	if _, err := flags.Parse(&cfg); err != nil {
		fmt.Println(err)
		return
	}

	err := run()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func run() error {
	leaves := strings.Split(cfg.Leaves, " ")

	var level [][32]byte
	for _, l := range leaves {
		if l == "" {
			continue
		}

		h, err := hex.DecodeString(l)
		if err != nil {
			return fmt.Errorf("unable to parse '%s': %w", l, err)
		}

		hash := sha256.Sum256(h)
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
			return fmt.Errorf("invalid number of leaves")
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

	for i := len(tree) - 1; i >= 0; i-- {
		fmt.Printf("%s\n", tree[i])
	}

	return nil
}
