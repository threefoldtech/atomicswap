# Atomic Swaps with Stellar based assets or Lumens


## Technical

Unlike bitcoin, Stellar has no outputscripts, nor can conditions be combined  or multisig set up for a single payment operation.

In order to solve this, an escrow account is created for every atomic swap on which the required amount of tokens is placed.
The signing conditions  of this escrow account are modified:
- require the signature of the destinee and the secret, 
- signature of the sender combine with the requirement that a specific transaction is present on the chain that can only be published in the future ( timeout mechanism)
