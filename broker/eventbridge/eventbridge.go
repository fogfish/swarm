//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

// // Option type for backward compatibility
// type Option = opts.Option[Client]

// var (
// 	// Configures
// 	WithService = opts.ForType[Client, EventBridge]()

// 	// Explicitly configures the event bus to use
// 	WithEventBus = opts.ForName[Client, string]("bus")
// )

// var defs = []Option{
// 	WithConfig(
// 		swarm.WithSource(os.Getenv(EnvConfigEventAgent)),
// 		swarm.WithLogStdErr(),
// 		swarm.WithConfigFromEnv(),
// 	),
// }

// // Global "kernel" configuration for backward compatibility
// func WithConfig(opt ...opts.Option[swarm.Config]) Option {
// 	return opts.Type[Client](func(c *Client) error {
// 		config := swarm.NewConfig()
// 		if err := opts.Apply(&config, opt); err != nil {
// 			return err
// 		}

// 		// Mandatory overrides
// 		config.PollFrequency = 5 * time.Microsecond

// 		c.config = config
// 		return nil
// 	})
// }

// EventBridge declares the subset of interface from AWS SDK used by the lib.
type EventBridge interface {
	PutEvents(context.Context, *eventbridge.PutEventsInput, ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
}

// EventBridge client
type Client struct {
	service EventBridge
	bus     string
	config  swarm.Config
}

// // Create writer to AWS EventBridge
// // Deprecated: Use eventbridge.Channels().NewEnqueuer() instead
// func NewEnqueuer(opts ...Option) (*kernel.Enqueuer, error) {
// 	cli, err := newEventBridge(opts...)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return kernel.NewEnqueuer(cli, cli.config), nil
// }

// // Create writer to AWS EventBridge
// // Deprecated: Use eventbridge.Channels().NewEnqueuer() instead
// func MustEnqueuer(opts ...Option) *kernel.Enqueuer {
// 	cli, err := NewEnqueuer(opts...)
// 	if err != nil {
// 		xlog.Emergency("eventbridge client has failed", err)
// 	}

// 	return cli
// }

// // Create reader from AWS EventBridge
// // Deprecated: Use eventbridge.Channels().NewDequeuer() instead
// func NewDequeuer(opt ...Option) (*kernel.Dequeuer, error) {
// 	c := &Client{}
// 	if err := opts.Apply(c, defs); err != nil {
// 		return nil, err
// 	}
// 	if err := opts.Apply(c, opt); err != nil {
// 		return nil, err
// 	}

// 	bridge := &bridge{kernel.NewBridge(c.config.TimeToFlight)}

// 	return kernel.NewDequeuer(bridge, c.config), nil
// }

// // Deprecated: Use eventbridge.Channels().NewDequeuer() instead
// func MustDequeuer(opt ...Option) *kernel.Dequeuer {
// 	cli, err := NewDequeuer(opt...)
// 	if err != nil {
// 		xlog.Emergency("eventbridge client has failed", err)
// 	}

// 	return cli
// }

// // Create enqueue & dequeue routine to AWS EventBridge
// // Deprecated: Use eventbridge.Channels().NewClient() instead
// func New(opts ...Option) (*kernel.Kernel, error) {
// 	cli, err := newEventBridge(opts...)
// 	if err != nil {
// 		return nil, err
// 	}

// 	bridge := &bridge{kernel.NewBridge(cli.config.TimeToFlight)}

// 	return kernel.New(
// 		kernel.NewEnqueuer(cli, cli.config),
// 		kernel.NewDequeuer(bridge, cli.config),
// 	), nil
// }

// // checkRequired validates that all mandatory configuration parameters are provided
// func (c *Client) checkRequired() error {
// 	return opts.Required(c, WithEventBus(""))
// }

// func newEventBridge(opt ...Option) (*Client, error) {
// 	c := &Client{bus: os.Getenv(EnvConfigSourceEventBridge)}
// 	if err := opts.Apply(c, defs); err != nil {
// 		return nil, err
// 	}
// 	if err := opts.Apply(c, opt); err != nil {
// 		return nil, err
// 	}

// 	if c.service == nil {
// 		aws, err := config.LoadDefaultConfig(context.Background())
// 		if err != nil {
// 			return nil, swarm.ErrServiceIO.With(err)
// 		}
// 		c.service = eventbridge.NewFromConfig(aws)
// 	}

// 	if err := c.checkRequired(); err != nil {
// 		return nil, err
// 	}

// 	return c, nil
// }

// Enq enqueues message to broker
func (cli *Client) Enq(ctx context.Context, bag swarm.Bag) error {
	ctx, cancel := context.WithTimeout(ctx, cli.config.NetworkTimeout)
	defer cancel()

	ret, err := cli.service.PutEvents(ctx,
		&eventbridge.PutEventsInput{
			Entries: []types.PutEventsRequestEntry{
				{
					EventBusName: aws.String(cli.bus),
					Source:       aws.String(cli.config.Agent),
					DetailType:   aws.String(bag.Category),
					Detail:       aws.String(string(bag.Object)),
				},
			},
		},
	)
	if err != nil {
		return swarm.ErrEnqueue.With(err)
	}

	if ret.FailedEntryCount > 0 {
		return fmt.Errorf("%s: %s",
			aws.ToString(ret.Entries[0].ErrorCode),
			aws.ToString(ret.Entries[0].ErrorMessage),
		)
	}

	return nil
}

//------------------------------------------------------------------------------

type bridge struct{ *kernel.Bridge }

func (s bridge) Run(ctx context.Context) {
	lambda.Start(s.run)
}

func (s bridge) run(ctx context.Context, evt events.CloudWatchEvent) error {
	deadline, ok := ctx.Deadline()
	if ok {
		remaining := time.Until(deadline)
		fmt.Println("Remaining time:", remaining)
	}

	bag := make([]swarm.Bag, 1)
	bag[0] = swarm.Bag{
		Category: evt.DetailType,
		Digest:   evt.ID,
		Object:   evt.Detail,
	}

	return s.Bridge.Dispatch(ctx, bag)
}
