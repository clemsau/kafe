package dialog

import (
	"github.com/clemsau/kafe/internal/ui"
	"github.com/rivo/tview"
)

// ShowError displays an error modal dialog
func ShowError(app *ui.App, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.RemovePage("error")
		})

	app.AddPage("error", modal, true)
}
