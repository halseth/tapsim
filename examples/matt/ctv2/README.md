### CHECKTEMPLATEVERIFY (OP_CTV)
Example using opcodes `OP_CHECKCONTRACTVERIFY` to achieve a CTV-style covenant
with two valid spending transactions.

### Introduction
CTV let us encode a static set of further spending transactions, meaning only a
known set of spends are valid for the output.

In this example we will encode two spending transactions using the hypothetical
opcodes mentioned above, that are both valid spends of the same input. Note
that this could be generalized to an arbitrary number of spending transactions.

## Usage
Compile the tools needed (assuming in root folder):
```bash
go build -v ./cmd/keys
go build -v ./cmd/tapsim
```

We assume one spending transaction that spends the output into two new (static)
outputs, and a second transaction that sends all funds to a single output. We
start by generating keys for the two transactions.

```bash
./keys -n 1 > keys1.txt
./keys -n 2 > keys2.txt
```

We'll now build a tapscript tree that commits to valid output keys and the
amount for each. For simplicity say we have the 32-bit hex amounts:
(TODO(halseth): opcodes don't actually enforce output values for now)

```bash
echo "00010000\n00020000" > amounts1.txt
echo "00030000" > amounts2.txt
```

The UTXO will be a taproot output encumbered by two tapscripts: one for the
single-output spending transaction in [script1.txt](script1.txt), and one for
the two-output transaction in [script2.txt](script2.txt). We replace the
placeholders in scripts with the keys we created above:

```bash
cat examples/matt/ctv2/script1.txt | sed -E "s/<key1>/`sed -n 1p keys1.txt | awk -F',' '{print $2}'`/" > script1.txt
cat examples/matt/ctv2/script2.txt |sed -E "s/<key1>/`sed -n 1p keys2.txt | awk -F',' '{print $2}'`/"  > script2.txt
cat script2.txt | sed -E "s/<key2>/`sed -n 2p keys2.txt | awk -F',' '{print $2}'`/" > script2.txt
```

The outputs of the spending transaction must match exactly what is specified in
the script being executed. This effectively creates the CTV style covenant.

To simplify debugging in Tapsim, we can also create a tag file, that adds human
readable names to the various data blobs:

```bash
echo "{}" > tags.json
cat tags.json | jq ". += {\"`sed -n 1p keys1.txt | awk -F"," '{print $2}'`\":\"output key[0]\"}" > tags.json
cat tags.json | jq ". += {\"`sed -n 1p keys2.txt | awk -F"," '{print $2}'`\":\"output key[0]\"}" > tags.json
cat tags.json | jq ". += {\"`sed -n 2p keys2.txt | awk -F"," '{print $2}'`\":\"output key[1]\"}" > tags.json
```

## Executing the script
Now we have all the pieces gathered to run the script.

Spending with tx1:

```bash
./tapsim execute --scripts "script1.txt,script2.txt" --scriptindex 0 --tagfile tags.json --outputs="`sed -n 1p keys1.txt | awk -F',' '{print $2}'`:100000000"
```

Spending with tx2:

```bash
./tapsim execute --scripts "script1.txt,script2.txt" --scriptindex 1 --tagfile tags.json  --outputs="`sed -n 1p keys2.txt | awk -F',' '{print $2}'`:10000,`sed -n 2p keys2.txt | awk -F',' '{print $2}'`:20000"
```

### Explanation
The script will check that the spending transaction has the correct outputs.
Note that more inputs and outputs can be added to the transaction without
invalidating the witness (this could be disallowed if one wanted to),

TODO(halseth): 
- add value verification
