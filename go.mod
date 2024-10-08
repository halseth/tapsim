module github.com/halseth/tapsim

go 1.21.0

toolchain go1.21.2

require (
	github.com/btcsuite/btcd v0.23.5-0.20231215221805-96c9fd8078fd
	github.com/btcsuite/btcd/btcec/v2 v2.3.2
	github.com/davecgh/go-spew v1.1.1
	github.com/halseth/mattlab v0.0.0-20231006112235-a4d3fca1d564
	github.com/jessevdk/go-flags v1.4.0
	github.com/pkg/term v1.1.0
	github.com/urfave/cli/v2 v2.23.7
)

require (
	github.com/btcsuite/btcd/btcutil v1.1.5 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0 // indirect
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
)

replace github.com/btcsuite/btcd => github.com/halseth/btcd v0.0.0-20241008122125-a734f1460bff
