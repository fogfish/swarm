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

//
//
type Sender struct {
	id      string
	adapter *adapter.Adapter
	sock    chan *swarm.Bag

	client sqsiface.SQSAPI
	queue  *string
}

//
//
func NewSender(
	sys swarm.System,
	queue string,
	policy *swarm.Policy,
	session *session.Session,
) *Sender {
	logger := logger.With(logger.Note{"type": "sqs", "q": queue})
	adapt := adapter.New(sys, policy, logger)

	return &Sender{
		id:      queue,
		adapter: adapt,
		client:  sqs.New(session),
	}
}

// Mock ...
func (q *Sender) Mock(mock sqsiface.SQSAPI) {
	q.client = mock
}

//
//
func (q *Sender) ID() string { return q.id }

//
//
func (q *Sender) Start() error {
	if q.queue == nil {
		spec, err := q.client.GetQueueUrl(
			&sqs.GetQueueUrlInput{
				QueueName: aws.String(q.id),
			},
		)
		if err != nil {
			return fmt.Errorf("AWS Queue Not Found %s: %w", q.id, err)
		}

		q.queue = spec.QueueUrl
	}

	return nil
}

//
//
func (q *Sender) Close() error {
	close(q.sock)

	return nil
}

//
//
func (q *Sender) Send() chan *swarm.Bag {
	if q.sock == nil {
		q.sock = adapter.Send(q.adapter, q.send)
	}
	return q.sock
}

func (q *Sender) send(msg *swarm.Bag) error {
	_, err := q.client.SendMessage(
		&sqs.SendMessageInput{
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"System":   {StringValue: aws.String(msg.System), DataType: aws.String("String")},
				"Queue":    {StringValue: aws.String(msg.Queue), DataType: aws.String("String")},
				"Category": {StringValue: aws.String(msg.Category), DataType: aws.String("String")},
			},
			MessageBody: aws.String(string(msg.Object.Bytes())),
			QueueUrl:    q.queue,
		},
	)
	return err
}
