package script

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/halseth/tapsim/output"
	"github.com/pkg/term"
)

// Validate builds a tap leaf using the passed pkScript and executes it step by
// step with the provided witness.
func Validate(tx *wire.MsgTx,
	prevOuts []wire.TxOut, inputIdx int, interactive, noStep bool, tags map[string]string,
	skipAhead int) error {
	// Check that the number of prevOuts matches the number of inputs.
	if len(prevOuts) != len(tx.TxIn) {
		return fmt.Errorf("number of prevouts does not match number of inputs")
	}

	// Check that the input index is within bounds.
	if inputIdx >= len(tx.TxIn) {
		return fmt.Errorf("input index out of range")
	}

	prevOutsMap := make(map[wire.OutPoint]*wire.TxOut)
	for i := range prevOuts {
		o := prevOuts[i]
		outpoint := tx.TxIn[i].PreviousOutPoint
		prevOutsMap[outpoint] = &o
	}

	prevOutFetcher := txscript.NewMultiPrevOutFetcher(prevOutsMap)

	setupFunc := func(cb func(*txscript.StepInfo) error) (*txscript.Engine, error) {
		sigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
		return txscript.NewDebugEngine(
			prevOuts[inputIdx].PkScript, tx, inputIdx, scriptFlags,
			nil, sigHashes, prevOuts[inputIdx].Value, prevOutFetcher,
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
