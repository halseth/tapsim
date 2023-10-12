package coinpool

import (
	"fmt"

	"github.com/halseth/mattlab/commitment"
	"github.com/halseth/mattlab/scripts/macros"
)

// stack:
// <root>
// <path> <pubkey1> <amt1>
// <path> <pubkey2> <amt2>
// ...
// <aggregate sig>
const script = `
OP_TOALTSTACK # merkle root to alt stack
%s # check size
OP_CAT # amt|pubkey
OP_0 #empty element for new leaf
OP_FROMALTSTACK # root from altstack
%s # amend merkle

# TODO: check input commitment against root


`

// <pub> <amt>
const checkSize = `
OP_2DUP
OP_SIZE
04 OP_EQUALVERIFY # amt must be 4 bytes
OP_DROP
OP_SIZE
20 OP_EQUALVERIFY # pubkey must be 32 bytes
OP_DROP
`

const toAltstack = `
OP_TOALTSTACK
`
const fromAltstack = `
OP_FROMALTSTACK
`
const swap = `
OP_SWAP
`
const rot = `
OP_ROT
`
const replaceLeaf = `
OP_SHA256 OP_SWAP OP_SHA256 OP_SWAP 
OP_CAT OP_0
`

const equalverify = `
OP_EQUALVERIFY
`
const empty = `
OP_0
`

const dup = `
OP_DUP
`

const dup2 = `
OP_2DUP
`
const dup3 = `
OP_3DUP
`
const drop = `
OP_DROP
`
const add = `
OP_ADD
`

const checksig = `
OP_CHECKSIGADD
`

const checkInput = `
OP_TOALTSTACK # input key to alt stack
OP_DUP # duplicate root
OP_0 # index
OP_FROMALTSTACK # input key
81 # current taptree
OP_1 # flags, check input
OP_CHECKCONTRACTVERIFY # check input commitment matches
`

const checkOutput = `
OP_0 # index
OP_FROMALTSTACK # inner key
81 # current taptree
OP_0 # flags, check output
OP_CHECKCONTRACTVERIFY # check output commitment matches
OP_TRUE
`

// stack:
// <input key>
// <root>
// <path> <pubkey1> <amt1>
// <path> <pubkey2> <amt2>
// ...
// <path> <pubkeyn> <amtn>
// <sig1>
// <sig2>
// ...
// <sign>
func LeafScript(numParticipants, numExitKeys int) string {

	numLevels := 1
	maxParticipants := 2
	for numParticipants > maxParticipants {
		numLevels++
		maxParticipants *= 2
	}

	s := ""

	// Input key to alt stack, will reuse for output.
	s += dup
	s += toAltstack

	// Check that the input data is according to contract.
	s += checkInput

	// Add our running amount to alt stack
	s += empty
	s += toAltstack

	for i := 0; i < numExitKeys; i++ {
		// on alt stack:
		// <running amt>
		// <exited pubkeys>
		// <input key>
		//
		// stack:
		// <root>
		// <amt>
		// <pubkey>
		// <merkle path>
		// ...
		// <signatures>

		// push <amt> <pubkey>  to alt stack
		s += dup3
		s += drop // drop root
		s += swap // swap so pubkey goes in the back

		s += fromAltstack // get running amount
		s += swap         // swap pubkey and running amount
		s += toAltstack   // push pubkey

		// add <amt> to running amount, push to alt stack
		s += add
		s += toAltstack

		// push root to alt stack
		s += toAltstack

		// we are replacing h(amt)|h(pub) with <>
		s += replaceLeaf

		// get root from alt stack
		s += fromAltstack

		// verify and replace leaf
		s += macros.AmendMerkle(numLevels)

		// new root on stack
	}

	// On stack:
	// <new root>
	// <signatures>
	//
	// alt stack:
	// <running amt>
	// <pub1> <pub2> ... <pubn>
	// <input key>

	// TODO: dropping running amt
	s += fromAltstack
	s += drop

	// Check signatures.
	s += empty
	for i := 0; i < numExitKeys; i++ {
		// move signature up the stack
		s += rot
		s += swap

		// get pubkey from alt stack
		s += fromAltstack

		// check signature
		s += checksig
	}

	n := commitment.ScriptNum(numExitKeys)
	s += fmt.Sprintf("%x", n.Bytes())

	s += equalverify

	// check output against new root
	// TODO: what do do with amount? can use deferred checks?

	s += checkOutput

	return s
}
