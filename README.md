# atomic swap tools

Utilities to perform cross-chain atomic swaps.

## Atomic Swaps with full clients

Supported wallets:

* Ethereum ([Ethereum](https://ethereum.org/))

Find more support coins/wallets on:

 + [rivine/atomicswap](https://github.com/rivine/atomicswap): atomic swap (full-client) tools to swap with BTC and various altcoins
 + [threefoldfoundation/tfchain](https://github.com/threefoldfoundation/tfchain): atomic swap (full-client) tool (`tfchainc`) to swap with TFT

## Atomic Swaps with thin clients

### Electrum

Currently only the Bitcoin Electrum wallet is supported:

* Bitcoin ([Electrum](https://electrum.org/))

The swaps are based upon and compatible with [Decred atomic swaps](https://github.com/decred/atomicswap).

#### Run Electrum as a daemon
Start Electrum on testnet and create a default wallet: 
`./Electrum --testnet`
Configure and start Electrum as a daemon :
```
./Electrum --testnet  setconfig rpcuser user
./Electrum --testnet  setconfig rpcpassword pass
./Electrum --testnet  setconfig rpcport 7777
./Electrum --testnet daemon
./Electrum --testnet daemon load_wallet
```

## Roadmap

- Structure the thin-client code as a library both for Go and C
- Add support for 
  - Litecoin ([Electrum-ltc](https://electrum-ltc.org))
  - Ethereum (light (electrum?) client)

And more coins later on.


