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

type Client struct {
	config swarm.Config
}

//------------------------------------------------------------------------------

type bridge struct{ *kernel.Bridge }

// Note: events.DynamoDBEvent decodes all records, the swarm kernel protocol requires bytes.
type DynamoDBEvent struct {
	Records []json.RawMessage `json:"Records"`
}

func (s bridge) Run(context.Context) { lambda.Start(s.run) }

func (s bridge) run(ctx context.Context, events DynamoDBEvent) error {
	bag := make([]swarm.Bag, len(events.Records))
	for i, obj := range events.Records {
		bag[i] = swarm.Bag{
			Category: Category,
			Digest:   swarm.Digest(guid.G(guid.Clock).String()),
			Object:   obj,
		}
	}

	return s.Bridge.Dispatch(ctx, bag)
}
