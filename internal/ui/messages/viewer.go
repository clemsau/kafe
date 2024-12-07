package messages

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/clemsau/kafe/internal/kafka"
	"github.com/clemsau/kafe/internal/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MessageViewer struct {
	*tview.TextView
	app       *ui.App
	client    *kafka.Client
	topic     string
	consumers []sarama.PartitionConsumer
	cancel    context.CancelFunc
	mutex     sync.Mutex
}

func NewMessageViewer(app *ui.App, client *kafka.Client, topic string) *MessageViewer {
	mv := &MessageViewer{
		TextView: tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(true).
			SetWrap(true),
		app:    app,
		client: client,
		topic:  topic,
	}

	mv.setupUI()
	return mv
}

func (mv *MessageViewer) setupUI() {
	mv.SetBorder(true).
		SetTitle(fmt.Sprintf(" Messages - %s ", mv.topic))

	mv.SetInputCapture(mv.handleInput)
}

func (mv *MessageViewer) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		mv.Stop()
		mv.app.RemovePage("messages")
		return nil
	}
	return event
}

func (mv *MessageViewer) Start() error {
	partitions, err := mv.client.Partitions(mv.topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %w", err)
	}

	numPartitions := len(partitions)
	if numPartitions > 5 {
		rand.Shuffle(len(partitions), func(i, j int) {
			partitions[i], partitions[j] = partitions[j], partitions[i]
		})
		partitions = partitions[:5]
	}

	ctx, cancel := context.WithCancel(context.Background())
	mv.cancel = cancel

	config := sarama.NewConfig()
	consumer, err := sarama.NewConsumer(mv.client.GetAddresses(), config)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	for _, partition := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(mv.topic, partition, sarama.OffsetNewest)
		if err != nil {
			mv.writeMessage(fmt.Sprintf("Failed to consume partition %d: %v", partition, err), partition, true)
			continue
		}
		mv.consumers = append(mv.consumers, partitionConsumer)

		go func(pc sarama.PartitionConsumer, partition int32) {
			for {
				select {
				case msg := <-pc.Messages():
					mv.writeMessage(string(msg.Value), partition, false)
				case err := <-pc.Errors():
					mv.writeMessage(fmt.Sprintf("Partition %d, error: %v", partition, err), partition, true)
				case <-ctx.Done():
					return
				}
			}
		}(partitionConsumer, partition)
	}

	return nil
}

func (mv *MessageViewer) Stop() {
	if mv.cancel != nil {
		mv.cancel()
	}
	for _, consumer := range mv.consumers {
		consumer.Close()
	}
}

func (mv *MessageViewer) writeMessage(msg string, partition int32, err bool) {
	mv.mutex.Lock()
	defer mv.mutex.Unlock()

	color := "green"
	if err {
		color = "red"
	}

	mv.app.QueueUpdateDraw(func() {
		fmt.Fprintf(mv.TextView, "[%s]%s [partition-%d][-] %s\n",
			color,
			time.Now().Format("15:04:05"),
			partition,
			msg)
	})
}
