package topics

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/clemsau/kafe/internal/kafka"
	"github.com/clemsau/kafe/internal/models"
	"github.com/clemsau/kafe/internal/ui"
	"github.com/clemsau/kafe/internal/ui/controls"
	"github.com/clemsau/kafe/internal/ui/topbar"
	"github.com/clemsau/kafe/internal/ui/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Table represents the topics table view
type Table struct {
	*tview.Table
	app        *ui.App
	client     *kafka.Client
	cache      *models.TopicCache
	updateChan chan []models.TopicInfo
	headers    []string
	searchBar  *SearchBar
	topBar     *topbar.TopBar
	layout     *tview.Flex
}

// NewTable creates a new topics table
func NewTable(app *ui.App, client *kafka.Client, cache *models.TopicCache) *tview.Flex {
	table := &Table{
		Table:      tview.NewTable().SetSelectable(true, false),
		app:        app,
		client:     client,
		cache:      cache,
		updateChan: make(chan []models.TopicInfo),
		headers: []string{
			"Topic",
			"Partitions",
			"Replicas",
			"Status",
			"Messages",
			"Throughput (msg/s)",
		},
	}

	table.searchBar = NewSearchBar(table)

	table.topBar = topbar.NewTopBar([]controls.Control{
		{Key: "Enter", Description: "view messages"},
		{Key: "g", Description: "consumer groups"},
		{Key: "/", Description: "search"},
		{Key: "Home/End", Description: "first/last"},
		{Key: "q", Description: "quit"},
	})

	table.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table.topBar, table.topBar.GetHeight(), 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(table.searchBar, 3, 0, false).
		AddItem(table, 0, 1, true)

	table.SetupUI()
	table.StartMonitoring()
	return table.layout
}

// SetupUI initializes the table UI
func (t *Table) SetupUI() {
	t.SetBorder(true).
		SetTitle("Topics").
		SetTitleAlign(tview.AlignLeft)

	// Set up headers
	for col, header := range t.headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetExpansion(1)
		t.SetCell(0, col, cell)
	}

	t.SetFixed(1, 0)
	t.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.ColorRoyalBlue).
		Foreground(tcell.ColorWhite))

	t.SetInputCapture(NewTopicTableHandler(t).Handle)
}

// UpdateTable updates the table with new topic information
func (t *Table) UpdateTable(topics []models.TopicInfo) {
	currentRow, _ := t.GetSelection()
	t.Clear()

	// Restore headers
	for col, header := range t.headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetExpansion(1)
		t.SetCell(0, col, cell)
	}

	// Update rows
	for row, topic := range topics {
		cells := []string{
			topic.Name,
			fmt.Sprintf("%d", topic.Partitions),
			fmt.Sprintf("%d", topic.Replicas),
			topic.Status,
			fmt.Sprintf("%d", topic.Messages),
			fmt.Sprintf("%.1f", topic.Throughput),
		}

		for col, content := range cells {
			cell := tview.NewTableCell(content).
				SetAlign(tview.AlignLeft).
				SetExpansion(1)

			if col == 3 { // Status column
				cell.SetTextColor(utils.GetStatusColor(topic.Status))
			}

			t.SetCell(row+1, col, cell)
		}
	}

	// Restore selection
	maxRow := len(topics)
	if currentRow > 0 && currentRow <= maxRow {
		t.Select(currentRow, 0)
	} else if maxRow > 0 {
		t.Select(1, 0)
	}
}

// StartMonitoring starts the topic monitoring goroutine
func (t *Table) StartMonitoring() {
	go t.monitorTopics()

	// Update handler
	go func() {
		for topics := range t.updateChan {
			func(info []models.TopicInfo) {
				t.app.QueueUpdateDraw(func() {
					t.UpdateTable(info)
				})
			}(topics)
		}
	}()
}

// monitorTopics periodically fetches topic information
func (t *Table) monitorTopics() {
	tickRateSeconds := 2
	ticker := time.NewTicker(time.Duration(tickRateSeconds) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		topics, err := t.client.Topics()
		if err != nil {
			continue
		}

		// Update visible topics
		_, _, _, height := t.GetInnerRect()
		rowOffset, _ := t.GetOffset()
		visibleRows := height - 1
		visibleStart := rowOffset + 1
		visibleEnd := visibleStart + visibleRows

		for i, topic := range topics {
			topicRow := i + 1 // account for header
			cachedInfo, exists := t.cache.Get(topic)
			if !exists {
				cachedInfo = models.TopicInfo{
					Name:       topic,
					LastUpdate: time.Time{},
				}
			}

			// Only fetch detailed info for visible topics or if never fetched
			if (topicRow >= visibleStart && topicRow <= visibleEnd) || cachedInfo.LastUpdate.IsZero() {
				info, err := t.client.GetTopicInfo(topic)
				if err != nil {
					log.Printf("Error fetching info for topic %s: %v", topic, err)
					continue
				}

				if prev, exists := t.cache.GetPreviousMessages(topic); exists {
					messageDelta := info.Messages - prev
					info.Throughput = float64(messageDelta) / float64(tickRateSeconds)
				}

				t.cache.SetPreviousMessages(topic, info.Messages)

				info.LastUpdate = time.Now()
				t.cache.UpsertTopic(info)
				continue
			}

			t.cache.UpsertTopic(cachedInfo)
		}

		filterText := t.searchBar.GetFilterText()
		if filterText != "" {
			filtered := make([]models.TopicInfo, 0)
			for _, topic := range t.cache.GetSortedTopics() {
				if strings.Contains(strings.ToLower(topic.Name), strings.ToLower(filterText)) {
					filtered = append(filtered, topic)
				}
			}
			t.updateChan <- filtered
			continue
		}
		t.updateChan <- t.cache.GetSortedTopics()
	}
}

// ApplyFilter filters the table based on search text
func (t *Table) ApplyFilter(filterText string) {
	topics := t.cache.GetSortedTopics()
	if filterText == "" {
		t.UpdateTable(topics)
		return
	}

	filtered := make([]models.TopicInfo, 0)
	for _, topic := range topics {
		if strings.Contains(strings.ToLower(topic.Name), strings.ToLower(filterText)) {
			filtered = append(filtered, topic)
		}
	}
	t.UpdateTable(filtered)
}
