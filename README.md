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
./tapsim execute --script "OP_DUP OP_HASH160 79510b993bd0c642db233e2c9f3d9ef0d653f229 OP_EQUALVERIFY" --witness "OP_4"
Script: OP_DUP OP_HASH160 79510b993bd0c642db233e2c9f3d9ef0d653f229 OP_EQUALVERIFY
Witness: OP_4
------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 script                                  | stack                                   | alt stack                               | witness
------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 OP_DUP                                  | 54                                      |                                         |
 OP_HASH160                              |                                         |                                         |
 OP_DATA_20 0x79510b993bd0c642db233e2c...|                                         |                                         |
 OP_EQUALVERIFY                          |                                         |                                         |
>                                        |                                         |                                         |
------------------------------------------------------------------------------------------------------------------------------------------------------------------------


script verified
```

## Contributing
Contributions to Tapsim are welcomed. Please open a pull request or issue.

## License
Tapsim is licensed under the MIT License - see the LICENSE.md file for details.
