package sqs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/backoff"
	"github.com/fogfish/swarm/queue"
)

type awssqs struct {
	*queue.Queue

	sqs  *sqs.SQS
	name string
	url  *string
}

//
func New(sys swarm.System, name string) (swarm.Queue, error) {
	api, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create aws client %w", err)
	}

	q := &awssqs{sqs: sqs.New(api), name: name}

	if err := q.lookupQueue(); err != nil {
		return nil, err
	}

	q.Queue = queue.New(sys, q.newRecv, q.newSend)
	return q, nil
}

//
func (q *awssqs) lookupQueue() error {
	spec, err := q.sqs.GetQueueUrl(
		&sqs.GetQueueUrlInput{
			QueueName: aws.String(q.name),
		},
	)

	if err != nil {
		return fmt.Errorf("Failed to discover %s aws sqs: %w", q.name, err)
	}

	q.url = spec.QueueUrl
	return nil
}

//
func (q *awssqs) newSend() chan<- *swarm.Message {
	sock := make(chan *swarm.Message)

	q.System.Spawn(func(ctx context.Context) {
		logger.Notice("init aws sqs send %s", q.name)
		defer close(sock)

		for {
			select {
			case <-ctx.Done():
				logger.Notice("free aws sqs send %s", q.name)
				return
			case msg := <-sock:
				err := backoff.
					Exp(10*time.Millisecond, 10, 0.5).
					Deadline(30 * time.Second).
					Retry(func() error { return q.send(msg) })

				if err != nil {
					// TODO: send to error channel
					fmt.Printf("==> %+v\n", msg)
					fmt.Println(err)
				}
			}
		}
	})

	return sock
}

func (q *awssqs) send(msg *swarm.Message) error {
	_, err := q.sqs.SendMessage(
		&sqs.SendMessageInput{
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"Category": {
					DataType:    aws.String("String"),
					StringValue: (*string)(&msg.Category),
				},
			},
			MessageBody: aws.String(string(msg.Object)),
			QueueUrl:    q.url,
		},
	)
	return err
}

//
func (q *awssqs) newRecv() <-chan *swarm.Message {
	sock := make(chan *swarm.Message)
	delay := 1 * time.Microsecond

	q.System.Spawn(func(ctx context.Context) {
		logger.Notice("init aws sqs recv %s", q.name)
		defer close(sock)

		for {
			select {
			case <-ctx.Done():
				logger.Notice("free aws sqs recv %s", q.name)
				return
			case <-time.After(delay):
				// service.Rec
				r, err := q.sqs.ReceiveMessage(
					&sqs.ReceiveMessageInput{
						MessageAttributeNames: []*string{
							aws.String("All"),
						},
						QueueUrl:            q.url,
						MaxNumberOfMessages: aws.Int64(1),
					},
				)
				// service.DeleteMessage(
				// 	&sqs.DeleteMessageInput{
				// 		QueueUrl: q.url,
				// 		ReceiptHandle: ,
				// 	}
				// )
				delay = 1 * time.Second
				logger.Debug("sqs i/o %v %+v", err, r)
			}
		}

	})

	return sock
}
