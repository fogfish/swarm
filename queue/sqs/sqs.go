package sqs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/fogfish/logger"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/queue"
)

// TODO: lazy sender & receiver

type awssqs struct {
	*queue.Queue

	name string
	url  *string
}

//
func New(sys swarm.System, name string) (swarm.Queue, error) {
	q := &awssqs{name: name}

	spec, err := q.newService().GetQueueUrl(
		&sqs.GetQueueUrlInput{
			QueueName: aws.String(name),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to discover %s AWS SQS: %w", name, err)
	}
	q.url = spec.QueueUrl

	// send := q.newSend(sys)
	// recv := q.createRecv(sys)
	q.Queue = queue.New(sys, q.newRecv, q.newSend)

	return q, nil
}

//
func (q *awssqs) newService() *sqs.SQS {
	return sqs.New(
		session.Must(
			session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}),
		),
	)
}

//
func (q *awssqs) newSend() chan<- *swarm.Message {
	sock := make(chan *swarm.Message)
	service := q.newService()

	q.System.Spawn(func(ctx context.Context) {
		logger.Notice("init aws sqs send %s", q.name)

		for {
			select {
			case <-ctx.Done():
				logger.Notice("free aws sqs send %p", q)
				return
			case msg := <-sock:
				fmt.Printf("==> %+v\n", msg)
				r, err := service.SendMessage(
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
				fmt.Println("==>")
				fmt.Println(err)
				fmt.Println(r)
			}
		}
	})

	return sock
}

//
func (q *awssqs) newRecv() <-chan *swarm.Message {
	sock := make(chan *swarm.Message)
	return sock
}
