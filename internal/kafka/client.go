package kafka

import (
	"fmt"

	"github.com/clemsau/kafe/internal/models"

	"github.com/IBM/sarama"
)

type Client struct {
	sarama.Client
	addresses []string
}

func NewClient(brokers []string, config *sarama.Config) (*Client, error) {
	if config == nil {
		config = sarama.NewConfig()
		config.Version = sarama.V2_8_0_0
	}

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	return &Client{
		Client:    client,
		addresses: brokers,
	}, nil
}

func (c *Client) GetAddresses() []string {
	return c.addresses
}

func (c *Client) GetTopicInfo(topic string) (models.TopicInfo, error) {
	partitions, err := c.Partitions(topic)
	if err != nil {
		return models.TopicInfo{}, err
	}

	info := models.TopicInfo{
		Name:       topic,
		Partitions: len(partitions),
	}

	if len(partitions) > 0 {
		replicas, err := c.Replicas(topic, partitions[0])
		if err == nil {
			info.Replicas = len(replicas)
		}
	}

	for _, partition := range partitions {
		oldest, err := c.GetOffset(topic, partition, sarama.OffsetOldest)
		if err != nil {
			continue
		}

		newest, err := c.GetOffset(topic, partition, sarama.OffsetNewest)
		if err != nil {
			continue
		}

		info.Messages += newest - oldest
	}

	info.Status = c.getTopicHealth(topic, partitions)
	return info, nil
}

func (c *Client) getTopicHealth(topic string, partitions []int32) string {
	for _, partition := range partitions {
		leader, err := c.Leader(topic, partition)
		if err != nil || leader == nil {
			return "Error"
		}

		isr, err := c.Replicas(topic, partition)
		if err != nil || len(isr) == 0 {
			return "Warning"
		}
	}
	return "Ready"
}
