package kafka

import (
	"fmt"
	"time"

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

func (c *Client) GetConsumerGroups(topic string) ([]models.ConsumerGroupInfo, error) {
	coordinator, err := c.Coordinator("")
	if err != nil {
		return nil, fmt.Errorf("failed to get coordinator: %w", err)
	}
	defer coordinator.Close()

	req := &sarama.ListGroupsRequest{}
	resp, err := coordinator.ListGroups(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	var result []models.ConsumerGroupInfo

	for group := range resp.Groups {
		greq := &sarama.DescribeGroupsRequest{
			Groups: []string{group},
		}
		gresp, err := coordinator.DescribeGroups(greq)
		if err != nil {
			continue
		}

		if len(gresp.Groups) == 0 {
			continue
		}

		gdesc := gresp.Groups[0]

		consuming := false
		members := 0
		for _, member := range gdesc.Members {
			metadata, err := member.GetMemberMetadata()
			if err != nil {
				continue
			}

			for _, t := range metadata.Topics {
				if t == topic {
					consuming = true
					members++
					break
				}
			}
		}

		if !consuming {
			continue
		}

		var totalLag int64
		partitions := c.partitionsForTopic(topic)

		for _, partition := range partitions {
			offReq := &sarama.OffsetFetchRequest{
				Version:       1,
				ConsumerGroup: group,
			}
			offReq.AddPartition(topic, partition)

			offResp, err := coordinator.FetchOffset(offReq)
			if err != nil {
				continue
			}

			block := offResp.Blocks[topic][partition]
			if block.Err != sarama.ErrNoError {
				continue
			}

			committed := block.Offset
			newest, err := c.GetOffset(topic, partition, sarama.OffsetNewest)
			if err == nil && committed != -1 { // -1 indicates no committed offset
				totalLag += newest - committed
			}
		}

		status := "Active"
		if gdesc.State == "Dead" {
			status = "Dead"
		} else if totalLag > 1000 {
			status = "Lagging"
		}

		result = append(result, models.ConsumerGroupInfo{
			ID:         group,
			Topic:      topic,
			Members:    members,
			TotalLag:   totalLag,
			Status:     status,
			LastUpdate: time.Now(),
		})
	}

	return result, nil
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

func (c *Client) partitionsForTopic(topic string) []int32 {
	partitions, err := c.Partitions(topic)
	if err != nil {
		return []int32{}
	}
	return partitions
}
