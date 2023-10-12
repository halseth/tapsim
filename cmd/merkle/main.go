package main

import (
	"fmt"
	"strings"

	"github.com/halseth/tapsim/cmd/merkle/build"
	flags "github.com/jessevdk/go-flags"
)

type config struct {
	Leaves string `short:"l" long:"leaves" description:"space separated string of hex values to commit to (must be power of 2). To group more values in single leaf, use (val1,val2)"`
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

	tree, err := build.Merkle(leaves)
	if err != nil {
		return err
	}

	for i := len(tree) - 1; i >= 0; i-- {
		fmt.Printf("%s\n", tree[i])
	}

	return nil
}
