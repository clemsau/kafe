package topics

import (
	"github.com/clemsau/kafe/internal/ui/consumer_groups"
	"github.com/clemsau/kafe/internal/ui/dialog"
	"github.com/clemsau/kafe/internal/ui/messages"
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
		switch event.Rune() {
		case '/':
			h.table.searchBar.Activate()
			return nil
		case 'g':
			selectedRow, _ := h.table.GetSelection()
			if selectedRow > 0 {
				topic := h.table.GetCell(selectedRow, 0).Text
				viewer := consumer_groups.NewGroupViewer(h.table.app, h.table.client, topic)
				h.table.app.AddPage("consumer-groups", viewer, true)
			}
			return nil
		}
	case tcell.KeyEnter:
		selectedRow, _ := h.table.GetSelection()
		if selectedRow > 0 {
			topic := h.table.GetCell(selectedRow, 0).Text
			viewer := messages.NewMessageViewer(h.table.app, h.table.client, topic)
			h.table.app.AddPage("messages", viewer, true)
			if err := viewer.Start(); err != nil {
				dialog.ShowError(h.table.app, err.Error())
				h.table.app.RemovePage("messages")
				return event
			}
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
