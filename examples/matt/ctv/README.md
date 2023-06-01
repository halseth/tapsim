### CHECKTEMPLATEVERIFY (OP_CTV)
Example using opcodes `OP_CAT`, `OP_CHECKINPUTCONTRACT` and
`OP_CHECKOUTPUTCONTRACT` to achieve a CTV-style covenant.

### Introduction
CTV lets and output encode a static set of further spending transactions,
meaning only a known set of spends are valid for the output.

In this example we will encode such a valid spending tx using the hypothetical
opcodes mentioned above. Note that one could create an unlimited amount of
valid spends by encoding each in a distinct tapleaf.

## Usage
Compile the tools needed (assuming in root folder):
```bash
go build -v ./cmd/keys
go build -v ./cmd/merkle
go build -v ./cmd/tweak
go build -v ./cmd/tapsim
```

We will assume a spending transaction that spends the output into two new
(static) outputs. We start by generating the keys.
```bash
./keys -n 2 > keys.txt
```

We'll now build a merkle tree that commits to valid output keys and the amount
for each. For simplicity say we have the 32-bit hex amounts

```bash
echo "00010000\n00020000" > amounts.txt
```

Each leaf in the merkle tree will be the public key concatenated with the
32-bit amount.

```bash
cat keys.txt | awk -F',' '{print $2}' | paste -d "\0" - amounts.txt | tr "\n" " " > merkle_leaves.txt
```

Now create a merkle tree from these values.
```bash
./merkle -l "`cat merkle_leaves.txt`" > merkle_tree.txt
```

Now we can simulate a UTXO that commits to this merkle root, and that way
encode the valid outpus. Note that if we wanted more than one valid spending
transaction, we would have to create a merkle root for each, then combine the
subtrees into a final merkle tree committing to each of them each.

The UTXO will be a taproot output encumbered by the tapscript in
[script.txt](script.txt). The keyspend path of the output will be a NUMS key,
since we don't want the keyspend path to be valid (we will just create a random
key for this example).

The way the output commits to the embedded data is by tweaking the "inner
internal key" with the merkle root. 

```bash
./tweak --merkle "`head -n 1 merkle_tree.txt`" > tweaks.txt
```

Now we have everything needed to simulate a spend of the output. 

The witness will have the following structure:

```
<output amount at index 1>
<output key at index 1>
<output amount at index 0>
<output key at index 0>
<inner internal key of input>
<merkle commitment of input>
```

We can generate the witness for Tapsim like this:

```bash
sed -n 2p balances.txt > witness.txt
sed -n 2p keys.txt | awk -F',' '{print $2}' >> witness.txt
sed -n 1p balances.txt >> witness.txt
sed -n 1p keys.txt | awk -F',' '{print $2}' >> witness.txt
sed -n 1p tweaks.txt | awk -F": " '{print $2}' >> witness.txt
sed -n 1p merkle_tree.txt | awk -F" " '{print $1}' >> witness.txt
```

The outputs of the spending transaction must match exactly what is specified in
the witness, and the input spend script enforces that the witness is exactly
what is committed in the embedded data. This effectively creates the CTV style
covenant.

To simplify debugging in Tapsim, we can also create a tag file, that adds human
readable names to the various data blobs:

```bash
echo "{}" > tags.json
cat tags.json | jq ". += {\"`sed -n 1p tweaks.txt | awk -F": " '{print $2}'`\":\"inner internal key\"}" > tags.json
cat tags.json | jq ". += {\"`sed -n 7p tweaks.txt | awk -F": " '{print $2}'`\":\"input internal key\"}" > tags.json

cat tags.json | jq ". += {\"`sed -n 1p merkle_tree.txt | awk -F" " '{print $1}'`\":\"input merkle root\"}" > tags.json
cat tags.json | jq ". += {\"`sed -n 2p merkle_tree.txt | awk -F" " '{print $1}'`\":\"input merkle[1][0]\"}" > tags.json
cat tags.json | jq ". += {\"`sed -n 2p merkle_tree.txt | awk -F" " '{print $2}'`\":\"input merkle[1][1]\"}" > tags.json

cat tags.json | jq ". += {\"`sed -n 1p keys.txt | awk -F"," '{print $2}'`\":\"output key[0]\"}" > tags.json
cat tags.json | jq ". += {\"`sed -n 2p keys.txt | awk -F"," '{print $2}'`\":\"output key[1]\"}" > tags.json

cat tags.json | jq ". += {\"`sed -n 1p amounts.txt`\":\"output amount[0]\"}" > tags.json
cat tags.json | jq ". += {\"`sed -n 2p amounts.txt`\":\"output amount[1]\"}" > tags.json
```

## Executing the script
Now we have all the pieces gathered to run the script.

```bash
./tapsim execute --script examples/matt/ctv/script.txt  --witness witness.txt --tagfile tags.json --inputkey "`sed -n 7p tweaks.txt | awk -F": " '{print $2}'`" --outputs="`cat keys.txt | awk -F',' '{print $2}' | paste -d ":" - amounts.txt | tr "\n" ","`"
```

### Explanation
The script will check that the spending transaction has the correct outputs.
Note that more inputs and outputs can be added to the transaction without
invalidating the witness (this could be disallowed if one wanted to),

TODO(halseth): 
- add value verification
- make example with more than one valid spending transction.

More detailed commentary can be found in [script.txt](script.txt).
