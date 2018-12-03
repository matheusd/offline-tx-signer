package main

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func readPass() []byte {
	for {
		fmt.Print("Type the private passphrase: ")
		var pass []byte
		var err error
		pass, err = terminal.ReadPassword(int(os.Stdin.Fd()))
		orPanic(err)
		fmt.Print("\n")

		pass = bytes.TrimSpace(pass)
		if len(pass) == 0 {
			return nil
		}

		return pass
	}
}

func zeroBytes(bts []byte) {
	for i := range bts {
		bts[i] = 0
	}
}
