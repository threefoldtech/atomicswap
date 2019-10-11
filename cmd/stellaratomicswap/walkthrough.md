 # Stellar atomic swap walkthrough

A walkthrough of an atomic swap of lumen and bitcoin using the the bitcoin Electrum thin client.

This example is a walkthrough of an actual atomic swap on the threefold and bitcoin testnets.

## perequisites  
In order to execute atomic swaps as described in this document, you need to have the [Electrum wallet](https://electrum.org/).


On osx the default location is `/Applications/Electrum.app/Contents/MacOS`.

Start Electrum on testnet and create a default wallet but do not set a password on it: `./Electrum --testnet`


Configure and start Electrum as a daemon :
```
./Electrum --testnet  setconfig rpcuser user
./Electrum --testnet  setconfig rpcpassword pass
./Electrum --testnet  setconfig rpcport 7777
./Electrum --testnet daemon
```
While the daemon is running, make it load the wallet in a different shell:
```
./Electrum --testnet daemon load_wallet
```

## Example

Let's assume Bob wants to buy 567 XLM from Alice for 0.1234BTC

Bob creates a bitcoin address and Alice creates( or reuses) a Stellar Account.

