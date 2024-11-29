package topics

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SearchBar represents the topic search functionality
type SearchBar struct {
	*tview.InputField
	table      *Table
	filterText string
}

// NewSearchBar creates a new search bar for filtering topics
func NewSearchBar(table *Table) *SearchBar {
	search := &SearchBar{
		InputField: tview.NewInputField(),
		table:      table,
	}

	search.
		SetLabel("/").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetLabelColor(tcell.ColorYellow).
		SetBorder(true).
		SetTitle("Search").
		SetTitleAlign(tview.AlignLeft)

	// Handle real-time filtering as text is typed
	search.SetChangedFunc(func(text string) {
		search.filterText = text
		search.table.ApplyFilter(text)
	})

	search.InputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
			search.Deactivate()
			return nil
		}
		return event
	})

	return search
}

// Activate enables the search bar
func (s *SearchBar) Activate() {
	s.SetBorderColor(tcell.ColorYellow)
	s.table.app.SetFocus(s)
}

// Deactivate disables the search bar but maintains the filter
func (s *SearchBar) Deactivate() {
	s.SetBorderColor(tcell.ColorDefault)
	s.table.app.SetFocus(s.table)
}

// GetFilterText returns the current filter text
func (s *SearchBar) GetFilterText() string {
	return s.filterText
}
