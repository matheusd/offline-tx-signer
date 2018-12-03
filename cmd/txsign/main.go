package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/decred/dcrd/chaincfg"

	"github.com/matheusd/offline-tx-signer/pkg/offlinetxs"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("txsign inputfile")
		return
	}

	f, err := os.Open(os.Args[1])
	orPanic(err)
	defer f.Close()

	dec := json.NewDecoder(f)
	tx := offlinetxs.UnsignedTx{}
	err = dec.Decode(&tx)
	orPanic(err)

	passphrase := readPass()
	defer zeroBytes(passphrase)

	chainParams := &chaincfg.TestNet3Params
	seedBytes, err := offlinetxs.RetrieveLocalSeed("seed.dat", passphrase)
	orPanic(err)
	defer zeroBytes(seedBytes)

	signer, err := offlinetxs.NewSigner(tx.AddrIdxs, seedBytes, chainParams)
	orPanic(err)

	signed, err := signer.Sign(tx)
	orPanic(err)

	bts, err := signed.Bytes()
	orPanic(err)
	fmt.Println(hex.EncodeToString(bts))

}
