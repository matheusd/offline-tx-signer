package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/decred/dcrwallet/walletseed"
	"golang.org/x/crypto/ssh/terminal"
)

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}

// collapseSpace takes a string and replaces any repeated areas of whitespace
// with a single space character.
func collapseSpace(in string) string {
	whiteSpace := false
	out := ""
	for _, c := range in {
		if unicode.IsSpace(c) {
			if !whiteSpace {
				out = out + " "
			}
			whiteSpace = true
		} else {
			out = out + string(c)
			whiteSpace = false
		}
	}
	return out
}

func readSeedInput() []byte {
	var seedStr string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type the seed (33 words or hex seed) followed by an empty line.")
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		seedStr += " " + line
	}

	seedStrTrimmed := strings.TrimSpace(seedStr)
	seedStrTrimmed = collapseSpace(seedStrTrimmed)

	seed, err := walletseed.DecodeUserInput(seedStrTrimmed)
	orPanic(err)

	return seed
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

		fmt.Print("Confirm the private passphrase: ")
		confirm, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		orPanic(err)
		fmt.Print("\n")

		confirm = bytes.TrimSpace(confirm)
		if !bytes.Equal(pass, confirm) {
			fmt.Println("The typed passwords do not match")
			continue
		}

		return pass
	}
}

func zeroBytes(bts []byte) {
	for i := range bts {
		bts[i] = 0
	}
}
