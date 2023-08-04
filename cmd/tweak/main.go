package main

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/txscript"
	"github.com/halseth/tapsim/file"
	"github.com/halseth/tapsim/script"
	flags "github.com/jessevdk/go-flags"
)

type config struct {
	Key    string `short:"k" long:"key" description:"key to use (random if empty)"`
	Script string `long:"script" description:"script or script file"`
	Merkle string `long:"merkle" description:"merkle commitment"`
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
	var scriptStr string
	scriptBytes, err := file.Read(cfg.Script)
	if err == nil {
		scriptStr, err = file.ParseScript(scriptBytes)
		if err != nil {
			return err
		}
	} else {
		// If we failed reading the file, assume it's the
		// script directly.
		scriptStr = cfg.Script
	}

	merkleBytes, err := file.Read(cfg.Merkle)
	if err != nil {
		// If we failed reading the file, assume it's hex encode merkle
		// root.
		var err error
		merkleBytes, err = hex.DecodeString(cfg.Merkle)
		if err != nil {
			return err
		}
	}

	pkScript, err := script.Parse(scriptStr)
	if err != nil {
		return err
	}

	tapLeaf := txscript.NewBaseTapLeaf(pkScript)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)
	tapScriptRootHash := tapScriptTree.RootNode.TapHash()

	// Random key.
	var keyBytes []byte
	if cfg.Key == "" {
		privKey, err := btcec.NewPrivateKey()
		if err != nil {
			return err
		}
		pubKey := privKey.PubKey()
		keyBytes = schnorr.SerializePubKey(pubKey)
	} else {
		var err error
		keyBytes, err = hex.DecodeString(cfg.Key)
		if err != nil {
			return err
		}

	}
	fmt.Println("inner internal key:", hex.EncodeToString(keyBytes))
	fmt.Println("taproot:", hex.EncodeToString(tapScriptRootHash[:]))
	fmt.Println("merkle root:", hex.EncodeToString(merkleBytes[:]))

	pubKey, err := schnorr.ParsePubKey(keyBytes)
	if err != nil {
		return err
	}

	// Tweak pubkey with data.
	tweaked := txscript.ComputeTaprootOutputKey(pubKey, merkleBytes[:])
	tweakedBytes := schnorr.SerializePubKey(tweaked)
	fmt.Println("tweaked(merkle):", hex.EncodeToString(tweakedBytes))

	tweaked2 := txscript.ComputeTaprootOutputKey(tweaked, tapScriptRootHash[:])
	tweakedBytes2 := schnorr.SerializePubKey(tweaked2)
	fmt.Println("taproot output key(merkle+taproot):", hex.EncodeToString(tweakedBytes2))

	empty := []byte{}
	merkleOut := txscript.ComputeTaprootOutputKey(tweaked, empty)
	merkleOutBytes := schnorr.SerializePubKey(merkleOut)
	fmt.Println("taproot output key(merkle), no script:", hex.EncodeToString(merkleOutBytes))

	emptyOut := txscript.ComputeTaprootOutputKey(pubKey, empty)
	emptyOutBytes := schnorr.SerializePubKey(emptyOut)
	fmt.Println("taproot output key, no tweak no script:", hex.EncodeToString(emptyOutBytes))

	return nil
}
