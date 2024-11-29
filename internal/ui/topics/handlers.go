package topics

import (
	"github.com/gdamore/tcell/v2"
)

// TopicTableHandler handles keyboard input for the topics table
type TopicTableHandler struct {
	table *Table
}

// NewTopicTableHandler creates a new topic table handler
func NewTopicTableHandler(table *Table) *TopicTableHandler {
	return &TopicTableHandler{
		table: table,
	}
}

// Handle processes keyboard events for the topics table
func (h *TopicTableHandler) Handle(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		if event.Rune() == '/' {
			h.table.searchBar.Activate()
			return nil
		}
	case tcell.KeyHome:
		h.table.Table.Select(1, 0)
		return nil
	case tcell.KeyEnd:
		h.table.Table.Select(h.table.Table.GetRowCount()-1, 0)
		return nil
	}

	return event
}
