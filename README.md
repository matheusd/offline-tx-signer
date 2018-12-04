# Decred Offline Transaction Signer

** Do not use in production **

Proof of concept offline transaction signer. This has two basic parts:

A transaction generator (txgen) which will connect to your wallet via grpc and generate a simple transaction paying to an address (similar to a `sendtoaddress` call). This can even be called on watch only wallets.

A transaction signer, which will pick up this transaction, sign it and generate the signed blob for transmission.

The point of this is to run the signer in a different machine (eg: a raspberry pi) than the regular wallet.

You can use any way of exchanging transaction data that you feel comfortable with and deem secure enough (usb pendrive, network fileshare, serial connection etc).

## Using

I'm assuming you're running from source, given I haven't generated public binaries.

Also, this is hard coded to use testnet.

You might need packages (apt-get/dnf/etc):

```
libglib2.0-dev libcairo2-dev libgtk2.0-dev
```

### On the regular wallet machine

This could be a watch only wallet.

```shell
$ go run ./cmd/txgen [amount] [dest-address] [walletserver] [walletcert]

# for example, sending 0.1 dcr to the faucet return address

$ go run ./cmd/txgen 0.1 TsfDLrRkk9ciUuwfp2b8PawwnukYD7yAjGd localhost:19121 ~/.dcrwallet/rpc.cert > unsigned-tx.json
```

### On the signer wallet

```shell
# not really a create as much as an import. :shrug:
$ go run ./cmd/createseed


$ go run ./cmd/txsign [unsigned-tx-file]

# for example

$ go run ./cmd/txsign unsigned-tx.json > signed.hex

# Now publish the transaction (eg: on testnet.dcrdata.org)
```

#### GUI version

You can run the GUI version (optimized for a raspi with a 3.5" LCD). It will monitor a given dir for a file named `unsigned-tx.json` and produce a `signed-tx.hex` file.

Obviously, this is currently vulnerable to someone changing the tx data. This is just a POC after all.

```shell
$ go run ./cmd/txsigngui /home/user/media/drive2/tx
```
