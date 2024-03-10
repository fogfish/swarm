//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventddb

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fogfish/guid/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/kernel"
)

// New creates broker for AWS EventBridge
func New(queue string, opts ...swarm.Option) (swarm.Broker, error) {
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

	return kernel.New(nil, sls), nil

	// conf := swarm.NewConfig()
	// for _, opt := range opts {
	// 	opt(&conf)
	// }

	// ctx, can := context.WithCancel(context.Background())

	// slog.Info("Broker is created", "type", "ddbstream")
	// return &broker{
	// 	config:   conf,
	// 	channels: swarm.NewChannels(),
	// 	context:  ctx,
	// 	cancel:   can,
	// 	router:   router.New(&conf, nil),
	// }, nil
}

// type broker struct {
// 	config   swarm.Config
// 	channels *swarm.Channels
// 	context  context.Context
// 	cancel   context.CancelFunc
// 	router   *router.Router
// }

// func (b *broker) Config() swarm.Config {
// 	return b.config
// }

// func (b *broker) Close() {
// 	b.channels.Sync()
// 	b.channels.Close()
// 	b.cancel()
// }

// func (b *broker) DSync() {
// 	b.channels.Sync()
// }

// func (b *broker) Enqueue(category string, channel swarm.Channel) swarm.Enqueue {
// 	panic("not implemented")
// }

// func (b *broker) Dequeue(category string, channel swarm.Channel) swarm.Dequeue {
// 	b.channels.Attach(category, channel)
// 	b.router.Register(category)

// 	return b.router
// }

// func (b *broker) Await() {
// 	starter := lambda.Start

// 	type Mock interface{ Start(interface{}) }
// 	if b.config.Service != nil {
// 		service, ok := b.config.Service.(Mock)
// 		if ok {
// 			starter = service.Start
// 		}
// 	}

// 	starter(
// 		func(events events.DynamoDBEvent) error {
// 			for _, evt := range events.Records {
// 				bag := swarm.Bag{
// 					Category: Category,
// 					Event: &Event{
// 						ID:     evt.EventID,
// 						Type:   curie.IRI(evt.EventName),
// 						Agent:  curie.IRI(evt.EventSourceArn),
// 						Object: &evt,
// 					},
// 					Digest: swarm.Digest{Brief: evt.EventID},
// 				}
// 				if err := b.router.Dispatch(bag); err != nil {
// 					return err
// 				}
// 			}

// 			return b.router.Await(b.config.TimeToFlight)
// 		},
// 	)
// }

//------------------------------------------------------------------------------

type spawner struct {
	c swarm.Config
	f func(any)
}

type DynamoDBEvent struct {
	Records []json.RawMessage `json:"Records"`
}

func (s spawner) Spawn(k *kernel.Kernel) error {
	s.f(
		func(events DynamoDBEvent) error {
			bag := make([]swarm.Bag, len(events.Records))
			for i, obj := range events.Records {
				bag[i] = swarm.Bag{
					Category: Category,
					Object:   obj,
					Digest:   swarm.Digest{Brief: guid.G(guid.Clock).String()},
				}
			}

			return k.Dispatch(bag, s.c.TimeToFlight)
		},
	)

	return nil
}

func (s spawner) Ack(digest string) error   { return nil }
func (s spawner) Ask() ([]swarm.Bag, error) { return nil, nil }
