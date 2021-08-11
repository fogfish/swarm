package eventbridge

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/queue"
)

/*

Queue ...
*/
type Queue struct {
	*queue.Queue

	Bus   eventbridgeiface.EventBridgeAPI
	Start func(interface{})
}

/*

Config ...
*/
type Config func(*Queue)

/*

New ...
*/
func New(sys swarm.System, id string, opts ...Config) (swarm.Queue, error) {
	q := &Queue{Start: lambda.Start}
	if err := q.newSession(); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(q)
	}

	q.Queue = queue.New(sys, id, q.newRecv, q.newSend)
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

//
func (q *Queue) newSend() chan<- *queue.Bag {
	return q.Sender("eventbridge", q.send)
}

func (q *Queue) send(msg *queue.Bag) error {
	_, err := q.Bus.PutEvents(&eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{
			{
				EventBusName: aws.String(q.ID),
				Source:       aws.String(msg.Source),
				DetailType:   aws.String(string(msg.Category)),
				Detail:       aws.String(string(msg.Object.Bytes())),
			},
		},
	})

	return err
}

//
func (q *Queue) newRecv() (<-chan *queue.Bag, chan<- *queue.Bag) {
	sock := make(chan *queue.Bag)
	acks := make(chan *queue.Bag)

	go func() {
		logger.Notice("init aws eventbridge recv %s", q.ID)

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

				<-acks
				// TODO: ack timeout
				return nil
			},
		)
		q.System.Stop()
	}()

	return sock, acks
}
