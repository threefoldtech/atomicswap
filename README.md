# atomic swap tools [![Build Status](https://travis-ci.org/threefoldtech/atomicswap.svg?branch=master)](https://travis-ci.org/threefoldtech/atomicswap)

Utilities to perform cross-chain atomic swaps.
The swaps are based upon and compatible with [Decred atomic swaps](https://github.com/decred/atomicswap).
This repository adds tools and support for coins and wallets not supported by the Decred atomic swap tools.

## Atomic Swaps with full clients

Supported wallets:

* Ethereum ([Ethereum](https://ethereum.org/))

Find more support coins/wallets on:

* [rivine/decredatomicswap](https://github.com/rivine/decredatomicswap): atomic swap (full-client) tools to swap with BTC and various altcoins. Binaries for the original Decred Atomic swap tools are also provided here.
* [threefoldfoundation/tfchain](https://github.com/threefoldfoundation/tfchain): atomic swap (full-client) tool (`tfchainc`) to swap with TFT

## Atomic Swaps with thin clients

### Electrum

Currently only the Bitcoin Electrum wallet is supported:

* Bitcoin ([Electrum](https://electrum.org/))

#### Run Electrum as a daemon

Start Electrum on testnet and create a default wallet:
`./Electrum --testnet`

Configure and start Electrum as a daemon :

```sh
./Electrum --testnet  setconfig rpcuser user
./Electrum --testnet  setconfig rpcpassword pass
./Electrum --testnet  setconfig rpcport 7777
./Electrum --testnet daemon
./Electrum --testnet daemon load_wallet
```

## Atomic swaps withouth an external wallet process

* [Stellar](https://stellar.org) based assets and Lumens: [StellarAtomicSwaps](cmd/stellaratomicswap/readme.md)

## Roadmap

* Add Stellar based non native asset support
* Structure the thin-client code as a library both for Go and C
* Add support for
  * Litecoin ([Electrum-ltc](https://electrum-ltc.org))
  * Ethereum (light (electrum?) client)

And more coins later on.

## Repository Owners

* Rob Van Mieghem ([@robvanmieghem](https://github.com/robvanmieghem))
* Lee Smet ([@leesmet](https://github.com/leesmet))
* Glen De Cauwsemaecker ([@glendc](https://github.com/glendc))
