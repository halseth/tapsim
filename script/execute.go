package script

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/halseth/bitcoinscriptsim/output"
	"github.com/pkg/term"
)

const scriptFlags = txscript.StandardVerifyFlags

// Execute builds a tap leaf using the passed pkScript and executes it step by
// step with the provided witness.
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

	setupFunc := func() (*txscript.Engine, error) {
		sigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
		return txscript.NewEngine(
			prevOut.PkScript, txCopy, 0, scriptFlags,
			nil, sigHashes, prevOut.Value, prevOutFetcher,
		)
	}

	// Set the terminal in raw mode, such that we can capture arrow
	// presses.
	t, err := term.Open("/dev/tty")
	if err != nil {
		return err
	}
	defer t.Close()

	term.RawMode(t)

	currentStep := 0
	prevLines := 0
	bytes := make([]byte, 3)
	for {
		numRead, err := t.Read(bytes)
		if err != nil {
			return err
		}

		if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
			switch bytes[2] {
			case 65:
				//fmt.Print("Up arrow key pressed\r\n")
			case 66:
				//fmt.Print("Down arrow key pressed\r\n")
			case 67:
				//fmt.Print("Right arrow key pressed\r\n")
				currentStep++
			case 68:
				//fmt.Print("Left arrow key pressed\r\n")
				currentStep--
			}

		} else if numRead == 1 && bytes[0] == 3 {
			// Ctrl+C pressed, quit the program
			return nil
		}

		// Based on the current step counter, we execute up until that
		// step, then print the state table.
		table, done, err := StepScript(setupFunc, currentStep)
		if err != nil {
			return err
		}

		output.DrawTable(table, prevLines)
		if done {
			break
		}

		// Take note of the number of lines just printed, such that we
		// can clear them on next iteration.
		prevLines = strings.Count(table, "\n") + 1
	}

	return nil
}

var errAbortVM = fmt.Errorf("aborting vm execution")

func StepScript(setupFunc func() (*txscript.Engine, error), numSteps int) (string, bool, error) {

	vm, err := setupFunc()
	if err != nil {
		return "", false, err
	}

	const (
		SCRIPT_SCRIPTSIG      = 0
		SCRIPT_SCRIPTPUBKEY   = 1
		SCRIPT_WITNESS_SCRIPT = 2
	)

	var finalState string

	// Set up a callback that we will use to inspect the engine state at
	// every execution step.
	currentScript := -1
	stepCounter := 0
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
			if currentScript != step.ScriptIndex &&
				stepCounter == numSteps {
				finalState += "witness program verified OK\n"
			}
		}

		stepCounter++
		currentScript = step.ScriptIndex

		// Parse the current script for output.
		if stepCounter == numSteps {
			scriptStr := output.VmScriptToString(vm, step.ScriptIndex)
			table := output.ExecutionTable(
				step.OpcodeIndex,
				scriptStr,
				output.StackToString(step.Stack),
				output.StackToString(step.AltStack),
				output.StackToString(step.Witness),
			)

			finalState += table
			finalState += "\n"
		}

		// If we have executed enough steps, signal to abort using our
		// custom error.
		if stepCounter > numSteps {
			return errAbortVM
		}
		return nil
	}

	vmErr := vm.Execute()

	// An unexpected error was encountered.
	if vmErr != nil && vmErr != errAbortVM {
		return "", false, vmErr
	}

	done := vmErr == nil
	return finalState, done, nil
}
