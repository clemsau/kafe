package main

import (
	"log"

	"github.com/clemsau/kafe/internal/kafka"
	"github.com/clemsau/kafe/internal/models"
	"github.com/clemsau/kafe/internal/ui"
	"github.com/clemsau/kafe/internal/ui/topics"
)

func main() {
	client, err := kafka.NewClient([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	defer client.Close()

	cache := models.NewTopicCache()

	app := ui.NewApp()
	layout := topics.NewTable(app, client, cache)

	app.AddPage("topics", layout, true)

	app.SetGlobalInputHandler(app.DefaultGlobalHandler)

	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
