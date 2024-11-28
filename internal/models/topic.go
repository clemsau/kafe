package models

import "time"

type TopicInfo struct {
	Name       string
	Partitions int
	Replicas   int
	Status     string
	Messages   int64
	LastUpdate time.Time
}
