package sqs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
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

	SQS sqsiface.SQSAPI
	url *string
}

/*

Config ...
*/
type Config func(*Queue)

/*

New ...
*/
func New(sys swarm.System, id string, opts ...Config) (swarm.Queue, error) {
	q := &Queue{}
	if err := q.newSession(); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(q)
	}

	if err := q.lookupQueue(id); err != nil {
		return nil, err
	}

	q.Queue = queue.New(sys, id, q.newRecvAcks, q.newSend)
	return q, nil
}

//
func (q *Queue) newSession() error {
	api, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return fmt.Errorf("Failed to create aws client %w", err)
	}

	q.SQS = sqs.New(api)
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

//
func (q *Queue) newSend() chan<- *queue.Bag {
	sock := make(chan *queue.Bag)

	q.System.Go(func(ctx context.Context) {
		logger.Notice("init aws sqs send %s", q.ID)
		defer close(sock)

		for {
			select {
			//
			case <-ctx.Done():
				logger.Notice("free aws sqs send %s", q.ID)
				return

			//
			case msg := <-sock:
				// TODO: queue config
				err := backoff.
					Exp(10*time.Millisecond, 10, 0.5).
					Deadline(30 * time.Second).
					Retry(func() error { return q.send(msg) })

				if err != nil {
					msg.StdErr <- msg.Object
					logger.Debug("failed to send sqs message %v", err)
				}
			}
		}
	})

	return sock
}

func (q *Queue) send(msg *queue.Bag) error {
	_, err := q.SQS.SendMessage(
		&sqs.SendMessageInput{
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"Target":   mkMsgAttr(msg.Target),
				"Source":   mkMsgAttr(msg.Source),
				"Category": mkMsgAttr(string(msg.Category)),
			},
			MessageBody: aws.String(string(msg.Object.Bytes())),
			QueueUrl:    q.url,
		},
	)
	return err
}

//
func (q *Queue) newRecvAcks() (<-chan *queue.Bag, chan<- *queue.Bag) {
	sock := q.newRecv()
	acks := q.newAcks()

	return sock, acks
}

//
func (q *Queue) newRecv() <-chan *queue.Bag {
	sock := make(chan *queue.Bag)
	delay := 1 * time.Microsecond

	q.System.Go(func(ctx context.Context) {
		logger.Notice("init aws sqs recv %s", q.ID)
		defer close(sock)

		for {
			select {
			//
			case <-ctx.Done():
				logger.Notice("free aws sqs recv %s", q.ID)
				return

			//
			case <-time.After(delay):
				// TODO: retry
				seq, err := q.recv()
				if err != nil {
					logger.Error("Unable to receive message %v", err)
					break
				}

				if len(seq) > 0 {
					// TODO: check memory issue with range
					for _, msg := range seq {
						sock <- mkMsgBag(msg)
					}
				}
			}
		}
	})

	return sock
}

//
func (q *Queue) newAcks() chan<- *queue.Bag {
	acks := make(chan *queue.Bag)

	q.System.Go(func(ctx context.Context) {
		logger.Notice("init aws sqs acks %s", q.ID)
		defer close(acks)

		for {
			select {
			//
			case <-ctx.Done():
				logger.Notice("free aws sqs acks %s", q.ID)
				return

			//
			case bag := <-acks:
				switch v := bag.Object.(type) {
				case *queue.Msg:
					// TODO: retry
					_, err := q.SQS.DeleteMessage(
						&sqs.DeleteMessageInput{
							QueueUrl:      q.url,
							ReceiptHandle: aws.String(string(v.Receipt)),
						},
					)
					if err != nil {
						logger.Error("Unable to delete message %v", err)
					}
				default:
					logger.Notice("Unsupported ack type %s", q.ID)
				}
			}
		}
	})

	return acks
}

func (q *Queue) recv() ([]*sqs.Message, error) {
	result, err := q.SQS.ReceiveMessage(
		&sqs.ReceiveMessageInput{
			MessageAttributeNames: []*string{aws.String("All")},
			QueueUrl:              q.url,
			MaxNumberOfMessages:   aws.Int64(1),
		},
	)
	if err != nil {
		return nil, err
	}

	return result.Messages, nil
}

func mkMsgAttr(value string) *sqs.MessageAttributeValue {
	return &sqs.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(value),
	}
}

func msgAttr(msg *sqs.Message, key string) string {
	val, exists := msg.MessageAttributes[key]
	if !exists {
		return ""
	}

	return *val.StringValue
}

func mkMsgBag(msg *sqs.Message) *queue.Bag {
	return &queue.Bag{
		Target:   msgAttr(msg, "Target"),
		Source:   msgAttr(msg, "Source"),
		Category: swarm.Category(msgAttr(msg, "Category")),
		Object: &queue.Msg{
			Payload: []byte(*msg.Body),
			Receipt: *msg.ReceiptHandle,
		},
	}
}
