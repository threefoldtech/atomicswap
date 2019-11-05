# Stellar non native assets  atomic swap walkthrough

A walkthrough of an atomic swap of a [custom issued Stellar testnet bitcoin](https://github.com/threefoldtech/rivine/blob/master/research/stellar/examples/issuetoken/readme.md), BTC:GDPHIMRSUSZNLNFWW7VJWWQ2NCH6D6ZVJ4RIME3FUGZLJRS3KKNIVYQ5 in our case and bitcoin using the the bitcoin Electrum thin client.

This example is a walkthrough of an actual atomic swap on the Stellar and bitcoin testnets.

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

Let's assume Bob wants to buy 0.05 XLMBTC from Alice for 0.05BTC

Alice creates a new bitcoin  address and provides this to Bob:

```￼sh
./Electrum --testnet getunusedaddress
mz7DkWBB9JQrReutRbtwU5JZK5WEvQs3Md
```

### initiate step

Bob initiates the process by using btcatomicswap to pay 0.05BTC into the Bitcoin contract using Alice's Bit coin address, sending the contract transaction, and sharing the secret hash (not the secret), and contract's transaction with Alice. The refund transaction can not be sent until the locktime expires, but should be saved in case a refund is necessary.

command:`btcatomicswap [-testnet]initiate <participant address> <amount>`

```sh
$ ./btcatomicswap -testnet --rpcuser=user --rpcpass=pass -s  "localhost:7777" initiate mz7DkWBB9JQrReutRbtwU5JZK5WEvQs3Md  0.05
Secret:      72312e9f0f696f26f9cb391e222ba9aa43bb8d1f0c1d26fce98262319d988231
Secret hash: 6036b8a2fedaadb621bf043395167e5ee1190f84261ac878ebd6fa38e100f98b

Contract fee: 0.00000224 BTC (0.00001004 BTC/kB)
Refund fee:   0.00000297 BTC (0.00001017 BTC/kB)

Contract (2N6DrM8kBbeNAfL65JywN8VnjyZeLhp67ik):
6382012088a8206036b8a2fedaadb621bf043395167e5ee1190f84261ac878ebd6fa38e100f98b8876a914cbed4f110fc43fe309975fd7c134608188ede5c06704de5cbd5db17576a914cbed4f110fc43fe309975fd7c134608188ede5c06888ac

Contract transaction (bf27f560e49a1b5f64a4d1b5443f6a14d05bf2056d500cecb392631e54b09dd6):
020000000123b234d4a59281756edb30f8d1701478b74035d63eded22f4d0877b1f10a9181010000006a473044022021792701795cb7988f5a86eaf0eccd5a914abbd39731a7ad66508db361cf3139022031e085f3b3305925b38bfeb7fb4110ef8daccff65e3d4f026ac1e62fcbb4973301210206b803413788c63fb600208c8226f5f396b340d9f41f74c137f09a0c6d058a6efdffffff0295b12300000000001976a914375651c9d6e5d5e9be2540dbe5df310ee3b4537f88ac404b4c000000000017a9148e57279669d66bcc0d22c09a5eb60b75c958d49b87ba2d1800

Refund transaction (0e6af079b03a6803988c32985dda5751dbee94defb579740efab2c24aada6a37):
0200000001d69db0541e6392b3ec0c506d05f25bd0146a3f44b5d1a4645f1b9ae460f527bf01000000cf4830450221009fa47f4dae2cff42469b7663dc4081a12354078b9e97572e7acc446924909111022023a653852060cf4978b7cca984ed0c2a530d14e7baa730388a7754639a360f010121028f02b7f1f0614d179c061cf84aa3cdb0332983764661fcd1228b043f94b8a0c6004c616382012088a8206036b8a2fedaadb621bf043395167e5ee1190f84261ac878ebd6fa38e100f98b8876a914cbed4f110fc43fe309975fd7c134608188ede5c06704de5cbd5db17576a914cbed4f110fc43fe309975fd7c134608188ede5c06888ac0000000001174a4c00000000001976a914cbed4f110fc43fe309975fd7c134608188ede5c088acde5cbd5d

Publish contract transaction? [y/N] y
Published contract transaction (bf27f560e49a1b5f64a4d1b5443f6a14d05bf2056d500cecb392631e54b09dd6)
```

You can check the transaction [on a bitcoin testnet blockexplorer](https://live.blockcypher.com/btc-testnet/tx/bf27f560e49a1b5f64a4d1b5443f6a14d05bf2056d500cecb392631e54b09dd6/) where you can see that 0.1234 BTC is sent to 2N6DrM8kBbeNAfL65JywN8VnjyZeLhp67ik (= the contract script hash) being a [p2sh](https://en.bitcoin.it/wiki/Pay_to_script_hash) address in the bitcoin testnet.

### audit contract

Bob sends Alice the contract and the contract transaction. Alice should now verify if

- the script is correct
- the locktime is far enough in the future
- the amount is correct
- she is the recipient

command:`btcatomicswap [-testnet] auditcontract <contract> <contract transaction>`

 ```sh
$./btcatomicswap -testnet --rpcuser=user --rpcpass=pass  -s  "localhost:7777" auditcontract 6382012088a8206036b8a2fedaadb621bf043395167e5ee1190f84261ac878ebd6fa38e100f98b8876a914cbed4f110fc43fe309975fd7c134608188ede5c06704de5cbd5db17576a914cbed4f110fc43fe309975fd7c134608188ede5c06888ac 020000000123b234d4a59281756edb30f8d1701478b74035d63eded22f4d0877b1f10a9181010000006a473044022021792701795cb7988f5a86eaf0eccd5a914abbd39731a7ad66508db361cf3139022031e085f3b3305925b38bfeb7fb4110ef8daccff65e3d4f026ac1e62fcbb4973301210206b803413788c63fb600208c8226f5f396b340d9f41f74c137f09a0c6d058a6efdffffff0295b12300000000001976a914375651c9d6e5d5e9be2540dbe5df310ee3b4537f88ac404b4c000000000017a9148e57279669d66bcc0d22c09a5eb60b75c958d49b87ba2d1800
Contract address:        2N6DrM8kBbeNAfL65JywN8VnjyZeLhp67ik
Contract value:          0.05 BTC
Recipient address:       mz7DkWBB9JQrReutRbtwU5JZK5WEvQs3Md
Author's refund address: mz7DkWBB9JQrReutRbtwU5JZK5WEvQs3Md

Secret hash: 6036b8a2fedaadb621bf043395167e5ee1190f84261ac878ebd6fa38e100f98b

Locktime: 2019-11-02 10:39:26 +0000 UTC
Locktime reached in 45h23m2s
```

WARNING:
A check on the blockchain should be done as the auditcontract does not do that so an already spent output could have been used as an input. Checking if the contract has been mined in a block should suffice

### Participate

Alice trusts the contract so she participates in the atomic swap by paying the XLMBTC into a new stellar holding account using the same secret hash as part of the signing conditions.  

Bob uses an existing Stellar account ( or creates a new one): *GADXVG3VLC7WQ5L3OQNFQ7XB7PPDSR3TVJ2PKJT4IEI5IJBXAZM7YUSL*

Bob has to make sure a trustline exists for his account for XLMBTC( BTC:GDPHIMRSUSZNLNFWW7VJWWQ2NCH6D6ZVJ4RIME3FUGZLJRS3KKNIVYQ5 ).

Bob sends this address to Alice who uses it to participate in the swap.
command:`stellaratomicswap [-testnet] -asset <code:issuer participate <participant seed> <initiator address> <amount> <secret hash>`

```sh
$ ./stellaratomicswap -testnet -asset BTC:GDPHIMRSUSZNLNFWW7VJWWQ2NCH6D6ZVJ4RIME3FUGZLJRS3KKNIVYQ5 participate SA76XPCY3SRDDPPBX2EUIH3D7TZ7BVWE7LVSBGNTV5ZYLNCVCOHGPD65 GADXVG3VLC7WQ5L3OQNFQ7XB7PPDSR3TVJ2PKJT4IEI5IJBXAZM7YUSL 0.05 6036b8a2fedaadb621bf043395167e5ee1190f84261ac878ebd6fa38e100f98b
participant address: GBOEIZUDHR3ZKWQIYNHDIAEQZ4WMSH7JJKTQMTFKZVALTPPTVY7LCRZV
holding account address: GAKEMMQCWNRSOONRTWJ7KWGXPSR3RPIANJO5DQ4HDVIIZYVKO5JD3PN4
refund transaction:
AAAAABRGMgKzYyc5sZ2T9VjXfKO4vQBqXdHDhx1QjOKqd1I9AAAAZAAASAwAAAACAAAAAQAAAABdvDFZAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAAFEYyArNjJzmxnZP1WNd8o7i9AGpd0cOHHVCM4qp3Uj0AAAAIAAAAAFxEZoM8d5VaCMNONACQzyzJH+lKpwZMqs1Aub3zrj6xAAAAAAAAAAA=
```

The above command creates a holding account with 0.05 XLMBTC on it with modified signing conditions, matching the atomic swap conditiuons. [On a testnet explorer one can view the details of the holding account](
https://testnet.steexp.com/).Testnet is reset every quarter so this transaction might not be visible anymore.

Alice now informs Bob that the Stellar contract account  has been created and provides him with the holding account address and the refund transaction.

### audit Stellar contract

Just as Alice had to audit Bob's contract, Bob now has to do the same with Alice's holding account before withdrawing.

Bob verifies if:

- the amount of tokens on the account is correct
- the locktime, hashed secret and wallet address defined in the signing conditions are correct

command:`stellaratomicswap [-tesnet] auditcontract holdingAccountAddress refundTransaction`
flags are available to automatically check the information in the contract.

```sh
$ ./stellaratomicswap -testnet auditcontract GBWN2WYYRGCT4PAWHDDFR35CRKO4QXFKME6ZQNUNESN4LMMPFKJ5VHYQ AAAAAGzdWxiJhT48FjjGWO+iip3IXKphPZg2jSSbxbGPKpPaAAAAZAAVx94AAAACAAAAAQAAAABdud1ZAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAAbN1bGImFPjwWOMZY76KKnchcqmE9mDaNJJvFsY8qk9oAAAAIAAAAAPKF25qDKjqowtn+HawwSBv+QBZllBeK3+ToJDhL5zasAAAAAAAAAAA=
Contract address:        GBWN2WYYRGCT4PAWHDDFR35CRKO4QXFKME6ZQNUNESN4LMMPFKJ5VHYQ
Contract value:
Amount: 0.0500000 Code: BTC Issuer: GAXT72C4PLIRVD4DI5IATBEIQXTV5XV4YNEALIF6WFALCOPUFYRDZBHL 
Amount: 9.9999600 XLM
Recipient address:       GB7CA4F3VLBJN5UXHFDARY62FKUB5C5UH244JSNWMDOJZ54WTEIK7XBV
Refund address: GDZILW42QMVDVKGC3H7B3LBQJAN74QAWMWKBPCW74TUCIOCL443KZJ72

Secret hash: f627b65f286b67e4bf906dc0a38b0711bfea146c40ed38ec957fda390d6133e4

Locktime: 2019-10-30 18:58:33 +0000 UTC
Locktime reached in 10h2m2s
```

### redeem tokens

Now that both Bob and Alice have paid into their respective contracts, Bob withdraws from the Stellar contract. This step involves publishing a transaction which reveals the secret to Alice, allowing her to withdraw from the Bitcoin contract.

command:`stellaratomicswap [-tesnet] redeem receiverseed  holdingAccountAddress secret`

```sh
$ ./stellaratomicswap -testnet redeem SCCSIKGU4F5JS4LY5QZ2OXYMTQTQVJF3TOA7PZVTDDPOPE7BXYAZOITZ GBWN2WYYRGCT4PAWHDDFR35CRKO4QXFKME6ZQNUNESN4LMMPFKJ5VHYQ afaee75c4dcf47094a3c2c62ee97200a14aeb643e6fa2095dc4b5385911501a6
***TransactionSuccess dump***
    Links: {{https://horizon-testnet.stellar.org/transactions/4eae3f7b85033a5e76ae4eefb3ba1997df81ca6ae9667d8836261b4c37d2be2f false}}
    Hash: 4eae3f7b85033a5e76ae4eefb3ba1997df81ca6ae9667d8836261b4c37d2be2f
    Ledger: 1233735
    Env: AAAAAEnpYsKdTFfwq/9dqZ8LqTnCLhDPRB5BU+Mrcx4d1bbkAAAAZAAS0pUAAAACAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAASeliwp1MV/Cr/12pnwupOcIuEM9EHkFT4ytzHh3VtuQAAAAIAAAAAH4gcLuqwpb2lzlGCOPaKqgei7Q+ucTJtmDcnPeWmRCvAAAAAAAAAALP3CQeAAAAILDZGggUk0oW+Nr8Yuda4DwYzHRIliiwus9iUvVwECTClpkQrwAAAEBvyl/3ZjsOwWavwTQwNLyiswe27fZbSbQyfUjv1XnYTvCMwyQdrVwIJ79uXi5nCyFHWo/tqf5zdL7x7PKoieMD
    Result: AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAIAAAAAAAAAAFR9VOMAAAAAA==
    Meta: AAAAAQAAAAIAAAADABLTRwAAAAAAAAAASeliwp1MV/Cr/12pnwupOcIuEM9EHkFT4ytzHh3VtuQAAAABUfVTjAAS0pUAAAABAAAAAwAAAAAAAAAAAAAAAAACAgIAAAADAAAAAH4gcLuqwpb2lzlGCOPaKqgei7Q+ucTJtmDcnPeWmRCvAAAAAQAAAAHWpnUj4pDcRbZM/CQT5Owpwmt9pytX8iuqJKH7NWGk8gAAAAIAAAACbDuK4yv8pyeCFIag95/x0mgA353xdjCZ6hiAmM/cJB4AAAABAAAAAAAAAAAAAAABABLTRwAAAAAAAAAASeliwp1MV/Cr/12pnwupOcIuEM9EHkFT4ytzHh3VtuQAAAABUfVTjAAS0pUAAAACAAAAAwAAAAAAAAAAAAAAAAACAgIAAAADAAAAAH4gcLuqwpb2lzlGCOPaKqgei7Q+ucTJtmDcnPeWmRCvAAAAAQAAAAHWpnUj4pDcRbZM/CQT5Owpwmt9pytX8iuqJKH7NWGk8gAAAAIAAAACbDuK4yv8pyeCFIag95/x0mgA353xdjCZ6hiAmM/cJB4AAAABAAAAAAAAAAAAAAABAAAABAAAAAMAEhg5AAAAAAAAAAB+IHC7qsKW9pc5Rgjj2iqoHou0PrnEybZg3Jz3lpkQrwAAABdIdugAABIYOQAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAEtNHAAAAAAAAAAB+IHC7qsKW9pc5Rgjj2iqoHou0PrnEybZg3Jz3lpkQrwAAABiabDuMABIYOQAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAMAEtNHAAAAAAAAAABJ6WLCnUxX8Kv/XamfC6k5wi4Qz0QeQVPjK3MeHdW25AAAAAFR9VOMABLSlQAAAAIAAAADAAAAAAAAAAAAAAAAAAICAgAAAAMAAAAAfiBwu6rClvaXOUYI49oqqB6LtD65xMm2YNyc95aZEK8AAAABAAAAAdamdSPikNxFtkz8JBPk7CnCa32nK1fyK6okofs1YaTyAAAAAgAAAAJsO4rjK/ynJ4IUhqD3n/HSaADfnfF2MJnqGICYz9wkHgAAAAEAAAAAAAAAAAAAAAIAAAAAAAAAAEnpYsKdTFfwq/9dqZ8LqTnCLhDPRB5BU+Mrcx4d1bbk
```

### redeem bitcoins

Now that Bob has withdrawn from the stellar contract and revealed the secret. If bob is really nice he could simply give the secret to Alice. However,even if he doesn't do this Alice can extract the secret from this redemption transaction. Alice may watch a block explorer to see when the stellar  holding account was merged or being spent fromand look up the redeeming transaction.

Alice can automatically extract the secret from the transaction where it is used by Bob, by simply giving the holding account address.

command:`stellaratomicswap [-tesnet] extractsecret holdingAccountAddress secretHash`

```sh
$./stellaratomicswap -testnet extractsecret GBE6SYWCTVGFP4FL75O2THYLVE44ELQQZ5CB4QKT4MVXGHQ52W3OJEJR 6c3b8ae32bfca727821486a0f79ff1d26800df9df1763099ea188098cfdc241e
Extracted secret: b0d91a0814934a16f8dafc62e75ae03c18cc74489628b0bacf6252f5701024c2
```

With the secret known, Alice may redeem from Bob's Bitcoin contract:

command: `btcatomicswap [-testnet] redeem <contract> <contract transaction> <secret>`

```sh
./btcatomicswap -testnet --rpcuser=user --rpcpass=pass  -s  "localhost:7777" redeem 6382012088a8206c3b8ae32bfca727821486a0f79ff1d26800df9df1763099ea188098cfdc241e8876a914c98094d0f6bbb95e6729539af1642f65d911099967040908ab5db17576a914c98094d0f6bbb95e6729539af1642f65d91109996888ac 0200000002d54a32316af9cf1fd2ccf94a9421519184467f8f1e8f4d030ad3fac911b8b00a000000006b4830450221008aceca8f93247644045590d72d3836cd7cc454c4ad9676b71ea545ce205cf61102205c1471209528df179cbaea2eef31739a52543a6b392c57a7b16de9131cefc51e0121035ceaa55c2bfb92c9c3266da89097fa4173d1561bd7b07c91d70b6ac1e01f0f94fdffffff71c5c5b8f1b28dd103d13197aaa961515ded308e52eebbf9df48e3714a08d623000000006b4830450221008b2e155816e217d29cc11a224c2bdfe394490a266f6c5b5b5cacb85062910c8f022011f844cf6ae5ef39f939bdc4d25993c83cb6d80b224605c8be36ec40848eed240121030b658c35df930aa8c17c94e8439e4a6daa3c025713997687eb481103aa3baa5dfdffffff024b069100000000001976a914dee4c7011fcf37ef36e91b76618dcf7f1e92488788ac1f4bbc000000000017a9140dad8d75de0448ba757e5337b3bea2b90f7e2acc8757281800 b0d91a0814934a16f8dafc62e75ae03c18cc74489628b0bacf6252f5701024c2
Redeem fee: 0.0000033 BTC (0.00001019 BTC/kB)

Redeem transaction (cff48a9997a178109b1fb907c2147682623d4bd11a8527449baa8f8f4012362d):
02000000010a78f18a3b18157e4b2aa0423a6748c97fd22222bc537755bed09734c475b18401000000ef473044022036026cc579cc8e3c7143eb6213c2a971ff36c31fb01b805ef4d248bb52170c0102205c029a2af93b7c54f4e1b29b5764164782986da525f7d1ba675a3a791efc26f00121020e0081ab8cfaee35f6080403efd17e377f493eec93b6a44181279419a913962a20b0d91a0814934a16f8dafc62e75ae03c18cc74489628b0bacf6252f5701024c2514c616382012088a8206c3b8ae32bfca727821486a0f79ff1d26800df9df1763099ea188098cfdc241e8876a914c98094d0f6bbb95e6729539af1642f65d911099967040908ab5db17576a914c98094d0f6bbb95e6729539af1642f65d91109996888acffffffff01d549bc00000000001976a914c98094d0f6bbb95e6729539af1642f65d911099988ac0908ab5d

Publish redeem transaction? [y/N] y
Published redeem transaction (cff48a9997a178109b1fb907c2147682623d4bd11a8527449baa8f8f4012362d)
```

## References

- [Electrum](https://electrum.org)
