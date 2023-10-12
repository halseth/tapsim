package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/txscript"
	"github.com/halseth/mattlab/commitment"
	"github.com/halseth/tapsim/cmd/merkle/build"
	"github.com/halseth/tapsim/examples/matt/coinpool/v2"
	"github.com/halseth/tapsim/file"
	"github.com/halseth/tapsim/script"
)

func main() {
	const numParticipants = 4
	///	numParticipants, err := strconv.Atoi(os.Args[1])
	///	if err != nil {
	///		panic(err)
	///	}
	///	numExit, err := strconv.Atoi(os.Args[2])
	///	if err != nil {
	///		panic(err)
	///	}
	///	_ = numExit
	//fmt.Println(coinpool.LeafScript(numParticipants, numExit))

	if err := run(numParticipants); err != nil {
		panic(err)
	}
}

func run(num int) error {

	var keys []*btcec.PrivateKey
	fKeys, err := os.Create("coinpool_v2_keys.txt")
	if err != nil {
		return err
	}
	defer fKeys.Close()

	for i := 0; i < num; i++ {
		privKey, err := btcec.NewPrivateKey()
		if err != nil {
			return err
		}
		keys = append(keys, privKey)
		privKeyBytes := privKey.Serialize()
		pubKey := privKey.PubKey()
		pubKeyBytes := schnorr.SerializePubKey(pubKey)
		l := fmt.Sprintf("privkey%d:%x,pubkey%d:%x\n", i+1, privKeyBytes, i+1, pubKeyBytes)
		_, err = fKeys.Write([]byte(l))
		if err != nil {
			return err
		}
	}

	const baseBalance = 10_000
	var balances []commitment.ScriptNum
	for i := 0; i < num; i++ {
		bal := commitment.ScriptNum(baseBalance * (i + 1))
		balances = append(balances, bal)
	}

	var leaves []string
	for i := 0; i < num; i++ {
		pubKey := keys[i].PubKey()
		pubKeyBytes := schnorr.SerializePubKey(pubKey)
		l := fmt.Sprintf("(%x,%x)", balances[i].Bytes(), pubKeyBytes)
		leaves = append(leaves, l)
	}

	tree, err := build.Merkle(leaves)
	if err != nil {
		return err
	}

	keyBytes := txscript.BIP341_NUMS_POINT
	innerKey, err := schnorr.ParsePubKey(keyBytes)
	if err != nil {
		return err
	}

	fInnerKey, err := os.Create("coinpool_v2_innerkey.txt")
	if err != nil {
		return err
	}
	defer fKeys.Close()

	var tapLeaves []txscript.TapLeaf
	for i := 0; i < num; i++ {
		scr := coinpool.LeafScript(num, i+1)

		f, err := os.Create(fmt.Sprintf("coinpool_v2_%dof%dexit.txt", i+1, num))
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(scr))
		if err != nil {
			return err
		}
		f.Close()

		s, err := file.ParseScript([]byte(scr))
		if err != nil {
			return err
		}

		fmt.Println("parsing script", num, i+1)

		pkScript, err := script.Parse(s)
		if err != nil {
			return fmt.Errorf("error parsing script %s: %w", scr, err)
		}
		t := txscript.NewBaseTapLeaf(pkScript)
		tapLeaves = append(tapLeaves, t)
	}

	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaves...)
	tapScriptRootHash := tapScriptTree.RootNode.TapHash()
	fmt.Printf("tap root: %x\n", tapScriptRootHash[:])
	fmt.Printf("inner key: %x\n",
		schnorr.SerializePubKey(innerKey),
	)

	fmt.Println("merkle tree 1:")
	for i := len(tree) - 1; i >= 0; i-- {
		fmt.Printf("%s\n", tree[i])
	}
	for _, l := range leaves {
		fmt.Printf("%s ", l)
	}
	fmt.Println()

	rootStr := tree[len(tree)-1]
	root, err := hex.DecodeString(rootStr)
	if err != nil {
		return err
	}
	tweaked1 := txscript.SingleTweakPubKey(innerKey, root[:])
	tapKey1 := txscript.ComputeTaprootOutputKey(tweaked1, tapScriptRootHash[:])
	fmt.Printf("tweaked key1: %x tapkey1: %x\n",
		schnorr.SerializePubKey(tweaked1),
		schnorr.SerializePubKey(tapKey1),
	)

	_, err = fInnerKey.Write([]byte(fmt.Sprintf("%x\n",
		schnorr.SerializePubKey(tweaked1),
	)))
	if err != nil {
		return err
	}

	var witness []string
	witness = append(witness, fmt.Sprintf("%x", schnorr.SerializePubKey(innerKey)))
	witness = append(witness, rootStr)

	// First exiting.
	leaves[0] = "<>"
	tree, err = build.Merkle(leaves)
	if err != nil {
		return err
	}

	fmt.Println("merkle tree 2:")
	for i := len(tree) - 1; i >= 0; i-- {
		fmt.Printf("%s\n", tree[i])
	}
	for _, l := range leaves {
		fmt.Printf("%s ", l)
	}
	fmt.Println()
	root, err = hex.DecodeString(rootStr)
	if err != nil {
		return err
	}

	tweaked2 := txscript.SingleTweakPubKey(innerKey, root[:])
	tapKey2 := txscript.ComputeTaprootOutputKey(tweaked2, tapScriptRootHash[:])
	fmt.Printf("tweaked key2: %x tapkey2: %x\n",
		schnorr.SerializePubKey(tweaked2),
		schnorr.SerializePubKey(tapKey2),
	)

	witness = append(witness, fmt.Sprintf("%x", balances[0].Bytes()))

	witness = append(witness,
		fmt.Sprintf("%x", schnorr.SerializePubKey(keys[0].PubKey())),
	)
	witness = append(witness, "<>")

	witness = append(witness, fmt.Sprintf("%s", strings.Split(tree[0], " ")[1]))
	witness = append(witness, "<>")
	witness = append(witness, fmt.Sprintf("%s", strings.Split(tree[1], " ")[1]))

	// Third user exiting.
	leaves[2] = "<>"
	tree, err = build.Merkle(leaves)
	if err != nil {
		return err
	}

	fmt.Println("merkle tree 3:")
	for i := len(tree) - 1; i >= 0; i-- {
		fmt.Printf("%s\n", tree[i])
	}
	for _, l := range leaves {
		fmt.Printf("%s ", l)
	}
	fmt.Println()
	root, err = hex.DecodeString(tree[len(tree)-1])
	if err != nil {
		return err
	}
	tweaked3 := txscript.SingleTweakPubKey(innerKey, root[:])
	tapKey3 := txscript.ComputeTaprootOutputKey(tweaked3, tapScriptRootHash[:])
	fmt.Printf("tweaked key3: %x tapkey3: %x\n",
		schnorr.SerializePubKey(tweaked3),
		schnorr.SerializePubKey(tapKey3),
	)

	_, err = fInnerKey.Write([]byte(fmt.Sprintf("%x\n",
		schnorr.SerializePubKey(tapKey3),
	)))
	if err != nil {
		return err
	}

	witness = append(witness, fmt.Sprintf("%x", balances[2].Bytes()))
	witness = append(witness, fmt.Sprintf("%x", schnorr.SerializePubKey(keys[2].PubKey())))

	witness = append(witness, "<>")

	witness = append(witness, fmt.Sprintf("%s", strings.Split(tree[0], " ")[3]))
	witness = append(witness, "01")

	witness = append(witness, fmt.Sprintf("%s", strings.Split(tree[1], " ")[0]))

	witness = append(witness, "<sig:privkey3>")
	witness = append(witness, "<sig:privkey1>")

	wFile, err := os.Create("witness_2of4exit.txt")
	if err != nil {
		return err
	}
	defer wFile.Close()
	for i := len(witness) - 1; i >= 0; i-- {
		_, err := wFile.Write([]byte(witness[i] + "\n"))
		if err != nil {
			return err
		}
	}

	return nil
}
