package consumer_groups

import (
	"fmt"
	"strings"
	"time"

	"github.com/clemsau/kafe/internal/kafka"
	"github.com/clemsau/kafe/internal/models"
	"github.com/clemsau/kafe/internal/ui"
	"github.com/clemsau/kafe/internal/ui/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type GroupViewer struct {
	*tview.Table
	app        *ui.App
	client     *kafka.Client
	topic      string
	cache      *models.ConsumerGroupCache
	updateChan chan []models.ConsumerGroupInfo
	headers    []string
	searchBar  *tview.InputField
	layout     *tview.Flex
}

func NewGroupViewer(app *ui.App, client *kafka.Client, topic string) *tview.Flex {
	viewer := &GroupViewer{
		Table:      tview.NewTable().SetSelectable(true, false),
		app:        app,
		client:     client,
		topic:      topic,
		cache:      models.NewConsumerGroupCache(),
		updateChan: make(chan []models.ConsumerGroupInfo),
		headers: []string{
			"Group ID",
			"Members",
			"Total Lag",
			"Status",
		},
	}

	viewer.searchBar = tview.NewInputField()
	viewer.searchBar.
		SetLabel("/").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetLabelColor(tcell.ColorYellow).
		SetBorder(true).
		SetTitle("Search").
		SetTitleAlign(tview.AlignLeft)

	viewer.searchBar.SetChangedFunc(func(text string) {
		viewer.applyFilter(text)
	})

	viewer.searchBar.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
			viewer.searchBar.SetBorderColor(tcell.ColorDefault)
			viewer.app.SetFocus(viewer)
			return nil
		}
		return event
	})

	viewer.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(viewer.searchBar, 3, 0, false).
		AddItem(viewer, 0, 1, true)

	viewer.setupUI()
	viewer.startMonitoring()
	return viewer.layout
}

func (v *GroupViewer) setupUI() {
	v.SetBorder(true).
		SetTitle(fmt.Sprintf(" Consumer Groups - %s ", v.topic)).
		SetTitleAlign(tview.AlignLeft)

	for col, header := range v.headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetExpansion(1)
		v.SetCell(0, col, cell)
	}

	v.SetFixed(1, 0)
	v.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.ColorRoyalBlue).
		Foreground(tcell.ColorWhite))

	v.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			v.app.RemovePage("consumer-groups")
			return nil
		case tcell.KeyRune:
			if event.Rune() == '/' {
				v.searchBar.SetBorderColor(tcell.ColorYellow)
				v.app.SetFocus(v.searchBar)
				return nil
			}
		}
		return event
	})
}

func (v *GroupViewer) updateTable(groups []models.ConsumerGroupInfo) {
	v.Clear()

	for col, header := range v.headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetExpansion(1)
		v.SetCell(0, col, cell)
	}

	for row, group := range groups {
		cells := []string{
			group.ID,
			fmt.Sprintf("%d", group.Members),
			fmt.Sprintf("%d", group.TotalLag),
			group.Status,
		}

		for col, content := range cells {
			cell := tview.NewTableCell(content).
				SetAlign(tview.AlignLeft).
				SetExpansion(1)

			if col == 3 {
				cell.SetTextColor(utils.GetConsumerGroupStatusColor(group.Status))
			}

			v.SetCell(row+1, col, cell)
		}
	}
}

func (v *GroupViewer) startMonitoring() {
	go v.monitorGroups()

	go func() {
		for groups := range v.updateChan {
			func(info []models.ConsumerGroupInfo) {
				v.app.QueueUpdateDraw(func() {
					v.updateTable(info)
				})
			}(groups)
		}
	}()
}

func (v *GroupViewer) monitorGroups() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	v.fetchAndUpdateGroups()

	for range ticker.C {
		v.fetchAndUpdateGroups()
	}
}

func (v *GroupViewer) fetchAndUpdateGroups() {
	groups, err := v.client.GetConsumerGroups(v.topic)
	if err != nil {
		return
	}

	for _, group := range groups {
		v.cache.UpsertGroup(group)
	}

	v.updateChan <- v.cache.GetSortedGroups()
}

func (v *GroupViewer) applyFilter(filterText string) {
	groups := v.cache.GetSortedGroups()
	if filterText == "" {
		v.updateTable(groups)
		return
	}

	filtered := make([]models.ConsumerGroupInfo, 0)
	for _, group := range groups {
		if strings.Contains(strings.ToLower(group.ID), strings.ToLower(filterText)) {
			filtered = append(filtered, group)
		}
	}
	v.updateTable(filtered)
}
