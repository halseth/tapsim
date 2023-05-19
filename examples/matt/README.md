### MATT
Example using opcodes OP_CAT, OP_CHECKINPUTCONTRACT and OP_CHECKOUTPUTCONTRACT.

### Usage
```bash
./tapsim execute --script examples/matt/script.txt  --witness examples/matt/witness.txt --tagfile examples/matt/tags.json --inputkey "04cb5a1bc1f576b90405274bb123d798cd9df47e85085648c8ba00299bd29427" --outputkey "9d756824ce50f52914818bae4ca5283c2fa500e89d3ad063d2dcf443c84859ce"
```

Input and output keys need to be specified as these are the tweaked keys where
the embedded data is committed.
