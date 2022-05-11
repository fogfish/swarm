package eventbridge

import (
	"errors"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

		start: lambda.Start,
	}
}

//
func (q *Dequeue) Mock(mock func(interface{})) {
	q.start = mock
}

//
// func (q *Recver) ID() string {
// 	return q.id
// }

//
func (q *Dequeue) Listen() error {
	go func() {
		logger.Notice("init aws eventbridge recv %s", q.id)

		// Note: this indirect synonym for lambda.Start
		q.start(
			func(evt events.CloudWatchEvent) error {
				logger.Debug("cloudwatch event %+v", evt)

				q.sock <- &swarm.Bag{
					Queue:    evt.Source,
					Category: evt.DetailType,
					Object:   evt.Detail,
					Digest:   evt.ID,
				}

				select {
				case <-q.sack:
					return nil
				case <-time.After(q.policy.TimeToFlight):
					return errors.New("message ack timeout: " + evt.ID)
				}
			},
		)
		q.sys.Close()
	}()

	return nil
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
