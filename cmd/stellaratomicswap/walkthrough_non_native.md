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

Bob initiates the process by using btcatomicswap to pay 0.05BTC into the Bitcoin contract using Alice's Bitcoin address, sending the contract transaction, and sharing the secret hash (not the secret), and contract's transaction with Alice. The refund transaction can not be sent until the locktime expires, but should be saved in case a refund is necessary.

command:`btcatomicswap [-testnet] initiate <participant address> <amount>`

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
$ ./stellaratomicswap -testnet auditcontract GAKEMMQCWNRSOONRTWJ7KWGXPSR3RPIANJO5DQ4HDVIIZYVKO5JD3PN4 AAAAABRGMgKzYyc5sZ2T9VjXfKO4vQBqXdHDhx1QjOKqd1I9AAAAZAAASAwAAAACAAAAAQAAAABdvDFZAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAAFEYyArNjJzmxnZP1WNd8o7i9AGpd0cOHHVCM4qp3Uj0AAAAIAAAAAFxEZoM8d5VaCMNONACQzyzJH+lKpwZMqs1Aub3zrj6xAAAAAAAAAAA=
Contract address:        GAKEMMQCWNRSOONRTWJ7KWGXPSR3RPIANJO5DQ4HDVIIZYVKO5JD3PN4
Contract value:
Amount: 0.0500000 Code: BTC Issuer: GDPHIMRSUSZNLNFWW7VJWWQ2NCH6D6ZVJ4RIME3FUGZLJRS3KKNIVYQ5
Amount: 9.9999600 XLM
Recipient address:       GADXVG3VLC7WQ5L3OQNFQ7XB7PPDSR3TVJ2PKJT4IEI5IJBXAZM7YUSL
Refund address: GBOEIZUDHR3ZKWQIYNHDIAEQZ4WMSH7JJKTQMTFKZVALTPPTVY7LCRZV

Secret hash: 6036b8a2fedaadb621bf043395167e5ee1190f84261ac878ebd6fa38e100f98b

Locktime: 2019-11-01 13:21:29 +0000 UTC
Locktime reached in 23h58m24s
```

### redeem tokens

Now that both Bob and Alice have paid into their respective contracts, Bob withdraws from the Stellar contract. This step involves publishing a transaction which reveals the secret to Alice, allowing her to withdraw from the Bitcoin contract.

command:`stellaratomicswap [-tesnet] redeem receiverseed  holdingAccountAddress secret`

```sh
/stellaratomicswap -testnet redeem SABFCYTLL6OGGMJDZYUYU43T54Q5JOHUKZ36ASUFOFJ4CHAA6I4BXOSP GAKEMMQCWNRSOONRTWJ7KWGXPSR3RPIANJO5DQ4HDVIIZYVKO5JD3PN4 72312e9f0f696f26f9cb391e222ba9aa43bb8d1f0c1d26fce98262319d988231
$ ***TransactionSuccess dump***
    Links: {{https://horizon-testnet.stellar.org/transactions/f34ed651c520f424587eef2a349f930d8ebc0e899d93f009c5ccef4775482c42 false}}
    Hash: f34ed651c520f424587eef2a349f930d8ebc0e899d93f009c5ccef4775482c42
    Ledger: 114011
    Env: AAAAABRGMgKzYyc5sZ2T9VjXfKO4vQBqXdHDhx1QjOKqd1I9AAABLAAASAwAAAADAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAwAAAAAAAAABAAAAAAd6m3VYv2h1e3QaWH7h+945R3OqdPUmfEER1CQ3Bln8AAAAAUJUQwAAAAAA3nQyMqSy1bS2t+qbWhpoj+H7NU8ihhNlobK0xltSmooAAAAAAAehIAAAAAEAAAAAFEYyArNjJzmxnZP1WNd8o7i9AGpd0cOHHVCM4qp3Uj0AAAAGAAAAAUJUQwAAAAAA3nQyMqSy1bS2t+qbWhpoj+H7NU8ihhNlobK0xltSmooAAAAAAAAAAAAAAAEAAAAAFEYyArNjJzmxnZP1WNd8o7i9AGpd0cOHHVCM4qp3Uj0AAAAIAAAAAAd6m3VYv2h1e3QaWH7h+945R3OqdPUmfEER1CQ3Bln8AAAAAAAAAALhAPmLAAAAIHIxLp8PaW8m+cs5HiIrqapDu40fDB0m/OmCYjGdmIIxNwZZ/AAAAEAxznQXESNMEdvdlvQsJQPcK8IoHIV3s1s6ep8LtwksGw9mN2dcLt9PwpR0DaL8Aqly6xI3EK6sUFZWgHmzFm4F
    Result: AAAAAAAAASwAAAAAAAAAAwAAAAAAAAABAAAAAAAAAAAAAAAGAAAAAAAAAAAAAAAIAAAAAAAAAAAF9d18AAAAAA==
    Meta: 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
```

### redeem bitcoins

Now that Bob has withdrawn from the stellar contract and revealed the secret. If bob is really nice he could simply give the secret to Alice. However,even if he doesn't do this Alice can extract the secret from this redemption transaction. Alice may watch a block explorer to see when the stellar  holding account was merged or being spent fromand look up the redeeming transaction.

Alice can automatically extract the secret from the transaction where it is used by Bob, by simply giving the holding account address.

command:`stellaratomicswap [-tesnet] extractsecret holdingAccountAddress secretHash`

```sh
$./stellaratomicswap -testnet extractsecret GAKEMMQCWNRSOONRTWJ7KWGXPSR3RPIANJO5DQ4HDVIIZYVKO5JD3PN4 6036b8a2fedaadb621bf043395167e5ee1190f84261ac878ebd6fa38e100f98b
Extracted secret: 72312e9f0f696f26f9cb391e222ba9aa43bb8d1f0c1d26fce98262319d988231
```

With the secret known, Alice may redeem from Bob's Bitcoin contract:

command: `btcatomicswap [-testnet] redeem <contract> <contract transaction> <secret>`

```sh
./btcatomicswap -testnet --rpcuser=user --rpcpass=pass  -s  "localhost:7777" redeem 6382012088a8206036b8a2fedaadb621bf043395167e5ee1190f84261ac878ebd6fa38e100f98b8876a914cbed4f110fc43fe309975fd7c134608188ede5c06704de5cbd5db17576a914cbed4f110fc43fe309975fd7c134608188ede5c06888ac 020000000123b234d4a59281756edb30f8d1701478b74035d63eded22f4d0877b1f10a9181010000006a473044022021792701795cb7988f5a86eaf0eccd5a914abbd39731a7ad66508db361cf3139022031e085f3b3305925b38bfeb7fb4110ef8daccff65e3d4f026ac1e62fcbb4973301210206b803413788c63fb600208c8226f5f396b340d9f41f74c137f09a0c6d058a6efdffffff0295b12300000000001976a914375651c9d6e5d5e9be2540dbe5df310ee3b4537f88ac404b4c000000000017a9148e57279669d66bcc0d22c09a5eb60b75c958d49b87ba2d1800 72312e9f0f696f26f9cb391e222ba9aa43bb8d1f0c1d26fce98262319d988231
Redeem fee: 0.0000033 BTC (0.00001015 BTC/kB)

Redeem transaction (f06b3ab18e2717df6d284a9e28c2d0d0812d3ec7813322629a10830205e0fa0b):
0200000001d69db0541e6392b3ec0c506d05f25bd0146a3f44b5d1a4645f1b9ae460f527bf01000000f0483045022100e4cef29d848fe38a7a8cbd913e0026d6c30b108ffdc03ddd93b9318463148099022029e47adf62b90e1f0b24adc69c7e8765e624e86be326aec06a50cf0865fc63d40121028f02b7f1f0614d179c061cf84aa3cdb0332983764661fcd1228b043f94b8a0c62072312e9f0f696f26f9cb391e222ba9aa43bb8d1f0c1d26fce98262319d988231514c616382012088a8206036b8a2fedaadb621bf043395167e5ee1190f84261ac878ebd6fa38e100f98b8876a914cbed4f110fc43fe309975fd7c134608188ede5c06704de5cbd5db17576a914cbed4f110fc43fe309975fd7c134608188ede5c06888acffffffff01f6494c00000000001976a914cbed4f110fc43fe309975fd7c134608188ede5c088acde5cbd5d

Publish redeem transaction? [y/N] y
Published redeem transaction (f06b3ab18e2717df6d284a9e28c2d0d0812d3ec7813322629a10830205e0fa0b)
```

## References

- [Electrum](https://electrum.org)
