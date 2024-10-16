# Testing the project

You can spin up a kafka cluster with the docker-compose.yaml file by running `docker-compose up -d`

You can then start a consumer with:

```bash
kafka-console-consumer --bootstrap-server broker:9092 \
                       --topic example-topic \
                       --from-beginning
```

Create a new topic in the cluster with:

```bash
docker exec broker kafka-topics --bootstrap-server localhost:9092 --topic new_topic --create --partitions 3 --replication-factor 1
```

Delete a topic in the cluster with:

```bash
docker exec broker kafka-topics --bootstrap-server localhost:9092 --topic new_topic --delete
```

And start an interactive producer with:

```bash
docker exec --interactive --tty broker \
            kafka-console-producer --bootstrap-server localhost:9092 \
            --topic new_topic
```
