package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestNewWindowBuildsWalkingSkeleton(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	w := NewWindow(a)
	if w.Title() != "Albion Helper" {
		t.Fatalf("window title = %q, want Albion Helper", w.Title())
	}
	if w.Content() == nil {
		t.Fatal("window content is nil")
	}
}
