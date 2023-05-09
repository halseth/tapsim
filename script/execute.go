package script

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/halseth/bitcoinscriptsim/output"
)

const scriptFlags = txscript.StandardVerifyFlags

// Execute builds a tap leaf using the passed pkScript and executes it with the
// provided witness.
func Execute(pkScript, witness []byte) error {
	// Get random key as we will use for the taproot internal key.
	privKey, err := btcec.NewPrivateKey()
	if err != nil {
		return err
	}
	internalKey := privKey.PubKey()

	tapLeaf := txscript.NewBaseTapLeaf(pkScript)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)

	ctrlBlock := tapScriptTree.LeafMerkleProofs[0].ToControlBlock(
		internalKey,
	)

	tapScriptRootHash := tapScriptTree.RootNode.TapHash()
	outputKey := txscript.ComputeTaprootOutputKey(
		internalKey, tapScriptRootHash[:],
	)
	p2trScript, err := txscript.PayToTaprootScript(outputKey)
	if err != nil {
		return err
	}

	tx := wire.NewMsgTx(2)
	tx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			Index: 0,
		},
	})
	prevOut := &wire.TxOut{
		Value:    1e8,
		PkScript: p2trScript,
	}

	prevOutFetcher := txscript.NewCannedPrevOutputFetcher(
		prevOut.PkScript, prevOut.Value,
	)

	ctrlBlockBytes, err := ctrlBlock.ToBytes()
	if err != nil {
		return err
	}

	txCopy := tx.Copy()
	txCopy.TxIn[0].Witness = wire.TxWitness{
		witness, pkScript, ctrlBlockBytes,
	}

	sigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
	vm, err := txscript.NewEngine(
		prevOut.PkScript, txCopy, 0, scriptFlags,
		nil, sigHashes, prevOut.Value, prevOutFetcher,
	)
	if err != nil {
		return err
	}

	const (
		SCRIPT_SCRIPTSIG      = 0
		SCRIPT_SCRIPTPUBKEY   = 1
		SCRIPT_WITNESS_SCRIPT = 2
	)

	// Set up a callback that we will use to inspect the engine state at
	// every execution step.
	currentScript := -1
	vm.StepCallback = func(step *txscript.StepInfo) error {
		switch step.ScriptIndex {
		// Script sig is empty and uninteresting under segwit, so we
		// just ignore it.
		case SCRIPT_SCRIPTSIG:
			currentScript = step.ScriptIndex
			return nil

		// The scriptpubkey contains the witness program and is used to
		// verify the script in the provided witness.
		case SCRIPT_SCRIPTPUBKEY:
			// Since to real script execution is done during the
			// script pubkey (only checking the witness program),
			// we will only output the step the first time we
			// encounter this script index.
			if currentScript == step.ScriptIndex {
				return nil
			}

		// Execution of the witness script is the interesting part.
		case SCRIPT_WITNESS_SCRIPT:
			if currentScript != step.ScriptIndex {
				fmt.Println("witness program verified OK")
			}
		}

		currentScript = step.ScriptIndex

		// Parse the current script for output.
		scriptStr := output.VmScriptToString(vm, step.ScriptIndex)
		table := output.ExecutionTable(
			step.OpcodeIndex,
			scriptStr,
			output.StackToString(step.Stack),
			output.StackToString(step.AltStack),
			output.StackToString(step.Witness),
		)
		fmt.Println(table)
		return nil
	}

	return vm.Execute()
}
