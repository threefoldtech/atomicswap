# atomic swap tools

Utilities to perform cross-chain atomic swaps.
The swaps are based upon and compatible with [Decred atomic swaps](https://github.com/decred/atomicswap).
This repository adds tools and support for coins and wallets not supported by the Decred atomic swap tools.

## Atomic Swaps with full clients

Supported wallets:

* Ethereum ([Ethereum](https://ethereum.org/))

## Atomic Swaps with thin clients

### Electrum

Currently only the Bitcoin Electrum wallet is supported:

* Bitcoin ([Electrum](https://electrum.org/)): [BTCAtomicwap](./cmd/btcatomicswap)

## Atomic swaps without an external wallet process

* [Stellar](https://stellar.org) based assets and Lumens: [StellarAtomicSwaps](cmd/stellaratomicswap/readme.md)

## Repository Owners

* Rob Van Mieghem ([@robvanmieghem](https://github.com/robvanmieghem))
* Lee Smet ([@leesmet](https://github.com/leesmet))
