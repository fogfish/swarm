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

type Recver struct {
	id      string
	adapter *adapter.Adapter
	sock    chan *swarm.Bag
	sack    chan *swarm.Bag

	client sqsiface.SQSAPI
	queue  *string
}

//
//
func NewRecver(
	sys swarm.System,
	queue string,
	policy *swarm.Policy,
	session *session.Session,
) *Recver {
	logger := logger.With(logger.Note{"type": "sqs", "q": queue})
	adapt := adapter.New(sys, policy, logger)

	return &Recver{
		id:      queue,
		adapter: adapt,
		client:  sqs.New(session),
	}
}

// Mock ...
func (q *Recver) Mock(mock sqsiface.SQSAPI) {
	q.client = mock
}

//
//
func (q *Recver) ID() string { return q.id }

//
//
func (q *Recver) Start() error {
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
func (q *Recver) Close() error {
	close(q.sock)
	close(q.sack)

	return nil
}

//
//
func (q *Recver) Recv() chan *swarm.Bag {
	if q.sock == nil {
		q.sock = adapter.Recv(q.adapter, q.recv)
	}
	return q.sock
}

func (q *Recver) recv() (*swarm.Bag, error) {
	result, err := q.client.ReceiveMessage(
		&sqs.ReceiveMessageInput{
			MessageAttributeNames: []*string{aws.String("All")},
			QueueUrl:              q.queue,
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
		Category: attr(head, "Category"),
		System:   attr(head, "System"),
		Queue:    attr(head, "Queue"),
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

//
//
func (q *Recver) Conf() chan *swarm.Bag {
	if q.sack == nil {
		q.sack = adapter.Conf(q.adapter, q.conf)
	}
	return q.sack
}

func (q *Recver) conf(msg *swarm.Msg) error {
	_, err := q.client.DeleteMessage(
		&sqs.DeleteMessageInput{
			QueueUrl:      q.queue,
			ReceiptHandle: aws.String(string(msg.Receipt)),
		},
	)
	return err
}
