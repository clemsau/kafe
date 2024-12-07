package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App wraps the tview application and provides common UI functionality
type App struct {
	*tview.Application
	pages *tview.Pages
}

// NewApp creates a new UI application
func NewApp() *App {
	app := &App{
		Application: tview.NewApplication(),
		pages:       tview.NewPages(),
	}

	app.SetRoot(app.pages, true)
	return app
}

// AddPage adds a new page to the application
func (a *App) AddPage(name string, item tview.Primitive, resize bool) {
	a.pages.AddPage(name, item, true, true)
}

// RemovePage removes a page from the application
func (a *App) RemovePage(name string) {
	a.pages.RemovePage(name)
}

// SetGlobalInputHandler sets up global keyboard shortcuts
func (a *App) SetGlobalInputHandler(handler func(*tcell.EventKey) *tcell.EventKey) {
	a.Application.SetInputCapture(handler)
}

// DefaultGlobalHandler provides common keyboard shortcuts
func (a *App) DefaultGlobalHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q':
			a.Stop()
			return nil
		}
	}
	return event
}
