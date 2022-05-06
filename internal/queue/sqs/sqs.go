package sqs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/queue/adapter"
)

/*

Queue ...
*/
type Queue struct {
	id      string
	adapter *adapter.Adapter

	SQS sqsiface.SQSAPI
	url *string
}

/*

New ...
*/
func New(
	sys swarm.System,
	queue string,
	policy *swarm.Policy,
	defSession ...*session.Session,
) (*Queue, error) {
	logger := logger.With(logger.Note{
		"type": "sqs",
		"q":    queue,
	})
	q := &Queue{
		id:      queue,
		adapter: adapter.New(sys, policy, logger),
	}

	if err := q.newSession(defSession); err != nil {
		return nil, err
	}

	return q, nil
}

// Mock ...
func (q *Queue) Mock(mock sqsiface.SQSAPI) {
	q.SQS = mock
}

//
func (q *Queue) newSession(defSession []*session.Session) error {
	if len(defSession) != 0 {
		q.SQS = sqs.New(defSession[0])
		return nil
	}

	awscli, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return err
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

func (q *Queue) ID() string {
	return q.id
}

/*

spawnSendIO create go routine for emiting messages
*/
func (q *Queue) Send() (chan *swarm.Bag, error) {
	if q.url == nil {
		if err := q.lookupQueue(q.id); err != nil {
			return nil, err
		}
	}

	return adapter.Send(q.adapter, q.send), nil
}

func (q *Queue) send(msg *swarm.Bag) error {
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
//
func (q *Queue) Recv() (chan *swarm.Bag, error) {
	if q.url == nil {
		if err := q.lookupQueue(q.id); err != nil {
			return nil, err
		}
	}

	return adapter.Recv(q.adapter, q.recv), nil
}

func (q *Queue) recv() (*swarm.Bag, error) {
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

	return &swarm.Bag{
		Target:   attr(head, "Target"),
		Source:   attr(head, "Source"),
		Category: swarm.Category(attr(head, "Category")),
		Object: &swarm.Msg{
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

func (q *Queue) Conf() (chan *swarm.Bag, error) {
	if q.url == nil {
		if err := q.lookupQueue(q.id); err != nil {
			return nil, err
		}
	}

	return adapter.Conf(q.adapter, q.conf), nil
}

func (q *Queue) conf(msg *swarm.Msg) error {
	_, err := q.SQS.DeleteMessage(
		&sqs.DeleteMessageInput{
			QueueUrl:      q.url,
			ReceiptHandle: aws.String(string(msg.Receipt)),
		},
	)
	return err
}
