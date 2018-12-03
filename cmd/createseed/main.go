package main

import (
	"fmt"

	"github.com/matheusd/offline-tx-signer/pkg/offlinetxs"
)

func main() {
	seed := readSeedInput()
	passphrase := readPass()
	defer zeroBytes(seed)
	defer zeroBytes(passphrase)

	err := offlinetxs.StoreLocalSeed("seed.dat", seed, passphrase)
	orPanic(err)

	fmt.Println("Saved seed.dat!")
}
