### MATT
Example using opcodes OP_CAT, OP_CHECKINPUTCONTRACT and OP_CHECKOUTPUTCONTRACT.

### Usage
```bash
./tapsim execute --script examples/matt/script.txt  --witness examples/matt/witness.txt --tagfile examples/matt/tags.json --inputkey "961ab014008a4dd5bb860ca0914f360c8b3167eb661abf513337ecbae1c6cf81" --outputkey "fb2d8d7c90b81ca8f250fbe24a88b5751fecb45d4459079793d673036f22b704"
```

Input and output keys need to be specified as these are the tweaked keys where
the embedded data is committed.
