//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
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
	"github.com/fogfish/swarm/internal/kernel"
)

// EventBridge declares the subset of interface from AWS SDK used by the lib.
type EventBridge interface {
	PutEvents(context.Context, *eventbridge.PutEventsInput, ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
}

type Client struct {
	service EventBridge
	bus     string
	config  swarm.Config
}

func New(queue string, opts ...swarm.Option) (swarm.Broker, error) {
	cli, err := NewEventBridge(queue, opts...)
	if err != nil {
		return nil, err
	}

	config := swarm.NewConfig()
	for _, opt := range opts {
		opt(&config)
	}

	starter := lambda.Start

	type Mock interface{ Start(interface{}) }
	if config.Service != nil {
		service, ok := config.Service.(Mock)
		if ok {
			starter = service.Start
		}
	}

	sls := spawner{f: starter, c: config}

	return kernel.New(cli, sls), err
}

func NewEventBridge(bus string, opts ...swarm.Option) (*Client, error) {
	config := swarm.NewConfig()
	for _, opt := range opts {
		opt(&config)
	}

	api, err := newService(&config)
	if err != nil {
		return nil, err
	}

	return &Client{
		service: api,
		bus:     bus,
		config:  config,
	}, nil
}

func newService(conf *swarm.Config) (EventBridge, error) {
	if conf.Service != nil {
		service, ok := conf.Service.(EventBridge)
		if ok {
			return service, nil
		}
	}

	aws, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, swarm.ErrServiceIO.New(err)
	}

	return eventbridge.NewFromConfig(aws), nil
}

// Enq enqueues message to broker
func (cli *Client) Enq(bag swarm.Bag) error {
	ctx, cancel := context.WithTimeout(context.Background(), cli.config.NetworkTimeout)
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
		return fmt.Errorf("%v: %v", ret.Entries[0].ErrorCode, ret.Entries[0].ErrorMessage)
	}

	return nil
}

//------------------------------------------------------------------------------

type spawner struct {
	c swarm.Config
	f func(any)
}

func (s spawner) Spawn(k *kernel.Kernel) error {
	s.f(
		func(evt events.CloudWatchEvent) error {
			bag := make([]swarm.Bag, 1)
			bag[0] = swarm.Bag{
				Category: evt.DetailType,
				Object:   evt.Detail,
				Digest:   swarm.Digest{Brief: evt.ID},
			}

			return k.Dispatch(bag, s.c.TimeToFlight)
		},
	)

	return nil
}

func (s spawner) Ack(digest string) error   { return nil }
func (s spawner) Ask() ([]swarm.Bag, error) { return nil, nil }
