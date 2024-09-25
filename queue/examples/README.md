# `swarm` library examples

## Examples about queueing systems and event brokers

- AWS EventBridge
  - [produce message](./eventbridge/enqueue/eventbridge.go)
  - [consume message](./eventbridge/dequeue/eventbridge.go)
  - [serverless app](./eventbridge/serverless/main.go)
- AWS SQS Serverless
  - [produce message](./eventsqs/enqueue/eventsqs.go)
  - [consume message](./eventsqs/dequeue/eventsqs.go)
  - [serverless app](./eventsqs/serverless/main.go)
- AWS SQS
  - [produce message](examples/sqs/enqueue/sqs.go)
  - [consume message](examples/sqs/dequeue/sqs.go)
- AWS S3 Serverless
  - [consume message](./events3/dequeue/ddbstream.go)
  - [serverless app](./events3/serverless/main.go)
- AWS DynamoDB Stream Serverless
  - [consume message](./eventddb/dequeue/ddbstream.go)
  - [serverless app](./eventddb/serverless/main.go)

## Examples about different data types

- Bytes
  - [produce message](./bytes/enqueue/bytes.go)
  - [consume message](./bytes/dequeue/bytes.go)
- Generic Events
  - [produce message](./events/enqueue/events.go)
  - [consume message](./events/dequeue/events.go)
