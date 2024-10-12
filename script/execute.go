package script

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/halseth/tapsim/output"
	"github.com/pkg/term"
)

type TxOutput struct {
	OutputKey *btcec.PublicKey
	Value     int64
}

const scriptFlags = txscript.StandardVerifyFlags

// Execute builds a tap leaf using the passed pkScript and executes it step by
// step with the provided witness.
//
// privKeyBytes should map names of private keys given in the input witness to
// key bytes. An empty key will generate a random one.
//
// If [input/output]KeyBytes is empty, a random key will be generated.
func Execute(privKeyBytes map[string][]byte, inputKeyBytes []byte,
	outputs []TxOutput, pkScripts [][]byte, scriptIndex int,
	tx *wire.MsgTx, prevOuts []wire.TxOut, inputIdx int,
	witnessGen []WitnessGen, interactive, noStep bool, tags map[string]string, skipAhead int) error {

	generatedTx := false
	if tx == nil {
		tx = wire.NewMsgTx(2)
		generatedTx = true

		// add necessary amount of inputs
		for i := 0; i <= inputIdx; i++ {
			op := wire.OutPoint{
				Index: 0,
			}

			tx.AddTxIn(&wire.TxIn{
				PreviousOutPoint: op,
			})
		}

		for i, o := range outputs {
			fmt.Printf("output[%d] taproot key: %x:%d\n",
				i, schnorr.SerializePubKey(o.OutputKey), o.Value)

			outputScript, err := txscript.PayToTaprootScript(o.OutputKey)
			if err != nil {
				return err
			}

			tx.AddTxOut(&wire.TxOut{
				Value:    o.Value,
				PkScript: outputScript,
			})
		}
	}

	// Parse the input private keys.
	privKeys := make(map[string]*btcec.PrivateKey)
	for k, v := range privKeyBytes {
		var (
			key *btcec.PrivateKey
			err error
		)
		// If the key is empty, generate a random one.
		if len(v) == 0 {
			key, err = btcec.NewPrivateKey()
			if err != nil {
				return err
			}
		} else {
			key, _ = btcec.PrivKeyFromBytes(v)
		}
		privKeys[k] = key
	}

	var inputKey *btcec.PublicKey
	if len(inputKeyBytes) == 0 {
		privKey, err := btcec.NewPrivateKey()
		if err != nil {
			return err
		}

		inputKey = privKey.PubKey()
	} else {
		var err error
		inputKey, err = schnorr.ParsePubKey(inputKeyBytes)
		if err != nil {
			return err
		}
	}

	if len(outputs) == 0 {
		privKey, err := btcec.NewPrivateKey()
		if err != nil {
			return err
		}

		outputKey := privKey.PubKey()
		outputs = append(outputs, TxOutput{outputKey, 1e8})
	}

	var inputScript []byte
	var tapLeaf txscript.TapLeaf
	var ctrlBlock txscript.ControlBlock

	if generatedTx {
		var tapLeaves []txscript.TapLeaf

		for i, pkScript := range pkScripts {
			t := txscript.NewBaseTapLeaf(pkScript)
			tapLeaves = append(tapLeaves, t)

			if i == scriptIndex {
				tapLeaf = t
			}
		}

		tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaves...)

		ctrlBlock = tapScriptTree.LeafMerkleProofs[scriptIndex].ToControlBlock(
			inputKey,
		)

		tapScriptRootHash := tapScriptTree.RootNode.TapHash()

		inputTapKey := txscript.ComputeTaprootOutputKey(
			inputKey, tapScriptRootHash[:],
		)

		var err error
		inputScript, err = txscript.PayToTaprootScript(inputTapKey)
		if err != nil {
			return err
		}

		fmt.Printf("taptree: %x\n", tapScriptRootHash[:])
		fmt.Printf("input internal key: %x\n", schnorr.SerializePubKey(inputKey))
		fmt.Printf("input taproot key: %x\n", schnorr.SerializePubKey(inputTapKey))
	}

	prevOutsMap := make(map[wire.OutPoint]*wire.TxOut)
	if prevOuts == nil {
		for _, in := range tx.TxIn {
			prev := &wire.TxOut{
				Value:    1e8,
				PkScript: inputScript,
			}
			prevOutsMap[in.PreviousOutPoint] = prev
		}
	} else {
		for i := range prevOuts {
			o := prevOuts[i]
			outpoint := tx.TxIn[i].PreviousOutPoint
			prevOutsMap[outpoint] = &o
		}
	}

	prevOutFetcher := txscript.NewMultiPrevOutFetcher(prevOutsMap)

	sigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
	signFunc := func(keyID string) ([]byte, error) {
		privKey, ok := privKeys[keyID]
		if !ok {
			return nil, fmt.Errorf("private key %s not known", keyID)
		}
		prevOut := prevOutsMap[tx.TxIn[inputIdx].PreviousOutPoint]
		return txscript.RawTxInTapscriptSignature(
			tx, sigHashes, inputIdx, prevOut.Value, prevOut.PkScript, tapLeaf,
			txscript.SigHashDefault, privKey,
		)
	}

	var combinedWitness wire.TxWitness
	for _, gen := range witnessGen {
		w, err := gen(signFunc)
		if err != nil {
			return err
		}

		combinedWitness = append(combinedWitness, w)
	}

	txCopy := tx.Copy()
	if generatedTx {
		ctrlBlockBytes, err := ctrlBlock.ToBytes()
		if err != nil {
			return err
		}

		combinedWitness = append(combinedWitness, pkScripts[scriptIndex], ctrlBlockBytes)
		txCopy.TxIn[inputIdx].Witness = combinedWitness
	}

	return ExecuteTx(txCopy, prevOutsMap, inputIdx, interactive, noStep, tags, skipAhead)
}

func ExecuteTx(tx *wire.MsgTx,
	prevOuts map[wire.OutPoint]*wire.TxOut, inputIdx int,
	interactive, noStep bool, tags map[string]string, skipAhead int) error {

	currentInput := prevOuts[tx.TxIn[inputIdx].PreviousOutPoint]

	prevOutFetcher := txscript.NewMultiPrevOutFetcher(prevOuts)

	setupFunc := func(cb func(*txscript.StepInfo) error) (*txscript.Engine, error) {
		sigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
		return txscript.NewDebugEngine(
			currentInput.PkScript, tx, inputIdx, scriptFlags,
			nil, sigHashes, currentInput.Value, prevOutFetcher,
			cb,
		)
	}

	var t *term.Term
	var err error
	if interactive {
		// Set the terminal in raw mode, such that we can capture arrow
		// presses.
		t, err = term.Open("/dev/tty")
		if err != nil {
			return err
		}
		defer t.Close()

		term.RawMode(t)
		defer t.Restore()
	}

	currentStep := 1
	prevLines := 0
	bytes := make([]byte, 3)

	// We'll start script execution and control the stepping by signalling
	// on a channel.
	stepChan := make(chan error, 1)
	tableChan, errChan := StepScript(
		setupFunc, stepChan, tx.TxIn[inputIdx].Witness, tags, currentStep,
	)

	for {
		// Always start by signalling a step.
		stepChan <- nil

		var table string
		var vmErr error
		select {
		case vmErr = <-errChan:
		case table = <-tableChan:
		}

		// Before handling any error, we draw the state table for the
		// step.
		clearLines := 0
		if interactive {
			clearLines = prevLines
		}

		if !noStep {
			output.DrawTable(table, clearLines)
		}
		if interactive {
			if currentStep > 1 {
				fmt.Printf("Script execution: \u2190 back | next \u2192 ")
			} else {
				fmt.Printf("Script execution: next \u2192 ")
			}
		}

		// Take note of the number of lines just printed, such that we
		// can clear them on next iteration in case we are using
		// interactive mode.
		prevLines = strings.Count(table, "\n") + 1

		// If the VM encountered no error, it means the script
		// successfully executed to completion.
		if table == "" && vmErr == nil {
			output.ClearLines(1)
			return nil
		}

		// If we encountered an error other than errAbortVM,
		// the script actually failed.
		if table == "" && vmErr != errAbortVM {
			output.ClearLines(1)
			return vmErr
		}

		// Otherwise script execution was aborted before it completed,
		// so we continue with the next step of the execution.

		if interactive && currentStep >= skipAhead {
			skipAhead = 0
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
					if currentStep < 1 {
						currentStep = 1
					}

					// We are stepping backwards. Since we
					// have not really optimized for this
					// direction, we'll just start a new VM
					// and have it execute up until the
					// current step.
					tableChan, errChan = StepScript(
						setupFunc, stepChan,
						tx.TxIn[inputIdx].Witness, tags,
						currentStep,
					)
				}

			} else if numRead == 1 && bytes[0] == 3 {
				// Ctrl+C pressed, quit the program
				output.ClearLines(1)
				return fmt.Errorf("execution aborted")
			}
		} else {
			currentStep++
		}
	}
}

var errAbortVM = fmt.Errorf("aborting vm execution")

func StepScript(setupFunc func(func(*txscript.StepInfo) error) (*txscript.Engine, error),
	stepChan <-chan error, witness [][]byte,
	tags map[string]string, numSteps int) (<-chan string, <-chan error) {

	var (
		vm  *txscript.Engine
		err error
	)

	const (
		SCRIPT_SCRIPTSIG      = 0
		SCRIPT_SCRIPTPUBKEY   = 1
		SCRIPT_WITNESS_SCRIPT = 2
	)

	// We'll send outut for each step, or if we encounter an error, on
	// these channels.
	outputChan := make(chan string)
	errChan := make(chan error, 1)

	// Set up a callback that we will use to inspect the engine state at
	// every execution step.
	var (
		currentScript = -1
		stepCounter   = 0
		finalState    string
	)
	stepCallback := func(step *txscript.StepInfo) error {
		finalState = ""
		var showWitness [][]byte

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

			showWitness = witness

		// Execution of the witness script is the interesting part.
		case SCRIPT_WITNESS_SCRIPT:
			if currentScript <= step.ScriptIndex {
				finalState += "witness program verified OK\n"
			}
		}

		stepCounter++
		currentScript = step.ScriptIndex

		// If we haven't reached the number of steps to execute, we'll
		// return here to allow the VM to continue execution.
		if stepCounter < numSteps {
			return nil
		}

		// Parse the current script for output.
		scriptStr := output.VmScriptToString(vm, step.ScriptIndex)
		table := output.ExecutionTable(
			step.OpcodeIndex,
			scriptStr,
			output.StackToString(step.Stack),
			output.StackToString(step.AltStack),
			output.StackToString(showWitness),
			tags,
		)

		finalState += table
		finalState += "\n"

		// Now that we have executed enough steps, send the resulting
		// output over the channel
		outputChan <- finalState

		// Now wait for a signal to either continue execution or exit.
		stepErr := <-stepChan
		if stepErr != nil {
			return stepErr
		}

		return nil
	}

	vm, err = setupFunc(stepCallback)
	if err != nil {
		errChan <- err
		return outputChan, errChan
	}

	// We wont' block on execution, but start the VM in a goroutine, such
	// that the caller can step through it.
	go func() {
		vmErr := vm.Execute()
		errChan <- vmErr
	}()

	return outputChan, errChan
}
