# atomic swap ending in a refund walkthrough

A walkthrough of an atomic swap of lumen and bitcoin using the the bitcoin Electrum thin client ending in a refund.

This example is a walkthrough of an actual atomic swap on the threefold and bitcoin testnets.

## perequisites

In order to execute atomic swaps as described in this document, you need to have the [Electrum wallet](https://electrum.org/).

On osx the default location is `/Applications/Electrum.app/Contents/MacOS`.

Start Electrum on testnet and create a default wallet but do not set a password on it: `./Electrum --testnet`

Configure and start Electrum as a daemon :

```sh
./Electrum --testnet  setconfig rpcuser user
./Electrum --testnet  setconfig rpcpassword pass
./Electrum --testnet  setconfig rpcport 7777
./Electrum --testnet daemon
```

While the daemon is running, make it load the wallet in a different shell:

```sh
./Electrum --testnet daemon load_wallet
```

## Example

Let's assume Bob wants to buy 567 XLM from Alice for 0.1234BTC

Alice creates a new bitcoin  address and provides this to Bob:

```￼sh
./Electrum --testnet getunusedaddress
mzQv1c5fatwFw1LoUAp9s8Zt6thzqiCWm2
```

### initiate step

Bob initiates the process by using btcatomicswap to pay 0.1234BTC into the Bitcoin contract using Alice's Bit coin address, sending the contract transaction, and sharing the secret hash (not the secret), and contract's transaction with Alice. The refund transaction can not be sent until the locktime expires, but should be saved in case a refund is necessary.

command:`btcatomicswap initiate <participant address> <amount>`

```sh
$ ./btcatomicswap -testnet --rpcuser=user --rpcpass=pass -s  "localhost:7777" initiate mzQv1c5fatwFw1LoUAp9s8Zt6thzqiCWm2  0.1234
Secret:      d049d9229fa1c0e6980d8c5db336af3b208a3c1707381647f81a55e0d08341c4
Secret hash: 4c6d4a1493c03c45ae259d3b3eaba98aab2b37c8c10bc93d9db77b577381adcd

Contract fee: 0.00000224 BTC (0.00001000 BTC/kB)
Refund fee:   0.00000297 BTC (0.00001021 BTC/kB)

Contract (2MwV3tC5BdhCNQ5cBdD7XML6x38hqQ5HhTc):
6382012088a8204c6d4a1493c03c45ae259d3b3eaba98aab2b37c8c10bc93d9db77b577381adcd8876a914cf45fce54ebfcd30681d99fdf060e2ab56478888670471eca65db17576a914cf45fce54ebfcd30681d99fdf060e2ab564788886888ac

Contract transaction (a98f71e26f495d6d6c75a9c3537d185b09ad10f7cb3bd5532506a3a111a77064):
020000000140f64f5e25a0571cd8fae3481586750f36237db1a92b710a8f9cea3b892a40db010000006b4830450221009d1240ae9cc0f5ffd1cc85743d18ac5bbcb505923aa438a9f35dd0243a98940302206a67579137976c5fcc7ee8658a88a73bf7d0666537e9c22ef84ff8b2716374e60121034ed8e2048cabfdb701b593045ce7314acfdbeeeffc7cde5027e31af8d7dab7c8fdffffff0283552800000000001976a9141388fd5a6faba38da249cbdd92d8951547298e9a88ac1f4bbc000000000017a9142e7de42ec6ebb700cba09eb743b543408ff1fca7876d271800

Refund transaction (0ab0b811c9fad30a034d8f1e8f7f4684915121944af9ccd21fcff96a31324ad5):
02000000016470a711a1a3062553d53bcbf710ad095b187d53c3a9756c6d5d496fe2718fa901000000ce473044022035295f681c4c3a7f28d5d889f982c6069d832b20a7eef8737cdba559ac3a226a02207644def8d6b3cd4bdef6edd8cdf5900ebe0dfda9f573ebeaf14e80ff9a2d36460121035ceaa55c2bfb92c9c3266da89097fa4173d1561bd7b07c91d70b6ac1e01f0f94004c616382012088a8204c6d4a1493c03c45ae259d3b3eaba98aab2b37c8c10bc93d9db77b577381adcd8876a914cf45fce54ebfcd30681d99fdf060e2ab56478888670471eca65db17576a914cf45fce54ebfcd30681d99fdf060e2ab564788886888ac0000000001f649bc00000000001976a914cf45fce54ebfcd30681d99fdf060e2ab5647888888ac71eca65d

Publish contract transaction? [y/N] y
Published contract transaction (a98f71e26f495d6d6c75a9c3537d185b09ad10f7cb3bd5532506a3a111a77064)
```

You can check the transaction [on a bitcoin testnet blockexplorer](https://live.blockcypher.com/btc-testnet/tx/a98f71e26f495d6d6c75a9c3537d185b09ad10f7cb3bd5532506a3a111a77064/) where you can see that 0.1234 BTC is sent to 2MwV3tC5BdhCNQ5cBdD7XML6x38hqQ5HhT (= the contract script hash) being a [p2sh](https://en.bitcoin.it/wiki/Pay_to_script_hash) address in the bitcoin testnet.

### audit contract

Bob sends Alice the contract and the contract transaction. Alice should now verify if

- the script is correct
- the locktime is far enough in the future
- the amount is correct
- she is the recipient

command:`btcatomicswap auditcontract <contract> <contract transaction>`

 ```sh
$./btcatomicswap --testnet --rpcuser=user --rpcpass=pass  -s  "localhost:7777" auditcontract 6382012088a8204c6d4a1493c03c45ae259d3b3eaba98aab2b37c8c10bc93d9db77b577381adcd8876a914cf45fce54ebfcd30681d99fdf060e2ab56478888670471eca65db17576a914cf45fce54ebfcd30681d99fdf060e2ab564788886888ac 020000000140f64f5e25a0571cd8fae3481586750f36237db1a92b710a8f9cea3b892a40db010000006b4830450221009d1240ae9cc0f5ffd1cc85743d18ac5bbcb505923aa438a9f35dd0243a98940302206a67579137976c5fcc7ee8658a88a73bf7d0666537e9c22ef84ff8b2716374e60121034ed8e2048cabfdb701b593045ce7314acfdbeeeffc7cde5027e31af8d7dab7c8fdffffff0283552800000000001976a9141388fd5a6faba38da249cbdd92d8951547298e9a88ac1f4bbc000000000017a9142e7de42ec6ebb700cba09eb743b543408ff1fca7876d271800
Contract address:        2MwV3tC5BdhCNQ5cBdD7XML6x38hqQ5HhTc
Contract value:          0.12339999 BTC
Recipient address:       mzQv1c5fatwFw1LoUAp9s8Zt6thzqiCWm2
Author's refund address: mzQv1c5fatwFw1LoUAp9s8Zt6thzqiCWm2 (is by accident the same as the recipient since it was executed on the same wallet on this walkthrough)

Secret hash: 4c6d4a1493c03c45ae259d3b3eaba98aab2b37c8c10bc93d9db77b577381adcd

Locktime: 2019-10-16 10:09:53 +0000 UTC
Locktime reached in 45h29m48s
```

WARNING:
A check on the blockchain should be done as the auditcontract does not do that so an already spent output could have been used as an input. Checking if the contract has been mined in a block should suffice

### Participate

Alice trusts the contract so she participates in the atomic swap by paying the lumens into a new stellar holding account using the same secret hash as part of the signing conditions.  

Bob uses an existing Stellar account ( or creates a new one): *GB7CA4F3VLBJN5UXHFDARY62FKUB5C5UH244JSNWMDOJZ54WTEIK7XBV*

Bob sends this address to Alice who uses it to participate in the swap.
command:`stellaratomicswap [-testnet] participate <participant seed> <initiator address> <amount> <secret hash>`

```sh
$ ./stellaratomicswap -testnet participate SCTSUDD37C7CYAE7YEGKT3DFQOOHPDGWDIIBGGLCIE4NWDJWRVGSM2W4 GB7CA4F3VLBJN5UXHFDARY62FKUB5C5UH244JSNWMDOJZ54WTEIK7XBV 567 4c6d4a1493c03c45ae259d3b3eaba98aab2b37c8c10bc93d9db77b577381adcd
participant address: GDZILW42QMVDVKGC3H7B3LBQJAN74QAWMWKBPCW74TUCIOCL443KZJ72
holding account address: GA2VHTC3JENF324AHNBNWTOCZMHYDA2GH3HZFYQJIAC6AH5VTWO35UH6
refund transaction:
AAAAADVTzFtJGl3rgDtC203Cyw+Bg0Y+z5LiCUAF4B+1nZ2+AAAAZAASGUIAAAACAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAANVPMW0kaXeuAO0LbTcLLD4GDRj7PkuIJQAXgH7Wdnb4AAAAIAAAAAPKF25qDKjqowtn+HawwSBv+QBZllBeK3+ToJDhL5zasAAAAAAAAAAA=
```

The above command creates a holding account with 567 Lumen on it with modified signing conditions, matching the atomic swap conditiuons. [On a testnet explorer one can view the details of the holding account](
https://testnet.steexp.com/account/GA2VHTC3JENF324AHNBNWTOCZMHYDA2GH3HZFYQJIAC6AH5VTWO35UH6).

Alice now informs Bob that the Stellar contract account  has been created and provides him with the holding account address and the refund transaction.

### audit Stellar contract

Just as Alice had to audit Bob's contract, Bob now has to do the same with Alice's holding account before withdrawing.

Bob verifies if:

- the amount of tokens on the account is correct
- the locktime, hashed secret and wallet address defined in the signing conditions are correct

command:`stellaratomicswap [-tesnet] auditcontract holdingAccountAddress refundTransaction`
flags are available to automatically check the information in the contract.

```sh
$ ./stellaratomicswap -testnet auditcontract GA2VHTC3JENF324AHNBNWTOCZMHYDA2GH3HZFYQJIAC6AH5VTWO35UH6 AAAAADVTzFtJGl3rgDtC203Cyw+Bg0Y+z5LiCUAF4B+1nZ2+AAAAZAASGUIAAAACAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAANVPMW0kaXeuAO0LbTcLLD4GDRj7PkuIJQAXgH7Wdnb4AAAAIAAAAAPKF25qDKjqowtn+HawwSBv+QBZllBeK3+ToJDhL5zasAAAAAAAAAAA=
Contract address:        GA2VHTC3JENF324AHNBNWTOCZMHYDA2GH3HZFYQJIAC6AH5VTWO35UH6
Contract value:          566.9999600
Recipient address:       GB7CA4F3VLBJN5UXHFDARY62FKUB5C5UH244JSNWMDOJZ54WTEIK7XBV
Refund address: GDZILW42QMVDVKGC3H7B3LBQJAN74QAWMWKBPCW74TUCIOCL443KZJ72

Secret hash: 4c6d4a1493c03c45ae259d3b3eaba98aab2b37c8c10bc93d9db77b577381adcd

Locktime: 1970-01-01 00:00:00 +0000 UTC
Refund time lock has expired
```

## refund

In case Bob does not withraw the stellar tokens and as such does not share the secret, Alice can withdraw her funds in the holding account.
command:`stellaratomicswap [-testnet] refund refundTransaction`

Bob can reclaim his bitcoins as well:

command:`btcatomicswap [-testnet] refund  contract contractTransaction`

```sh
./btcatomicswap --testnet --rpcuser=user --rpcpass=pass  -s  "localhost:7777" refund 6382012088a8204c6d4a1493c03c45ae259d3b3eaba98aab2b37c8c10bc93d9db77b577381adcd8876a914cf45fce54ebfcd30681d99fdf060e2ab56478888670471eca65db17576a914cf45fce54ebfcd30681d99fdf060e2ab564788886888ac 020000000140f64f5e25a0571cd8fae3481586750f36237db1a92b710a8f9cea3b892a40db010000006b4830450221009d1240ae9cc0f5ffd1cc85743d18ac5bbcb505923aa438a9f35dd0243a98940302206a67579137976c5fcc7ee8658a88a73bf7d0666537e9c22ef84ff8b2716374e60121034ed8e2048cabfdb701b593045ce7314acfdbeeeffc7cde5027e31af8d7dab7c8fdffffff0283552800000000001976a9141388fd5a6faba38da249cbdd92d8951547298e9a88ac1f4bbc000000000017a9142e7de42ec6ebb700cba09eb743b543408ff1fca7876d271800
Refund fee: 0.00000297 BTC (0.00001021 BTC/kB)

Refund transaction (0ab0b811c9fad30a034d8f1e8f7f4684915121944af9ccd21fcff96a31324ad5):
02000000016470a711a1a3062553d53bcbf710ad095b187d53c3a9756c6d5d496fe2718fa901000000ce473044022035295f681c4c3a7f28d5d889f982c6069d832b20a7eef8737cdba559ac3a226a02207644def8d6b3cd4bdef6edd8cdf5900ebe0dfda9f573ebeaf14e80ff9a2d36460121035ceaa55c2bfb92c9c3266da89097fa4173d1561bd7b07c91d70b6ac1e01f0f94004c616382012088a8204c6d4a1493c03c45ae259d3b3eaba98aab2b37c8c10bc93d9db77b577381adcd8876a914cf45fce54ebfcd30681d99fdf060e2ab56478888670471eca65db17576a914cf45fce54ebfcd30681d99fdf060e2ab564788886888ac0000000001f649bc00000000001976a914cf45fce54ebfcd30681d99fdf060e2ab5647888888ac71eca65d

Publish refund transaction? [y/N] y
Published refund transaction (0ab0b811c9fad30a034d8f1e8f7f4684915121944af9ccd21fcff96a31324ad5)
```
