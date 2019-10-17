# Stellar atomic swap walkthrough

A walkthrough of an atomic swap of lumen and bitcoin using the the bitcoin Electrum thin client.

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
mytQ9begA7nkKaNfAu8mXBPhJAD3LWGAbp
```

### initiate step

Bob initiates the process by using btcatomicswap to pay 0.1234BTC into the Bitcoin contract using Alice's Bit coin address, sending the contract transaction, and sharing the secret hash (not the secret), and contract's transaction with Alice. The refund transaction can not be sent until the locktime expires, but should be saved in case a refund is necessary.

command:`btcatomicswap [-testnet]initiate <participant address> <amount>`

```sh
$ ./btcatomicswap -testnet --rpcuser=user --rpcpass=pass -s  "localhost:7777" initiate mytQ9begA7nkKaNfAu8mXBPhJAD3LWGAbp  0.1234
ecret:      b0d91a0814934a16f8dafc62e75ae03c18cc74489628b0bacf6252f5701024c2
Secret hash: 6c3b8ae32bfca727821486a0f79ff1d26800df9df1763099ea188098cfdc241e

Contract fee: 0.00000372 BTC (0.00001000 BTC/kB)
Refund fee:   0.00000297 BTC (0.00001021 BTC/kB)

Contract (2MtVYgRTBNjdzoyQ6v3sVLALpngcFCVgTYo):
6382012088a8206c3b8ae32bfca727821486a0f79ff1d26800df9df1763099ea188098cfdc241e8876a914c98094d0f6bbb95e6729539af1642f65d911099967040908ab5db17576a914c98094d0f6bbb95e6729539af1642f65d91109996888ac

Contract transaction (84b175c43497d0be557753bc2222d27fc948673a42a02a4b7e15183b8af1780a):
0200000002d54a32316af9cf1fd2ccf94a9421519184467f8f1e8f4d030ad3fac911b8b00a000000006b4830450221008aceca8f93247644045590d72d3836cd7cc454c4ad9676b71ea545ce205cf61102205c1471209528df179cbaea2eef31739a52543a6b392c57a7b16de9131cefc51e0121035ceaa55c2bfb92c9c3266da89097fa4173d1561bd7b07c91d70b6ac1e01f0f94fdffffff71c5c5b8f1b28dd103d13197aaa961515ded308e52eebbf9df48e3714a08d623000000006b4830450221008b2e155816e217d29cc11a224c2bdfe394490a266f6c5b5b5cacb85062910c8f022011f844cf6ae5ef39f939bdc4d25993c83cb6d80b224605c8be36ec40848eed240121030b658c35df930aa8c17c94e8439e4a6daa3c025713997687eb481103aa3baa5dfdffffff024b069100000000001976a914dee4c7011fcf37ef36e91b76618dcf7f1e92488788ac1f4bbc000000000017a9140dad8d75de0448ba757e5337b3bea2b90f7e2acc8757281800

Refund transaction (8b966a55144f3710a967d8d553e463eebb699d36c11c05782ad460073fe567db):
02000000010a78f18a3b18157e4b2aa0423a6748c97fd22222bc537755bed09734c475b18401000000ce47304402201299d3b185d0342cdc8ebf04bffc8f7caa88fc0212739e59220155db56bb0c9302202e24382efd388455e3e8f566889eb1cebfdc5e0258a5e894fd09516b3099fa510121020e0081ab8cfaee35f6080403efd17e377f493eec93b6a44181279419a913962a004c616382012088a8206c3b8ae32bfca727821486a0f79ff1d26800df9df1763099ea188098cfdc241e8876a914c98094d0f6bbb95e6729539af1642f65d911099967040908ab5db17576a914c98094d0f6bbb95e6729539af1642f65d91109996888ac0000000001f649bc00000000001976a914c98094d0f6bbb95e6729539af1642f65d911099988ac0908ab5d

Publish contract transaction? [y/N] y
Published contract transaction (84b175c43497d0be557753bc2222d27fc948673a42a02a4b7e15183b8af1780a)
```

You can check the transaction [on a bitcoin testnet blockexplorer](https://live.blockcypher.com/btc-testnet/tx/84b175c43497d0be557753bc2222d27fc948673a42a02a4b7e15183b8af1780a/) where you can see that 0.1234 BTC is sent to 2MtVYgRTBNjdzoyQ6v3sVLALpngcFCVgTYo (= the contract script hash) being a [p2sh](https://en.bitcoin.it/wiki/Pay_to_script_hash) address in the bitcoin testnet.

### audit contract

Bob sends Alice the contract and the contract transaction. Alice should now verify if

- the script is correct
- the locktime is far enough in the future
- the amount is correct
- she is the recipient

command:`btcatomicswap [-testnet] auditcontract <contract> <contract transaction>`

 ```sh
$./btcatomicswap -testnet --rpcuser=user --rpcpass=pass  -s  "localhost:7777" auditcontract 6382012088a8206c3b8ae32bfca727821486a0f79ff1d26800df9df1763099ea188098cfdc241e8876a914c98094d0f6bbb95e6729539af1642f65d911099967040908ab5db17576a914c98094d0f6bbb95e6729539af1642f65d91109996888ac 0200000002d54a32316af9cf1fd2ccf94a9421519184467f8f1e8f4d030ad3fac911b8b00a000000006b4830450221008aceca8f93247644045590d72d3836cd7cc454c4ad9676b71ea545ce205cf61102205c1471209528df179cbaea2eef31739a52543a6b392c57a7b16de9131cefc51e0121035ceaa55c2bfb92c9c3266da89097fa4173d1561bd7b07c91d70b6ac1e01f0f94fdffffff71c5c5b8f1b28dd103d13197aaa961515ded308e52eebbf9df48e3714a08d623000000006b4830450221008b2e155816e217d29cc11a224c2bdfe394490a266f6c5b5b5cacb85062910c8f022011f844cf6ae5ef39f939bdc4d25993c83cb6d80b224605c8be36ec40848eed240121030b658c35df930aa8c17c94e8439e4a6daa3c025713997687eb481103aa3baa5dfdffffff024b069100000000001976a914dee4c7011fcf37ef36e91b76618dcf7f1e92488788ac1f4bbc000000000017a9140dad8d75de0448ba757e5337b3bea2b90f7e2acc8757281800
Contract address:        2MtVYgRTBNjdzoyQ6v3sVLALpngcFCVgTYo
Contract value:          0.12339999 BTC
Recipient address:       mytQ9begA7nkKaNfAu8mXBPhJAD3LWGAbp
Refund address: mytQ9begA7nkKaNfAu8mXBPhJAD3LWGAbp

Secret hash: 6c3b8ae32bfca727821486a0f79ff1d26800df9df1763099ea188098cfdc241e

Locktime: 2019-10-19 12:56:41 +0000 UTC
Locktime reached in 47h57m5s
```

WARNING:
A check on the blockchain should be done as the auditcontract does not do that so an already spent output could have been used as an input. Checking if the contract has been mined in a block should suffice

### Participate

Alice trusts the contract so she participates in the atomic swap by paying the lumens into a new stellar holding account using the same secret hash as part of the signing conditions.  

Bob uses an existing Stellar account ( or creates a new one): *GB7CA4F3VLBJN5UXHFDARY62FKUB5C5UH244JSNWMDOJZ54WTEIK7XBV*

Bob sends this address to Alice who uses it to participate in the swap.
command:`stellaratomicswap [-testnet] participate <participant seed> <initiator address> <amount> <secret hash>`

```sh
$ ./stellaratomicswap -testnet participate SCTSUDD37C7CYAE7YEGKT3DFQOOHPDGWDIIBGGLCIE4NWDJWRVGSM2W4 GB7CA4F3VLBJN5UXHFDARY62FKUB5C5UH244JSNWMDOJZ54WTEIK7XBV 567 6c3b8ae32bfca727821486a0f79ff1d26800df9df1763099ea188098cfdc241e
participant address: GDZILW42QMVDVKGC3H7B3LBQJAN74QAWMWKBPCW74TUCIOCL443KZJ72
holding account address: GBE6SYWCTVGFP4FL75O2THYLVE44ELQQZ5CB4QKT4MVXGHQ52W3OJEJR
refund transaction:
AAAAAEnpYsKdTFfwq/9dqZ8LqTnCLhDPRB5BU+Mrcx4d1bbkAAAAZAAS0pUAAAACAAAAAQAAAABdqbf/AAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAASeliwp1MV/Cr/12pnwupOcIuEM9EHkFT4ytzHh3VtuQAAAAIAAAAAPKF25qDKjqowtn+HawwSBv+QBZllBeK3+ToJDhL5zasAAAAAAAAAAA=
```

The above command creates a holding account with 567 Lumen on it with modified signing conditions, matching the atomic swap conditiuons. [On a testnet explorer one can view the details of the holding account](
https://testnet.steexp.com/account/GBE6SYWCTVGFP4FL75O2THYLVE44ELQQZ5CB4QKT4MVXGHQ52W3OJEJR).

Alice now informs Bob that the Stellar contract account  has been created and provides him with the holding account address and the refund transaction.

### audit Stellar contract

Just as Alice had to audit Bob's contract, Bob now has to do the same with Alice's holding account before withdrawing.

Bob verifies if:

- the amount of tokens on the account is correct
- the locktime, hashed secret and wallet address defined in the signing conditions are correct

command:`stellaratomicswap [-tesnet] auditcontract holdingAccountAddress refundTransaction`
flags are available to automatically check the information in the contract.

```sh
$ ./stellaratomicswap -testnet auditcontract GBE6SYWCTVGFP4FL75O2THYLVE44ELQQZ5CB4QKT4MVXGHQ52W3OJEJR AAAAAEnpYsKdTFfwq/9dqZ8LqTnCLhDPRB5BU+Mrcx4d1bbkAAAAZAAS0pUAAAACAAAAAQAAAABdqbf/AAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAASeliwp1MV/Cr/12pnwupOcIuEM9EHkFT4ytzHh3VtuQAAAAIAAAAAPKF25qDKjqowtn+HawwSBv+QBZllBeK3+ToJDhL5zasAAAAAAAAAAA=
Contract address:        GBE6SYWCTVGFP4FL75O2THYLVE44ELQQZ5CB4QKT4MVXGHQ52W3OJEJR
Contract value:          566.9999600
Recipient address:       GB7CA4F3VLBJN5UXHFDARY62FKUB5C5UH244JSNWMDOJZ54WTEIK7XBV
Refund address: GDZILW42QMVDVKGC3H7B3LBQJAN74QAWMWKBPCW74TUCIOCL443KZJ72

Secret hash: 6c3b8ae32bfca727821486a0f79ff1d26800df9df1763099ea188098cfdc241e

Locktime: 2019-10-18 13:02:55 +0000 UTC
Locktime reached in 23h57m17s
```

### redeem tokens

Now that both Bob and Alice have paid into their respective contracts, Bob withdraws from the Stellar contract. This step involves publishing a transaction which reveals the secret to Alice, allowing her to withdraw from the Bitcoin contract.

command:`stellaratomicswap [-tesnet] redeem receiverseed  holdingAccountAddress secret`

```sh
$ ./stellaratomicswap -testnet redeem SCCSIKGU4F5JS4LY5QZ2OXYMTQTQVJF3TOA7PZVTDDPOPE7BXYAZOITZ GBE6SYWCTVGFP4FL75O2THYLVE44ELQQZ5CB4QKT4MVXGHQ52W3OJEJR b0d91a0814934a16f8dafc62e75ae03c18cc74489628b0bacf6252f5701024c2
***TransactionSuccess dump***
    Links: {{https://horizon-testnet.stellar.org/transactions/4eae3f7b85033a5e76ae4eefb3ba1997df81ca6ae9667d8836261b4c37d2be2f false}}
    Hash: 4eae3f7b85033a5e76ae4eefb3ba1997df81ca6ae9667d8836261b4c37d2be2f
    Ledger: 1233735
    Env: AAAAAEnpYsKdTFfwq/9dqZ8LqTnCLhDPRB5BU+Mrcx4d1bbkAAAAZAAS0pUAAAACAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAASeliwp1MV/Cr/12pnwupOcIuEM9EHkFT4ytzHh3VtuQAAAAIAAAAAH4gcLuqwpb2lzlGCOPaKqgei7Q+ucTJtmDcnPeWmRCvAAAAAAAAAALP3CQeAAAAILDZGggUk0oW+Nr8Yuda4DwYzHRIliiwus9iUvVwECTClpkQrwAAAEBvyl/3ZjsOwWavwTQwNLyiswe27fZbSbQyfUjv1XnYTvCMwyQdrVwIJ79uXi5nCyFHWo/tqf5zdL7x7PKoieMD
    Result: AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAIAAAAAAAAAAFR9VOMAAAAAA==
    Meta: AAAAAQAAAAIAAAADABLTRwAAAAAAAAAASeliwp1MV/Cr/12pnwupOcIuEM9EHkFT4ytzHh3VtuQAAAABUfVTjAAS0pUAAAABAAAAAwAAAAAAAAAAAAAAAAACAgIAAAADAAAAAH4gcLuqwpb2lzlGCOPaKqgei7Q+ucTJtmDcnPeWmRCvAAAAAQAAAAHWpnUj4pDcRbZM/CQT5Owpwmt9pytX8iuqJKH7NWGk8gAAAAIAAAACbDuK4yv8pyeCFIag95/x0mgA353xdjCZ6hiAmM/cJB4AAAABAAAAAAAAAAAAAAABABLTRwAAAAAAAAAASeliwp1MV/Cr/12pnwupOcIuEM9EHkFT4ytzHh3VtuQAAAABUfVTjAAS0pUAAAACAAAAAwAAAAAAAAAAAAAAAAACAgIAAAADAAAAAH4gcLuqwpb2lzlGCOPaKqgei7Q+ucTJtmDcnPeWmRCvAAAAAQAAAAHWpnUj4pDcRbZM/CQT5Owpwmt9pytX8iuqJKH7NWGk8gAAAAIAAAACbDuK4yv8pyeCFIag95/x0mgA353xdjCZ6hiAmM/cJB4AAAABAAAAAAAAAAAAAAABAAAABAAAAAMAEhg5AAAAAAAAAAB+IHC7qsKW9pc5Rgjj2iqoHou0PrnEybZg3Jz3lpkQrwAAABdIdugAABIYOQAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAEtNHAAAAAAAAAAB+IHC7qsKW9pc5Rgjj2iqoHou0PrnEybZg3Jz3lpkQrwAAABiabDuMABIYOQAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAMAEtNHAAAAAAAAAABJ6WLCnUxX8Kv/XamfC6k5wi4Qz0QeQVPjK3MeHdW25AAAAAFR9VOMABLSlQAAAAIAAAADAAAAAAAAAAAAAAAAAAICAgAAAAMAAAAAfiBwu6rClvaXOUYI49oqqB6LtD65xMm2YNyc95aZEK8AAAABAAAAAdamdSPikNxFtkz8JBPk7CnCa32nK1fyK6okofs1YaTyAAAAAgAAAAJsO4rjK/ynJ4IUhqD3n/HSaADfnfF2MJnqGICYz9wkHgAAAAEAAAAAAAAAAAAAAAIAAAAAAAAAAEnpYsKdTFfwq/9dqZ8LqTnCLhDPRB5BU+Mrcx4d1bbk
```
