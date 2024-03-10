package events3

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/swarm"
	queue "github.com/fogfish/swarm/queue"
)

// const Category = "events3.Event"
const Category = "S3EventRecord"

// type Event swarm.Event[*events.S3EventRecord]

// func (Event) HKT1(swarm.EventType)       {}
// func (Event) HKT2(*events.S3EventRecord) {}

func Dequeue(q swarm.Broker) (<-chan swarm.Msg[*events.S3EventRecord], chan<- swarm.Msg[*events.S3EventRecord]) {
	return queue.Dequeue[*events.S3EventRecord](q)
	//return queue.Dequeue[*events.S3EventRecord, *Event](q)
}
