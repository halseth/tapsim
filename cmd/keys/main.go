package main

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	flags "github.com/jessevdk/go-flags"
)

const usage = "Returns n keys on the format 'privkey,pubkey'"

type config struct {
	Num int `short:"n" long:"num" description:"number of keys to generate"`
}

var cfg = config{}

func main() {
	parser := flags.NewParser(&cfg, flags.Default)
	parser.Usage = usage
	if _, err := parser.Parse(); err != nil {
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
	if cfg.Num < 1 {
		return fmt.Errorf("number of keys mus be positive")
	}

	for i := 0; i < cfg.Num; i++ {
		privKey, err := btcec.NewPrivateKey()
		if err != nil {
			return err
		}
		privKeyBytes := privKey.Serialize()
		pubKey := privKey.PubKey()
		pubKeyBytes := schnorr.SerializePubKey(pubKey)
		fmt.Printf("%x,%x\n", privKeyBytes, pubKeyBytes)
	}
	return nil
}
