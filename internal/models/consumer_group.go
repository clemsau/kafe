package models

import "time"

type ConsumerGroupInfo struct {
	ID         string
	Topic      string
	Members    int
	TotalLag   int64
	Status     string
	LastUpdate time.Time
}

// Cache to store consumer group information
type ConsumerGroupCache struct {
	groups map[string]ConsumerGroupInfo
	order  []string
}

func NewConsumerGroupCache() *ConsumerGroupCache {
	return &ConsumerGroupCache{
		groups: make(map[string]ConsumerGroupInfo),
		order:  make([]string, 0),
	}
}

func (c *ConsumerGroupCache) UpsertGroup(info ConsumerGroupInfo) {
	if _, exists := c.groups[info.ID]; !exists {
		c.order = append(c.order, info.ID)
	}
	c.groups[info.ID] = info
}

func (c *ConsumerGroupCache) Get(id string) (ConsumerGroupInfo, bool) {
	info, exists := c.groups[id]
	return info, exists
}

func (c *ConsumerGroupCache) GetSortedGroups() []ConsumerGroupInfo {
	groups := make([]ConsumerGroupInfo, 0, len(c.order))
	for _, id := range c.order {
		if info, exists := c.groups[id]; exists {
			groups = append(groups, info)
		}
	}
	return groups
}
