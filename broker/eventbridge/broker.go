package eventbridge

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/router"
)

// EventBridge declares the subset of interface from AWS SDK,
// which is used by the library
type EventBridge interface {
	PutEvents(context.Context, *eventbridge.PutEventsInput, ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
}

type broker struct {
	config   *swarm.Config
	client   *client
	channels *swarm.Channels
	context  context.Context
	cancel   context.CancelFunc
	router   *router.Router
}

// New create broker for AWS EventBridge service
func New(bus string, opts ...swarm.Option) (swarm.Broker, error) {
	conf := swarm.NewConfig()
	for _, opt := range opts {
		opt(conf)
	}

	cli, err := newClient(bus, conf)
	if err != nil {
		return nil, err
	}

	ctx, can := context.WithCancel(context.Background())

	return &broker{
		config:   conf,
		client:   cli,
		channels: swarm.NewChannels(),
		context:  ctx,
		cancel:   can,
		router:   router.New(nil),
	}, nil
}

func (b *broker) Config() *swarm.Config {
	return b.config
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
				Category: evt.DetailType,
				Object:   evt.Detail,
				Digest:   evt.ID,
			}

			if err := b.router.Dispatch(bag); err != nil {
				return err
			}

			return b.router.Await(b.config.TimeToFlight)
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
