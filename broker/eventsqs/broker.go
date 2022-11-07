package eventsqs

import (
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/internal/router"
)

// New creates broker for AWS SQS
func New(queue string, opts ...swarm.Option) (swarm.Broker, error) {
	conf := swarm.NewConfig()
	for _, opt := range opts {
		opt(conf)
	}

	bro, err := sqs.New(queue, opts...)
	if err != nil {
		return nil, err
	}

	return &broker{
		Broker: bro,
		config: conf,
		router: router.New(nil),
	}, nil
}

type broker struct {
	swarm.Broker
	config *swarm.Config
	router *router.Router
}

func (b *broker) Dequeue(category string, channel swarm.Channel) (swarm.Dequeue, error) {
	b.Broker.Dequeue(category, channel)
	b.router.Register(category)

	return b.router, nil
}

func (b *broker) Await() {
	lambda.Start(
		func(events events.SQSEvent) error {
			for _, evt := range events.Records {
				bag := swarm.Bag{
					Category: attr(&evt, "Category"),
					Object:   []byte(evt.Body),
					Digest:   evt.ReceiptHandle,
				}
				if err := b.router.Dispatch(bag); err != nil {
					return err
				}
			}

			return b.router.Await(1 * time.Second /*q.policy.TimeToFlight*/)
		},
	)
}

// func (b *broker) Dequeue(string, swarm.Channel) (swarm.Dequeue, error) {

// }

// func (b *broker) Await() {
// 	lambda.Start(
// 		func(events events.SQSEvent) error {
// 			acks := map[string]bool{}

// 			for _, evt := range events.Records {
// 				acks[evt.ReceiptHandle] = false
// 				q.sock <- &swarm.Bag{
// 					Category: attr(&evt, "Category"),
// 					Queue:    attr(&evt, "Queue"),
// 					Object:   []byte(evt.Body),
// 					Digest:   evt.ReceiptHandle,
// 				}
// 			}

// 			for {
// 				select {
// 				case bag := <-q.sack:
// 					delete(acks, bag.Digest)
// 					if len(acks) == 0 {
// 						return nil
// 					}
// 				case <-time.After(500 * time.Millisecond /*q.policy.TimeToFlight*/):
// 					// q.logger.Error("timeout message ack")
// 					return fmt.Errorf("timeout message ack")
// 				}
// 			}

// 		},
// 	)
// }

func attr(msg *events.SQSMessage, key string) string {
	val, exists := msg.MessageAttributes[key]
	if !exists {
		return ""
	}

	return *val.StringValue
}
