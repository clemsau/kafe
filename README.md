# kafe

Kafe is a TUI tool which helps observe and manage Kafka clusters.

## Features

- [ ] Topic management
  - [ ] Listing topics
  - [ ] Viewing partition details
  - [ ] Leader/follower status
- [ ] Message inspection
  - [ ] Peak at messages in the topic
  - [ ] Filter
- [ ] Consumer group monitoring
  - [ ] List consumer groups
  - [ ] Track offsets and lag perpartition
- [ ] Broker health
  - [ ] Monitor brokers health
  - [ ] Performances
  - [ ] Partition reassignment infos
- [ ] Performances metrics
  - [ ] Throughput
  - [ ] Lag
  - [ ] Error rate
- [ ] Interactive query
  - [ ] Query topic
  - [ ] Query offset

## TODO

- [ ] Effective information fetching (fetch topics constantly, but information only of displayed topics)

## Ideas

- As constantly pulling all the infomation from the cluster might be expensive both for the client and the cluster, we can think of only pulling the topics information on startup. The informations for a given topic could also be fetched for the currently hovered topic. But also a watch list could be implemented to keep track of the topics we are interested in (e.g Shift + W to add currently hovered topic to watch list).

- Some standards warnings could be raised for certain topics for irrational behaviors (maybe let's do this in another tab). For example:
  - A topic with a single partition (which is a dangerous configuration)
  - Consumer group that is lagging only on a subset of partitions
