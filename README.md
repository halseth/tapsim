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


script verified
```

## Contributing
Contributions to Tapsim are welcomed. Please open a pull request or issue.

## Acknowledgements
This project is heavily inspired by the excellent [btcdeb](https://github.com/bitcoin-core/btcdeb).

## License
Tapsim is licensed under the MIT License - see the LICENSE.md file for details.
