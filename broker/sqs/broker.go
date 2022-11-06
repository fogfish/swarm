package sqs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/router"
)

// SQS
type SQS interface {
	GetQueueUrl(context.Context, *sqs.GetQueueUrlInput, ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
	SendMessage(context.Context, *sqs.SendMessageInput, ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	ReceiveMessage(context.Context, *sqs.ReceiveMessageInput, ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(context.Context, *sqs.DeleteMessageInput, ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type broker struct {
	client   *client
	channels *swarm.Channels
	context  context.Context
	cancel   context.CancelFunc
	router   *router.Router
}

func New(service SQS, queue string) (swarm.Broker, error) {
	cli, err := newClient(service, queue)
	if err != nil {
		return nil, err
	}

	ctx, can := context.WithCancel(context.Background())

	return &broker{
		client:   cli,
		channels: swarm.NewChannels(),
		context:  ctx,
		cancel:   can,
		router:   router.New(cli.Ack),
	}, nil
}

func (b *broker) Close() {
	b.channels.Sync()
	b.channels.Close()
	b.cancel()
}

func (b *broker) Await() {
	for {
		select {
		case <-b.context.Done():
			return
		default:
			bag, err := b.client.Deq("")
			if err != nil {
				fmt.Println(err)
				return
			}

			if bag.Object != nil {
				b.router.Dispatch(bag)
				if err := b.router.Await(1 * time.Second); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}

func (b *broker) Enqueue(category string, channel swarm.Channel) (swarm.Enqueue, error) {
	b.channels.Attach(category, channel)

	return b.client, nil
}

func (b *broker) Dequeue(category string, channel swarm.Channel) (swarm.Dequeue, error) {
	b.channels.Attach(category, channel)
	b.router.Register(category)

	return b.router, nil
}
