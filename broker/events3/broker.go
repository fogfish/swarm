//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events3

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fogfish/guid"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/router"
)

// New creates broker for AWS EventBridge
func New(queue string, opts ...swarm.Option) (swarm.Broker, error) {
	conf := swarm.NewConfig()
	for _, opt := range opts {
		opt(&conf)
	}

	ctx, can := context.WithCancel(context.Background())

	return &broker{
		config:   conf,
		channels: swarm.NewChannels(),
		context:  ctx,
		cancel:   can,
		router:   router.New(&conf, nil),
	}, nil
}

type broker struct {
	config   swarm.Config
	channels *swarm.Channels
	context  context.Context
	cancel   context.CancelFunc
	router   *router.Router
}

func (b *broker) Config() swarm.Config {
	return b.config
}

func (b *broker) Close() {
	b.channels.Sync()
	b.channels.Close()
	b.cancel()
}

func (b *broker) DSync() {
	b.channels.Sync()
}

func (b *broker) Enqueue(category string, channel swarm.Channel) swarm.Enqueue {
	panic("not implemented")
}

func (b *broker) Dequeue(category string, channel swarm.Channel) swarm.Dequeue {
	b.channels.Attach(category, channel)
	b.router.Register(category)

	return b.router
}

func (b *broker) Await() {
	starter := lambda.Start

	type Mock interface{ Start(interface{}) }
	if b.config.Service != nil {
		service, ok := b.config.Service.(Mock)
		if ok {
			starter = service.Start
		}
	}

	starter(
		func(events events.S3Event) error {
			for _, evt := range events.Records {
				bag := swarm.Bag{
					Category: Category,
					Event:    &Event{Object: &evt},
					Digest:   guid.L.K(guid.Clock).String(),
				}
				if err := b.router.Dispatch(bag); err != nil {
					return err
				}
			}

			return b.router.Await(b.config.TimeToFlight)
		},
	)
}
