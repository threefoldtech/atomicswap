# Bitcoin Atomic swaps for the Electrum client

## Compatibility

Electrum-3.3.8 or up


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

