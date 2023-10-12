### Coin pools V2
Example using opcodes `OP_CAT` and `OP_CHECKCONTRACTVERIFY` to achieve a coin
pool construct.

This example is a bit more complete than V1, and supports multiple participants
exiting from the pool in a single transaction (intended to support off-chain
updates among a subset of the participants incase some are offline).

### Introduction
A coin pool contract lets participants in the contract share an UTXO, tracking
the balances of each amongst themselves. As long as all participants are
cooperating, they are gladly signing updated state transactions that reflect
the current balances. However, if one or more of the participants stops
cooperating or goes offline, the state cannot progress and we'll need an
unilateral exit clause.[^1]

In this example we will encode such an unilateral exit clause using the
hypothetical opcodes mentioned above. This can be used to make a subset of the
users keep updating their balances, even though some participant goes offline.
Note that this would require Eltoo style replacements.

Also note that value inspections on the outputs are currently not implemented,
and would be needed to make this really useful.

## Usage
Compile the tools needed (assuming in root folder):
```bash
go build -v ./cmd/tapsim
```

We will assume there are four participants in the contract, and that 2 of them
are exiting. The following script will generate all the data needed to simulate
such spend:
```bash
go run examples/matt/coinpool/v2/cmd/main.go
```

This script generates the pubkey:balance merkle tree, both in starting state,
the intermediate state (after one has exited) and final state (two exited).

## Commitments
The first merkle tree will be built from the leaves
```
(<bal1>,<pub1>) (<bal2>,<pub2>) (<bal3>,<pub3>) (<bal4>,<pub4>) 
```

which commits to the current balances of the participants.

We will then trigger an exit of the first participant, resulting in a new merkle tree built from (note we replace the balance with an empty element)

```
(<>) (<bal2>,<pub2>) (<bal3>,<pub3>) (<bal4>,<pub4>) 
```

Finally, the third participant will exit, resulting in the final state

```
(<>) (<bal2>,<pub2>) (<>) (<bal4>,<pub4>) 
```

Note that the two participants can exit in the same transaction, as long as
they collaborate. This also opens up the possibility for them to exit into a
new pool - allowing online members create new pools among themselves in case
some of the other participants go offline.

## Scripts
The above go script will also generate the Bitcoin tapscripts that is used in
the Taproot coin pool utxo. You can look at them, where they differ in the
number of participants exiting from the pool (output redacted for brevity):

```bash
cat coinpool_v2_1of4exit.txt
...
cat coinpool_v2_2of4exit.txt
...
cat coinpool_v2_3of4exit.txt
...
cat coinpool_v2_4of4exit.txt
...
```

These are scripts (m,n) for m participants spending from a coinpool of in total
n participants.

The keyspend path of the output will be a the same n-of-n multisig key as in
the original pool. This means that when they offline participants come online,
they will need the collaboration of the exited participants to use the keyspend
path (but they can always do an unilateral exit).

## Spend
Now we have everything needed to simulate a spend of the coin pool output. Let
us simulate the participant 1 and 3 exiting the pool by crafting a witness that
satisfies the spend of their combined balance.

User 1 will need to show that its balance is included in the original merkle
root, while user 2 has to do the same on the merkle root where user 1 already
has exited. We also need a merkle root for the case where they have both
exited, as this is needed for the final output. The

The witness will have the following structure:

```
<aggregate pubkey>
<merkle root>
<user 1 balance>
<public key 1>
<merkle path direction>
<merkle sibling level 2>
<merkle path direction>
<merkle sibling level 1>

<user 3 balance>
<public key 3>
<merkle path direction>
<merkle sibling level 2>
<merkle path direction>
<merkle sibling level 1>

<signature user 3>
<signature user 1>
```


## Executing the script
Now we have all the pieces gathered to run the script.

```bash
./tapsim execute --scripts "coinpool_v2_1of4exit.txt,coinpool_v2_2of4exit.txt,coinpool_v2_3of4exit.txt,coinpool_v2_4of4exit.txt" --scriptindex 1 --witness witness_2of4exit.txt --colwidth 80 --privkeys  "`sed -n 1p coinpool_v2_keys.txt | awk -F"," '{print $1}'`,`sed -n 3p coinpool_v2_keys.txt | awk -F"," '{print $1}'`" --inputkey="`sed -n 1p coinpool_v2_innerkey.txt`" --outputkey="`sed -n 2p coinpool_v2_innerkey.txt`"
```

Note that we supply the private keys corresponding to the pubkeys that is exiting
the pool, since it will need a valid signature in order to validate the spend.

### Explanation
The script will check that the public keys and balances supplied in the witness
indeed is committed to in the input. It will then go ahead and build the new
commitment by replacing the publics key with an empty value. It is also
checking that the transaction is signed by the two exiting public keys. Finally
it checks that the created output is indeed committing to the new commitment
and is of the expected value (TODO(halseth): need to add value verification).

[^1]: https://coinpool.dev/
