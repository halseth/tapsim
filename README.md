# Tapsim: Bitcoin Tapscript Debugger

Tapsim is a simple tool built in Go for debugging Bitcoin Tapscript
transactions. It's aimed at developers wanting play with Bitcoin script
primitives, aid in script debugging, and visualize the VM state as scripts are
executed.

## Description
Tapsim hooks into the [btcd](https://github.com/btcsuite/btcd) script execution
engine to retrieve state at every step of script execution.

The script execution is controlled using the left/right arrow keys.

Currently visualized during execution:
- Script
- Stack
- Altstack
- Witness stack

## Installation

Before installing Tapsim, please ensure you have the latest version of Go (Go
1.20 or later) installed on your computer.

```bash
git clone https://github.com/halseth/tapsim.git
cd tapsim
go build ./cmd/tapsim
```

### Tools
Also found in the `cmd` folder are a few tools useful for certain script
releated tasks:
- `keys`: generates random key pairs
- `merkle`: builds merkle trees
- `scriptnum`: convert to and from the Bitcoin CScriptNum format
- `tweak`: tweak public keys with data and taproot

## Usage
```bash
$ ./tapsim -h
NAME:
   tapsim - parse and debug bitcoin scripts

USAGE:
   tapsim [global options] command [command options] [arguments...]

COMMANDS:
   parse
   execute
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

## Examples
```bash
$ ./tapsim execute --script "OP_HASH160 79510b993bd0c642db233e2c9f3d9ef0d653f229 OP_EQUAL" --witness "54"
Script: OP_HASH160 79510b993bd0c642db233e2c9f3d9ef0d653f229 OP_EQUAL
Witness: 54
------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 script                                  | stack                                   | alt stack                               | witness
------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 OP_HASH160                              | 01                                      |                                         |
 OP_DATA_20 0x79510b993bd0c642db23...f229|                                         |                                         |
 OP_EQUAL                                |                                         |                                         |
>                                        |                                         |                                         |
------------------------------------------------------------------------------------------------------------------------------------------------------------------------


script execution verified
```

## Options
```bash
$ ./tapsim execute -h
NAME:
   tapsim execute

USAGE:
   tapsim execute [command options] [arguments...]

OPTIONS:
   --script value           filename or output script as string
   --scripts value          list of filenames with output scripts to assemble into a taptree
   --scriptindex value      index of script from "scripts" to execute (default: 0)
   --witness value          filename or witness stack as string
   --non-interactive, --ni  disable interactive mode (default: false)
   --no-step, --ns          don't show step by step, just validate (default: false)
   --privkeys value         specify private keys as "key1:<hex>,key2:<hex>" to sign the transaction. Set <hex> empty to generate a random key with the given ID.
   --inputkey value         use specified internal key for the input
   --outputkey value        use specified internal key for the output
   --outputs value          specify taproot outputs as "<pubkey>:<value>"
   --tagfile value          optional json file map from hex values to human-readable tags
   --colwidth value         output column width (default: 40)
   --rows value             max rows to print in execution table (default: 25)
   --skip value             skip aheead (default: 0)
   --help, -h               show help (default: false)
```

## Additional script features
In addition to the regular Bitcoin tapscript opcodes, tapsim has added support
for scripts using
- OP_CAT
- OP_CHECKCONTRACTVERIFY

## Contributing
Contributions to Tapsim are welcomed. Please open a pull request or issue.

## Acknowledgements
This project is heavily inspired by the excellent [btcdeb](https://github.com/bitcoin-core/btcdeb).

## License
Tapsim is licensed under the MIT License - see the LICENSE.md file for details.
