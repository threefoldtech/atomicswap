# Atomic Swaps with Stellar based assets or Lumens

A [detailed walkthrough](walkthrough.md) is available that demonstrates an atomic swap with bitcoin.

## Technical

Unlike bitcoin, Stellar has no outputscripts, nor can conditions be combined  or multisig set up for a single payment operation.

In order to solve this, an escrow account is created for every atomic swap on which the required amount of tokens is placed.
The signing conditions of this escrow account are modified:

- signature of the destinee and the secret
- hash of a specific transaction that is present on the chain  that merges the escrow account to the account that needs to withdraw and that can only be published in the future ( timeout mechanism)
