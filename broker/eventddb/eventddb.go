//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventddb

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fogfish/guid/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
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

	return kernel.New(nil, sls, config), nil
}

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
					Ctx:    swarm.NewContext(context.Background(), Category, guid.G(guid.Clock).String()),
					Object: obj,
				}
			}

			return k.Dispatch(bag, s.c.TimeToFlight)
		},
	)

	return nil
}

func (s spawner) Ack(digest string) error   { return nil }
func (s spawner) Ask() ([]swarm.Bag, error) { return nil, nil }
