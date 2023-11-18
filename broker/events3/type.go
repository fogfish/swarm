package events3

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/swarm"
	queue "github.com/fogfish/swarm/queue/events"
)

const Category = "events3:Event"

type Event swarm.Event[*events.S3EventRecord]

func (Event) HKT1(swarm.EventType)       {}
func (Event) HKT2(*events.S3EventRecord) {}

func Dequeue(q swarm.Broker) (<-chan *Event, chan<- *Event) {
	return queue.Dequeue[*events.S3EventRecord, Event](q)
}
