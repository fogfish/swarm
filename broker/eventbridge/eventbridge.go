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

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

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

// Create writer to AWS EventBridge
func NewEnqueuer(bus string, opts ...Option) (*kernel.Enqueuer, error) {
	cli, err := newEventBridge(bus, opts...)
	if err != nil {
		return nil, err
	}

	return kernel.NewEnqueuer(cli, cli.config), nil
}

// Create reader from AWS EventBridge
func NewDequeuer(bus string, opts ...Option) (*kernel.Dequeuer, error) {
	c := &Client{bus: bus}

	for _, opt := range defs {
		opt(c)
	}
	for _, opt := range opts {
		opt(c)
	}

	bridge := &bridge{kernel.NewBridge(c.config.TimeToFlight)}

	return kernel.NewDequeuer(bridge, c.config), nil
}

// Create enqueue & dequeue routine to AWS EventBridge
func New(bus string, opts ...Option) (*kernel.Kernel, error) {
	cli, err := newEventBridge(bus, opts...)
	if err != nil {
		return nil, err
	}

	bridge := &bridge{kernel.NewBridge(cli.config.TimeToFlight)}

	return kernel.New(
		kernel.NewEnqueuer(cli, cli.config),
		kernel.NewDequeuer(bridge, cli.config),
	), nil
}

func newEventBridge(queue string, opts ...Option) (*Client, error) {
	c := &Client{bus: queue}

	for _, opt := range defs {
		opt(c)
	}
	for _, opt := range opts {
		opt(c)
	}

	if c.service == nil {
		aws, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, swarm.ErrServiceIO.New(err)
		}
		c.service = eventbridge.NewFromConfig(aws)
	}

	return c, nil
}

// Enq enqueues message to broker
func (cli *Client) Enq(ctx context.Context, bag swarm.Bag) error {
	ctx, cancel := context.WithTimeout(ctx, cli.config.NetworkTimeout)
	defer cancel()

	ret, err := cli.service.PutEvents(ctx,
		&eventbridge.PutEventsInput{
			Entries: []types.PutEventsRequestEntry{
				{
					EventBusName: aws.String(cli.bus),
					Source:       aws.String(cli.config.Source),
					DetailType:   aws.String(bag.Category),
					Detail:       aws.String(string(bag.Object)),
				},
			},
		},
	)
	if err != nil {
		return swarm.ErrEnqueue.New(err)
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

func (s bridge) Run() { lambda.Start(s.run) }

func (s bridge) run(evt events.CloudWatchEvent) error {
	bag := make([]swarm.Bag, 1)
	bag[0] = swarm.Bag{
		Category: evt.DetailType,
		Digest:   evt.ID,
		Object:   evt.Detail,
	}

	return s.Bridge.Dispatch(bag)
}
