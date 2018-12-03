package main

import (
	"bytes"
	"fmt"

	"github.com/mattn/go-gtk/gtk"
)

var (
	currentPassphrase string
	pwdPanel          *gtk.VPaned
	waitingPanel      *gtk.VPaned
)

func showPassphrasePanel() {
	waitingPanel.Hide()
	pwdPanel.Show()
}

func showWaitingPanel() {
	pwdPanel.Hide()
	waitingPanel.Show()
}

func buildUI() gtk.IWidget {
	vbox := gtk.NewVBox(false, 4)

	label := gtk.NewLabel("Password")
	label.SetAlignment(0, 0)
	vbox.PackStart(label, false, false, 2)

	pwdLabel := gtk.NewLabel("")
	label.SetAlignment(0, 0)
	vbox.PackStart(pwdLabel, false, false, 2)

	var lineBox *gtk.HBox

	newButton := func(label string) {
		button := gtk.NewButtonWithLabel(label)
		lineBox.PackStart(button, true, true, 2)
		button.Clicked(func() {
			currentPassphrase += label
			pwdLabel.SetText(string(bytes.Repeat([]byte("*"), len(currentPassphrase))))
		})
	}

	lineBox = gtk.NewHBox(false, 4)
	newButton("7")
	newButton("8")
	newButton("9")
	vbox.PackStart(lineBox, false, false, 2)

	lineBox = gtk.NewHBox(false, 4)
	newButton("4")
	newButton("5")
	newButton("6")
	vbox.PackStart(lineBox, false, false, 2)

	lineBox = gtk.NewHBox(false, 4)
	newButton("1")
	newButton("2")
	newButton("3")
	vbox.PackStart(lineBox, false, false, 2)

	lineBox = gtk.NewHBox(false, 4)
	newButton("0")
	vbox.PackStart(lineBox, false, false, 2)

	okButton := gtk.NewButtonWithLabel("OK")
	vbox.PackStart(okButton, false, false, 2)

	okButton.Clicked(func() {
		err := signFile(currentPassphrase)
		currentPassphrase = ""
		if err != nil {
			fmt.Println(err)
			pwdLabel.SetText(fmt.Sprintf("Error: %v", err.Error()))
		} else {
			pwdLabel.SetText("Success!")
		}
	})

	pwdPanel = gtk.NewVPaned()
	pwdPanel.Pack1(vbox, false, false)

	label = gtk.NewLabel("Waiting to detect TX file")
	label.SetAlignment(0, 0)
	label.ModifyFontEasy("DejaVu Serif 15")

	waitingPanel = gtk.NewVPaned()
	waitingPanel.Pack1(label, false, false)

	vbox = gtk.NewVBox(false, 0)
	vbox.PackStart(pwdPanel, false, false, 0)
	vbox.PackStart(waitingPanel, false, false, 0)

	topbox := gtk.NewVBox(false, 0)
	align := gtk.NewAlignment(0, 0, 1, 1)
	align.SetPadding(10, 10, 10, 10)
	align.Add(vbox)
	topbox.Add(align)
	topbox.ShowAll()

	pwdPanel.Hide()

	return topbox
}
