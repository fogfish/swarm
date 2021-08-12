package eventbridge

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/backoff"
	"github.com/fogfish/swarm/queue"
)

/*

Queue ...
*/
type Queue struct {
	*queue.Queue
	adapter *queue.Adapter

	Bus   eventbridgeiface.EventBridgeAPI
	Start func(interface{})
}

/*

Config ...
*/
type Config func(*Queue)

/*

PolicyIO configures retry of queue I/O

	sqs.New(sys, "q", sqs.PolicyIO(backoff.Exp(...)) )

*/
func PolicyIO(policy backoff.Seq) Config {
	return func(q *Queue) {
		q.adapter.Policy.IO = policy
	}
}

/*

PolicyTimeToFlight configures ack timeout
*/
func PolicyTimeToFlight(t time.Duration) Config {
	return func(q *Queue) {
		q.adapter.Policy.TimeToFlight = t
	}
}

/*

New ...
*/
func New(sys swarm.System, id string, opts ...Config) (swarm.Queue, error) {
	q := &Queue{
		adapter: queue.Adapt(sys, "eventbridge", id),
		Start:   lambda.Start,
	}
	if err := q.newSession(); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(q)
	}

	q.Queue = queue.New(sys, id, q.newRecv, q.spawnSendIO)
	return q, nil
}

//
func (q *Queue) newSession() error {
	awscli, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return fmt.Errorf("Failed to create aws client %w", err)
	}

	q.Bus = eventbridge.New(awscli)
	return nil
}

/*

spawnSendIO create go routine for emiting messages
*/
func (q *Queue) spawnSendIO() chan<- *queue.Bag {
	return q.adapter.SendIO(q.send)
}

func (q *Queue) send(msg *queue.Bag) error {
	ret, err := q.Bus.PutEvents(&eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{
			{
				EventBusName: aws.String(q.ID),
				Source:       aws.String(msg.Source),
				DetailType:   aws.String(string(msg.Category)),
				Detail:       aws.String(string(msg.Object.Bytes())),
			},
		},
	})

	if err != nil {
		return err
	}

	if *ret.FailedEntryCount > 0 {
		return fmt.Errorf("%v: %v", ret.Entries[0].ErrorCode, ret.Entries[0].ErrorMessage)
	}

	return nil
}

//
func (q *Queue) newRecv() (<-chan *queue.Bag, chan<- *queue.Bag) {
	sock := make(chan *queue.Bag)
	conf := make(chan *queue.Bag)

	go func() {
		logger.Notice("init aws eventbridge recv %s", q.ID)

		// Note: this indirect synonym for lambda.Start
		q.Start(
			func(evt events.CloudWatchEvent) error {
				sock <- &queue.Bag{
					Source:   evt.Source,
					Category: swarm.Category(evt.DetailType),
					Object: &queue.Msg{
						Payload: evt.Detail,
						Receipt: evt.ID,
					},
				}

				select {
				case <-conf:
					return nil
				case <-time.After(q.adapter.Policy.TimeToFlight):
					return errors.New("message ack timeout: " + evt.ID)
				}
			},
		)
		q.System.Stop()
	}()

	return sock, conf
}
