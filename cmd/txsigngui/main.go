package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/matheusd/offline-tx-signer/pkg/offlinetxs"

	"github.com/decred/dcrd/chaincfg"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
)

var unsignedTxFile string
var signedTxFile string

func zeroBytes(bts []byte) {
	for i := range bts {
		bts[i] = 0
	}
}

func checkFile() bool {
	// this should really be done with fsnotify, but I don't care about it right
	// now.

	if _, err := os.Stat(unsignedTxFile); err != nil {
		// don't have unsigned tx file yet
		showWaitingPanel()
		return true
	}

	f, err := os.Open(unsignedTxFile)
	if err != nil {
		fmt.Println(err)
		showWaitingPanel()
		return true
	}

	dec := json.NewDecoder(f)
	tx := offlinetxs.UnsignedTx{}
	err = dec.Decode(&tx)
	if err != nil {
		fmt.Println(err)
		f.Close()
		showWaitingPanel()
		return true
	}

	f.Close()

	showPassphrasePanel()

	return true
}

func signFile(passphrase string) error {

	f, err := os.Open(unsignedTxFile)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(f)
	tx := offlinetxs.UnsignedTx{}
	err = dec.Decode(&tx)
	if err != nil {
		f.Close()
		return err
	}
	f.Close()

	chainParams := &chaincfg.TestNet3Params
	seedBytes, err := offlinetxs.RetrieveLocalSeed("seed.dat", []byte(passphrase))
	if err != nil {
		return err
	}
	defer zeroBytes(seedBytes)


	signer, err := offlinetxs.NewSigner(tx.AddrIdxs, seedBytes, chainParams)
	if err != nil {
		return err
	}

	signed, err := signer.Sign(tx)
	if err != nil {
		return err
	}

	bts, err := signed.Bytes()
	if err != nil {
		return err
	}
	encoded := hex.EncodeToString(bts)
	fmt.Println(encoded)

	f, err = os.Create(signedTxFile)
	if err != nil {
		return err
	}
	f.Write([]byte(encoded))
	f.Close()

	go func () {
		time.Sleep(time.Second * 2)
		os.Rename(unsignedTxFile, unsignedTxFile+".processed")
	}()

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: txsigngui [dir]")
		fmt.Println("will monitor [dir]/unsigned-tx.json and produce [dir]/signed-tx.hex")
		return
	}

	gtk.Init(nil)
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("Tx Signer")
	window.SetIconName("gtk-dialog-info")
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		gtk.MainQuit()
	}, "foo")

	ui := buildUI()

	window.Add(ui)
	window.SetSizeRequest(320, 400)
	window.Show()

	unsignedTxFile = os.Args[1] + "/unsigned-tx.json"
	signedTxFile = os.Args[1] + "/signed-tx.hex"

	glib.TimeoutAdd(1000, checkFile)

	gtk.Main()
}
