package main

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/IBM/sarama"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TopicInfo struct {
	Name       string
	Partitions int
	Replicas   int
	Status     string
	Messages   int64
	Size       int64
	LastUpdate time.Time
}

// Cache to store topic information
type TopicCache struct {
	topics map[string]TopicInfo
	order  []string
}

// NewTopicCache creates a new TopicCache
func NewTopicCache() *TopicCache {
	return &TopicCache{
		topics: make(map[string]TopicInfo),
		order:  make([]string, 0),
	}
}

// UpsertTopic updates or inserts a topic
func (tc *TopicCache) UpsertTopic(info TopicInfo) {
	if _, ok := tc.topics[info.Name]; !ok {
		tc.order = append(tc.order, info.Name)
		sort.Strings(tc.order)
	}
	tc.topics[info.Name] = info
}

// Get returns a topic by name if it exists
func (tc *TopicCache) Get(name string) (TopicInfo, bool) {
	info, exists := tc.topics[name]
	return info, exists
}

func main() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0

	client, err := sarama.NewClient([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	defer client.Close()

	app := tview.NewApplication()
	table := tview.NewTable().
		SetSelectable(true, false). // Enable row selection, disable column selection
		SetSelectedStyle(tcell.StyleDefault.
			Background(tcell.ColorRoyalBlue).
			Foreground(tcell.ColorWhite))

	headers := []string{
		"Topic",
		"Partitions",
		"Replicas",
		"Status",
		"Messages",
		"Size",
	}

	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetExpansion(1)
		table.SetCell(0, col, cell)
	}

	table.SetFixed(1, 0)
	table.SetBorder(true).
		SetTitle("Kafka Topics").
		SetTitleAlign(tview.AlignLeft)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()
		rowCount := table.GetRowCount()

		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'j': // vim-style down
				if row < rowCount-1 {
					table.Select(row+1, 0)
				}
				return nil
			case 'k': // vim-style up
				if row > 1 {
					table.Select(row-1, 0)
				}
				return nil
			case 'g': // vim-style top
				table.Select(1, 0)
				return nil
			case 'G': // vim-style bottom
				table.Select(rowCount-1, 0)
				return nil
			}
		case tcell.KeyHome:
			table.Select(1, 0)
			return nil
		case tcell.KeyEnd:
			table.Select(rowCount-1, 0)
			return nil
		}
		return event
	})

	cache := NewTopicCache()
	updateChan := make(chan []TopicInfo)

	go monitorTopics(client, cache, table, updateChan)

	go func() {
		for topicsInfo := range updateChan {
			func(info []TopicInfo) {
				app.QueueUpdateDraw(func() {
					currentRow, _ := table.GetSelection()
					table.Clear()

					for col, header := range headers {
						cell := tview.NewTableCell(header).
							SetTextColor(tcell.ColorYellow).
							SetSelectable(false).
							SetAlign(tview.AlignLeft).
							SetExpansion(1)
						table.SetCell(0, col, cell)
					}

					for row, topic := range info {
						cells := []string{
							topic.Name,
							fmt.Sprintf("%d", topic.Partitions),
							fmt.Sprintf("%d", topic.Replicas),
							topic.Status,
							fmt.Sprintf("%d", topic.Messages),
							formatSize(topic.Size),
						}

						for col, content := range cells {
							cell := tview.NewTableCell(content).
								SetAlign(tview.AlignLeft).
								SetExpansion(1)

							if col == 3 {
								cell.SetTextColor(getStatusColor(topic.Status))
							}

							table.SetCell(row+1, col, cell)
						}
					}

					maxRow := len(info)
					if currentRow > 0 && currentRow <= maxRow {
						table.Select(currentRow, 0)
					} else if maxRow > 0 {
						table.Select(1, 0)
					}
				})
			}(topicsInfo)
		}
	}()

	if err := app.SetRoot(table, true).SetFocus(table).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}

func monitorTopics(client sarama.Client, cache *TopicCache, table *tview.Table, updateChan chan<- []TopicInfo) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		topics, err := client.Topics()
		if err != nil {
			log.Printf("Error fetching topics: %v", err)
			continue
		}

		sort.Strings(topics)

		_, _, visibleFrom, visibleTo := table.GetInnerRect()
		rowOffset, _ := table.GetOffset()
		visibleStart := rowOffset + 1 // +1 to account for header
		visibleEnd := visibleStart + (visibleTo - visibleFrom)

		for i, topic := range topics {
			topicRow := i + 1 // +1 to account for header
			cachedInfo, exists := cache.Get(topic)
			if !exists {
				cachedInfo = TopicInfo{
					Name:       topic,
					LastUpdate: time.Time{},
				}
			}

			// Only fetch detailed info for visible topics or if never fetched
			if (topicRow >= visibleStart && topicRow <= visibleEnd) || cachedInfo.LastUpdate.IsZero() {
				info, err := getTopicInfo(client, topic)
				if err != nil {
					log.Printf("Error fetching info for topic %s: %v", topic, err)
					continue
				}
				info.LastUpdate = time.Now()
				cache.UpsertTopic(info)
			} else {
				cache.UpsertTopic(cachedInfo)
			}
		}

		topicsInfo := make([]TopicInfo, 0, len(cache.order))
		for _, topicName := range cache.order {
			if info, exists := cache.Get(topicName); exists {
				topicsInfo = append(topicsInfo, info)
			}
		}

		updateChan <- topicsInfo
	}
}

func getTopicInfo(client sarama.Client, topic string) (TopicInfo, error) {
	partitions, err := client.Partitions(topic)
	if err != nil {
		return TopicInfo{}, err
	}

	info := TopicInfo{
		Name:       topic,
		Partitions: len(partitions),
		Messages:   0,
		Size:       0,
	}

	if len(partitions) > 0 {
		replicas, err := client.Replicas(topic, partitions[0])
		if err == nil {
			info.Replicas = len(replicas)
		}
	}

	// Calculate total messages and size across all partitions
	for _, partition := range partitions {
		oldest, err := client.GetOffset(topic, partition, sarama.OffsetOldest)
		if err != nil {
			continue
		}

		newest, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
		if err != nil {
			continue
		}

		info.Messages += newest - oldest
		info.Size += (newest - oldest) * 1024 // Assuming average message size of 1KB
	}

	info.Status = getTopicHealth(client, topic, partitions)
	return info, nil
}

func getTopicHealth(client sarama.Client, topic string, partitions []int32) string {
	for _, partition := range partitions {
		leader, err := client.Leader(topic, partition)
		if err != nil || leader == nil {
			return "Error"
		}

		isr, err := client.Replicas(topic, partition)
		if err != nil || len(isr) == 0 {
			return "Warning"
		}
	}
	return "Ready"
}

func getStatusColor(status string) tcell.Color {
	switch status {
	case "Ready":
		return tcell.ColorGreen
	case "Warning":
		return tcell.ColorYellow
	case "Error":
		return tcell.ColorRed
	default:
		return tcell.ColorWhite
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
