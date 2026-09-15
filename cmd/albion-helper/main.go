package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/blackscorp/albion-helper/internal/ui"
)

func main() {
	a := app.NewWithID("de.blackscorp.albion-helper")
	w := ui.NewWindow(a)
	w.ShowAndRun()
}
