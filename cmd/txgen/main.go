package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/dcrutil"
)

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}
func usage() {
	fmt.Println("txgen amount address walletserver walletcert")
}

func main() {
	if len(os.Args) < 5 {
		usage()
		return
	}

	amountStr := os.Args[1]
	addrStr := os.Args[2]
	walletServer := os.Args[3]
	walletCert := os.Args[4]

	wc, err := connectToWallet(walletServer, walletCert, &chaincfg.TestNet3Params)
	orPanic(err)

	dst, err := dcrutil.DecodeAddress(addrStr)
	orPanic(err)

	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	orPanic(err)

	amount, err := dcrutil.NewAmount(amountFloat)
	orPanic(err)

	resTx, err := wc.genTx(0, []*destination{&destination{addr: dst, amount: amount}})
	if err != nil {
		panic(err)
	}

	resJson, err := json.Marshal(resTx)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(resJson))
}
