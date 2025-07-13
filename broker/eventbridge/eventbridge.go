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

func (cli *Client) Close() error {
	return nil
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
	bag := make([]swarm.Bag, 1)
	bag[0] = swarm.Bag{
		Category: evt.DetailType,
		Digest:   evt.ID,
		Object:   evt.Detail,
	}

	return s.Bridge.Dispatch(ctx, bag)
}
