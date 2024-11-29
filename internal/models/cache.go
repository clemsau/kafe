package models

import "sort"

// Cache to store topic information
type TopicCache struct {
	topics           map[string]TopicInfo
	previousMessages map[string]int64
	order            []string
}

// NewTopicCache creates a new TopicCache
func NewTopicCache() *TopicCache {
	return &TopicCache{
		topics:           make(map[string]TopicInfo),
		previousMessages: make(map[string]int64),
		order:            make([]string, 0),
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

// GetSortedTopics returns all topics in sorted order
func (tc *TopicCache) GetSortedTopics() []TopicInfo {
	topics := make([]TopicInfo, 0, len(tc.order))
	for _, topicName := range tc.order {
		if info, exists := tc.topics[topicName]; exists {
			topics = append(topics, info)
		}
	}
	return topics
}

// GetOrder returns the order of topics
func (tc *TopicCache) GetOrder() []string {
	return tc.order
}

// GetPreviousMessages returns the previous message count for a topic
func (tc *TopicCache) GetPreviousMessages(topic string) (int64, bool) {
	count, exists := tc.previousMessages[topic]
	return count, exists
}

// SetPreviousMessages sets the previous message count for a topic
func (tc *TopicCache) SetPreviousMessages(topic string, count int64) {
	tc.previousMessages[topic] = count
}
