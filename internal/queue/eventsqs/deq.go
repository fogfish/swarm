package eventsqs

import (
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
)

/*

Queue ...
*/
type Dequeue struct {
	id     string
	sys    swarm.System
	policy *swarm.Policy

	sock chan *swarm.Bag
	sack chan *swarm.Bag

	start func(interface{})
}

/*

New ...
*/
func NewDequeue(
	sys swarm.System,
	queue string,
	policy *swarm.Policy,
) *Dequeue {
	return &Dequeue{
		id:     queue,
		sys:    sys,
		policy: policy,

		sock: make(chan *swarm.Bag),
		sack: make(chan *swarm.Bag),
	}
}

//
func (q *Dequeue) Mock(mock func(interface{})) {
	q.start = mock
}

// func (q *Recver) ID() string {
// return q.id
// }

func (q *Dequeue) Listen() error {
	go func() {
		logger.Notice("init aws eventsqs recv %s", q.id)

		// Note: this indirect synonym for lambda.Start
		q.start(
			func(events events.SQSEvent) error {
				logger.Debug("cloudwatch event %+v", events)

				acks := map[string]bool{}

				for _, evt := range events.Records {
					acks[evt.ReceiptHandle] = false
					q.sock <- &swarm.Bag{
						Category: attr(&evt, "Category"),
						System:   attr(&evt, "System"),
						Queue:    attr(&evt, "Queue"),
						Object:   []byte(evt.Body),
						Digest:   evt.ReceiptHandle,
					}
				}

				for {
					select {
					case bag := <-q.sack:
						delete(acks, bag.Digest)
						if len(acks) == 0 {
							return nil
						}
					case <-time.After(q.policy.TimeToFlight):
						return fmt.Errorf("sqs message ack timeout")
					}
				}

			},
		)
		q.sys.Close()
	}()

	return nil
}

func attr(msg *events.SQSMessage, key string) string {
	val, exists := msg.MessageAttributes[key]
	if !exists {
		return ""
	}

	return *val.StringValue
}

//
//
func (q *Dequeue) Close() error {
	close(q.sock)
	close(q.sack)

	return nil
}

func (q *Dequeue) Deq() chan *swarm.Bag {
	return q.sock
}

func (q *Dequeue) Ack() chan *swarm.Bag {
	return q.sack
}
