package eventbridge

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/router"
)

type EventBridge interface {
	PutEvents(context.Context, *eventbridge.PutEventsInput, ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
}

type broker struct {
	client   *client
	channels *swarm.Channels
	context  context.Context
	cancel   context.CancelFunc
	router   *router.Router
}

func New(service EventBridge, bus string) (swarm.Broker, error) {
	cli, err := newClient(service, bus)
	if err != nil {
		return nil, err
	}

	ctx, can := context.WithCancel(context.Background())

	return &broker{
		client:   cli,
		channels: swarm.NewChannels(),
		context:  ctx,
		cancel:   can,
		router:   router.New(),
	}, nil
}

func (b *broker) Close() {
	b.channels.Sync()
	b.channels.Close()
	b.cancel()
}

func (b *broker) Await() {
	lambda.Start(
		func(evt events.CloudWatchEvent) error {
			bag := swarm.Bag{
				Queue:    evt.Source,
				Category: evt.DetailType,
				Object:   evt.Detail,
				Digest:   evt.ID,
			}

			if err := b.router.Dispatch(bag); err != nil {
				return err
			}

			return b.router.Await(1 * time.Second /*q.policy.TimeToFlight*/)
		},
	)
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
