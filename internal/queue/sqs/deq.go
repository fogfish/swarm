//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

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

type Dequeue struct {
	id      string
	adapter *adapter.Adapter
	sock    chan *swarm.Bag
	sack    chan *swarm.Bag

	client sqsiface.SQSAPI
	queue  *string
	logger logger.Logger
}

//
//
func NewDequeue(
	sys swarm.System,
	queue string,
	policy *swarm.Policy,
	session *session.Session,
) *Dequeue {
	logger := logger.With(
		logger.Note{
			"type":  "sqs",
			"queue": sys.ID() + "://" + queue,
		},
	)

	adapt := adapter.New(policy, logger)
	return &Dequeue{
		id:      queue,
		adapter: adapt,
		client:  sqs.New(session),
		logger:  logger,
	}
}

// Mock ...
func (q *Dequeue) Mock(mock sqsiface.SQSAPI) {
	q.client = mock
}

//
//
func (q *Dequeue) Listen() error {
	q.logger.Info("dequeue listening")

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
func (q *Dequeue) Close() error {
	close(q.sock)
	close(q.sack)

	q.logger.Info("dequeue closed")
	return nil
}

//
//
func (q *Dequeue) Deq() chan *swarm.Bag {
	if q.sock == nil {
		q.sock = adapter.Deq(q.adapter, q.deq)
	}
	return q.sock
}

func (q *Dequeue) deq() (*swarm.Bag, error) {
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
		System:   attr(head, "System"),
		Queue:    attr(head, "Queue"),
		Category: attr(head, "Category"),
		Object:   []byte(*head.Body),
		Digest:   *head.ReceiptHandle,
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
func (q *Dequeue) Ack() chan *swarm.Bag {
	if q.sack == nil {
		q.sack = adapter.Ack(q.adapter, q.ack)
	}
	return q.sack
}

func (q *Dequeue) ack(msg *swarm.Bag) error {
	_, err := q.client.DeleteMessage(
		&sqs.DeleteMessageInput{
			QueueUrl:      q.queue,
			ReceiptHandle: aws.String(string(msg.Digest)),
		},
	)
	return err
}
