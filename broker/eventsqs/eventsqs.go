//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventsqs

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel"
)

type Client struct {
	config swarm.Config
}

//------------------------------------------------------------------------------

type bridge struct{ *kernel.Bridge }

func (s bridge) Run() { lambda.Start(s.run) }

func (s bridge) run(events events.SQSEvent) error {
	bag := make([]swarm.Bag, len(events.Records))
	for i, evt := range events.Records {
		bag[i] = swarm.Bag{
			Category: attr(&evt, "Category"),
			Digest:   evt.ReceiptHandle,
			Object:   []byte(evt.Body),
		}
	}

	return s.Bridge.Dispatch(bag)
}

func attr(msg *events.SQSMessage, key string) string {
	val, exists := msg.MessageAttributes[key]
	if !exists {
		return ""
	}

	return *val.StringValue
}
