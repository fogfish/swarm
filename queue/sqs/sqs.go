package sqs

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
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

	SQS sqsiface.SQSAPI
	url *string
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

PolicyPoll configures frequency of polling loop

	sqs.New(sys, "q", sqs.PolicyPoll(5*time.Seconds) )

*/
func PolicyPoll(frequency time.Duration) Config {
	return func(q *Queue) {
		q.adapter.Policy.PollFrequency = frequency
	}
}

/*

New ...
*/
func New(sys swarm.System, id string, opts ...Config) (swarm.Queue, error) {
	q := &Queue{adapter: queue.Adapt(sys, "sqs", id)}
	if err := q.newSession(); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(q)
	}

	if err := q.lookupQueue(id); err != nil {
		return nil, err
	}

	q.Queue = queue.New(sys, id, q.spawnRecvIO, q.spawnSendIO)
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

	q.SQS = sqs.New(awscli)
	return nil
}

//
func (q *Queue) lookupQueue(id string) error {
	spec, err := q.SQS.GetQueueUrl(
		&sqs.GetQueueUrlInput{
			QueueName: aws.String(id),
		},
	)

	if err != nil {
		return fmt.Errorf("Failed to discover %s aws sqs: %w", id, err)
	}

	q.url = spec.QueueUrl
	return nil
}

/*

spawnSendIO create go routine for emiting messages
*/
func (q *Queue) spawnSendIO() chan<- *queue.Bag {
	return q.adapter.SendIO(q.send)
}

func (q *Queue) send(msg *queue.Bag) error {
	_, err := q.SQS.SendMessage(
		&sqs.SendMessageInput{
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"Target":   {StringValue: aws.String(msg.Target), DataType: aws.String("String")},
				"Source":   {StringValue: aws.String(msg.Source), DataType: aws.String("String")},
				"Category": {StringValue: aws.String(string(msg.Category)), DataType: aws.String("String")},
			},
			MessageBody: aws.String(string(msg.Object.Bytes())),
			QueueUrl:    q.url,
		},
	)
	return err
}

//
func (q *Queue) spawnRecvIO() (<-chan *queue.Bag, chan<- *queue.Bag) {
	sock := q.adapter.RecvIO(q.recv)
	conf := q.adapter.ConfIO(q.conf)
	return sock, conf
}

func (q *Queue) recv() (*queue.Bag, error) {
	result, err := q.SQS.ReceiveMessage(
		&sqs.ReceiveMessageInput{
			MessageAttributeNames: []*string{aws.String("All")},
			QueueUrl:              q.url,
			MaxNumberOfMessages:   aws.Int64(1),
			WaitTimeSeconds:       aws.Int64(10),
		},
	)
	if err != nil {
		return nil, err
	}

	if len(result.Messages) == 0 {
		return nil, nil
	}

	head := result.Messages[0]

	return &queue.Bag{
		Target:   attr(head, "Target"),
		Source:   attr(head, "Source"),
		Category: swarm.Category(attr(head, "Category")),
		Object: &queue.Msg{
			Payload: []byte(*head.Body),
			Receipt: *head.ReceiptHandle,
		},
	}, nil
}

func attr(msg *sqs.Message, key string) string {
	val, exists := msg.MessageAttributes[key]
	if !exists {
		return ""
	}

	return *val.StringValue
}

func (q *Queue) conf(msg *queue.Msg) error {
	_, err := q.SQS.DeleteMessage(
		&sqs.DeleteMessageInput{
			QueueUrl:      q.url,
			ReceiptHandle: aws.String(string(msg.Receipt)),
		},
	)
	return err
}
