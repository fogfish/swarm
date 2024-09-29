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
	"github.com/fogfish/swarm/kernel"
)

type Client struct {
	config swarm.Config
}

// New creates broker for AWS EventBridge
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

//------------------------------------------------------------------------------

type bridge struct{ *kernel.Bridge }

// Note: events.DynamoDBEvent decodes all records, the swarm kernel protocol requires bytes.
type DynamoDBEvent struct {
	Records []json.RawMessage `json:"Records"`
}

func (s bridge) Run() { lambda.Start(s.run) }

func (s bridge) run(events DynamoDBEvent) error {
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
