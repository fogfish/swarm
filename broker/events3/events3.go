//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package events3

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fogfish/guid/v2"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/internal/kernel"
)

// New creates broker for AWS S3 (serverless events)
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

type S3Event struct {
	Records []json.RawMessage `json:"Records"`
}

func (s spawner) Spawn(k *kernel.Kernel) error {
	s.f(
		func(events S3Event) error {
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
