package main

import "github.com/rivo/tview"

func main() {
	app := tview.NewApplication()
	topicList := tview.NewList().
		AddItem("Topic 1", "Description 1", 'a', nil).
		AddItem("Topic 2", "Description 2", 'b', nil).
		AddItem("Topic 3", "Description 3", 'c', nil).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		})

	topicList.SetBorder(true).SetTitle("Kafka Topics")

	if err := app.SetRoot(topicList, true).Run(); err != nil {
		panic(err)
	}
}
