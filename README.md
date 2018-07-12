# atomic swaps using thin clients.

Utilities to  perform cross-chain atomic swaps using thin clients.  
Currently only the Bitcoin Electrum wallet is supported:

* Bitcoin ([Electrum](https://electrum.org/))

The swaps are based upon and compatible with [Decred atomic swaps](https://github.com/decred/atomicswap).

## Run Electrum as a daemon
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

- Structure the code as a library both for Go and C
- Add support for 
  - Litecoin ([Electrum-ltc](https://electrum-ltc.org))
  - Ethereum

And more coins later on.


