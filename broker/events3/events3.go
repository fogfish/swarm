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
	"github.com/fogfish/swarm/kernel"
)

type Client struct {
	bucket string
	config swarm.Config
}

// Create reader from AWS S3 Events
func NewReader(opts ...Option) (*kernel.Dequeuer, error) {
	c := &Client{}

	for _, opt := range defs {
		opt(c)
	}
	for _, opt := range opts {
		opt(c)
	}

	bridge := &bridge{kernel.NewBridge(c.config.TimeToFlight)}

	return kernel.NewDequeuer(bridge, c.config), nil
}

// New creates broker for AWS S3 (serverless events)
// func New(queue string, opts ...swarm.Option) (swarm.Broker, error) {
// 	config := swarm.NewConfig()
// 	for _, opt := range opts {
// 		opt(&config)
// 	}
//
// 	starter := lambda.Start
//
// 	type Mock interface{ Start(interface{}) }
// 	if config.Service != nil {
// 		service, ok := config.Service.(Mock)
// 		if ok {
// 			starter = service.Start
// 		}
// 	}
//
// 	sls := bridge{f: starter, c: config}
//
// 	return kernel.New(nil, sls, config), nil
// }

//------------------------------------------------------------------------------

type bridge struct{ *kernel.Bridge }

// Note: events.S3Event decodes all records, the swarm kernel protocol requires bytes.
type S3Event struct {
	Records []json.RawMessage `json:"Records"`
}

func (s bridge) Run() { lambda.Start(s.run) }

func (s bridge) run(events S3Event) error {
	bag := make([]swarm.Bag, len(events.Records))
	for i, obj := range events.Records {
		bag[i] = swarm.Bag{
			Category: Category,
			Digest:   guid.G(guid.Clock).String(),
			Object:   obj,
		}
	}

	return s.Bridge.Dispatch(bag)
}
