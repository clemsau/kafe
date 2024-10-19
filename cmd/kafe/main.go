package main

import (
	"log"
	"sort"
	"time"

	"github.com/IBM/sarama"
	"github.com/rivo/tview"
)

func main() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0

	client, err := sarama.NewClient([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	defer client.Close()

	app := tview.NewApplication()
	topicList := tview.NewList().
		SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {})

	topicList.AddItem("Fetching topics...", "", 0, nil)
	topicList.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})

	topicList.SetBorder(true).SetTitle("Kafka Topics")

	// Channel to send topic updates
	updateChan := make(chan []string)

	// Start Kafka monitoring in a separate goroutine
	go monitorTopics(client, updateChan)

	// Handle updates in the main thread
	go func() {
		for topics := range updateChan {
			func(t []string) {
				app.QueueUpdateDraw(func() {
					topicList.Clear()
					for _, topic := range t {
						topicList.AddItem(topic, "", 0, nil)
					}
					topicList.AddItem("Quit", "Press to exit", 'q', func() {
						app.Stop()
					})
				})
			}(topics)
		}
	}()

	if err := app.SetRoot(topicList, true).SetFocus(topicList).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}

func monitorTopics(client sarama.Client, updateChan chan<- []string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var currentTopics []string

	for ; ; <-ticker.C {
		topics, err := client.Topics()
		if err != nil {
			log.Printf("Error fetching topics: %v", err)
			continue
		}

		sort.Strings(topics)

		if !equal(topics, currentTopics) {
			currentTopics = topics
			updateChan <- topics
		}
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
